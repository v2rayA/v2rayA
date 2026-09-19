// @vitest-environment happy-dom
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { defineComponent, h, nextTick } from "vue";
import type { VueWrapper } from "@vue/test-utils";
import type * as Api from "@/api";
import type { OutboundStatus, TouchResponse, Which } from "@/api/types";
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";
import timezone from "dayjs/plugin/timezone";
import { mountWithApp } from "@/test/mount";
import { sameWhich } from "../nodes/model";
import { groupMembers, useProxies } from "./model";
import { fixture } from "./fixture";

const api = vi.hoisted(() => ({
  getTouch: vi.fn(),
  getPingLatency: vi.fn(),
  getHttpLatency: vi.fn(),
  putOutboundConnections: vi.fn(),
  putOutboundSelection: vi.fn(),
}));
vi.mock("@/api", async (original) => ({
  ...(await original<typeof Api>()),
  ...api,
}));
dayjs.extend(utc);
dayjs.extend(timezone);
let response: TouchResponse;
let wrapper: VueWrapper;
let getModel = () => useProxies();
beforeEach(async () => {
  localStorage.clear();
  vi.clearAllMocks();
  response = fixture();
  api.getTouch.mockImplementation(async () => structuredClone(response));
  api.putOutboundConnections.mockImplementation(
    async ({ outbound, touches }) => {
      response.touch.connectedServer = [
        ...response.touch.connectedServer!.filter(
          (w) => (w.outbound ?? "proxy") !== outbound,
        ),
        ...touches.map((w: Which) => ({ ...w, outbound })),
      ];
      return structuredClone(response);
    },
  );
  api.putOutboundSelection.mockImplementation(async ({ outbound, which }) => {
    response.touch.connectedServer = response.touch.connectedServer!.map((w) =>
      (w.outbound ?? "proxy") === outbound
        ? { ...w, selected: !!which && sameWhich(w, which) }
        : w,
    );
    return structuredClone(response);
  });
  api.getPingLatency.mockResolvedValue({
    whiches: [
      { id: 1, _type: "subscriptionServer", sub: 0, pingLatency: "20ms" },
    ],
  });
  api.getHttpLatency.mockResolvedValue({
    whiches: [
      { id: 1, _type: "subscriptionServer", sub: 0, pingLatency: "40ms" },
    ],
  });
  wrapper = mountWithApp(
    defineComponent({
      setup() {
        const model = useProxies();
        getModel = () => model;
        return () => h("div");
      },
    }),
  );
  getModel().store.outboundName = "media";
  await getModel().sync();
});
afterEach(() => wrapper.unmount());

describe("proxy node management", () => {
  test("resolves members by outbound and subscription index, ignoring missing nodes", () => {
    const touch = response.touch;
    touch.connectedServer!.push({ id: 99, _type: "server", outbound: "media" });
    expect(
      groupMembers(touch, touch.connectedServer!, "media").map((r) => r.name),
    ).toEqual(["North", "West"]);
    expect(
      groupMembers(touch, touch.connectedServer!, "proxy").map((r) => r.name),
    ).toEqual(["East"]);
    expect(groupMembers(touch, touch.connectedServer!, "empty")).toEqual([]);
  });
  test("combines source, name/address/protocol search and current group membership", () => {
    const model = getModel();
    for (const [query, names] of [
      [" NoRtH ", ["North"]],
      ["west.example", ["West"]],
      ["VLESS", ["West", "East"]],
      ["missing", []],
    ] as const) {
      model.query.value = query;
      expect(model.listed.value.map((r) => r.name)).toEqual(names);
      expect(model.members.value.map((r) => r.name)).toEqual(["North", "West"]);
    }
    model.query.value = "vless";
    model.source.value = response.touch.subscriptions[1].address;
    expect(model.listed.value.map((r) => r.name)).toEqual(["East"]);
    model.membersOnly.value = true;
    expect(model.listed.value).toEqual([]);
    model.source.value = "local";
    model.query.value = "";
    expect(model.listed.value.map((r) => r.name)).toEqual(["North"]);
  });
  test("toggles membership without replacing other members or another group", async () => {
    const model = getModel();
    await model.toggleGroup(model.rows.value[1]);
    expect(model.members.value.map((r) => r.name)).toEqual([
      "North",
      "West",
      "South",
    ]);
    expect(
      groupMembers(
        model.nodes.touch.value,
        model.store.connectedServer,
        "other",
      ).map((r) => r.name),
    ).toEqual(["South"]);
    await model.toggleGroup(model.rows.value[0]);
    expect(model.members.value.map((r) => r.name)).toEqual(["West", "South"]);
  });
  test("selects and clears a member through the selection endpoint, retaining all memberships", async () => {
    const model = getModel();
    const before = structuredClone(response.touch.connectedServer);
    await model.selectMember(model.rows.value[2]);
    expect(api.putOutboundSelection).toHaveBeenLastCalledWith({
      outbound: "media",
      which: { _type: "subscriptionServer", id: 1, sub: 0 },
    });
    expect(model.isSelected(model.rows.value[2])).toBe(true);
    expect(model.selectedMember.value?.selected).toBe(true);
    expect(model.mode.value).toBe("manual");
    expect(model.inUse("media")?.name).toBe("West");
    await model.setMode("auto");
    expect(api.putOutboundSelection).toHaveBeenLastCalledWith({
      outbound: "media",
      which: null,
    });
    expect(model.selectedMember.value).toBeUndefined();
    expect(model.mode.value).toBe("auto");
    expect(
      model.store.connectedServer.map(({ selected, ...w }) => {
        expect(selected).not.toBe(true);
        return w;
      }),
    ).toEqual(before);
    expect(api.putOutboundConnections).not.toHaveBeenCalled();
    await model.selectMember(model.rows.value[1]);
    expect(api.putOutboundSelection).toHaveBeenCalledTimes(2);
  });
  test("manual mode waits for a member choice; observatory ignores dead and nonmember probes", async () => {
    const model = getModel();
    const status = (
      which: Which,
      alive: boolean,
      delay: number,
    ): OutboundStatus => ({
      which,
      alive,
      delay,
      outbound_tag: "media",
      last_seen_time: 0,
      last_try_time: 0,
    });
    model.store.observatory.media = [
      status({ id: 2, _type: "server" }, true, 1),
      status({ id: 1, _type: "server" }, false, 2),
      status({ id: 1, _type: "subscriptionServer", sub: 0 }, true, 50),
    ];
    expect(model.inUse("media")?.name).toBe("West");
    model.store.observatory.media[1] = status(
      { id: 1, _type: "server" },
      true,
      20,
    );
    expect(model.inUse("media")?.name).toBe("North");
    await model.setMode("manual");
    expect(model.mode.value).toBe("manual");
    expect(api.putOutboundSelection).not.toHaveBeenCalled();
    model.store.outboundName = "other";
    await nextTick();
    expect(model.mode.value).toBe("auto");
    expect(model.inUse("other")).toBeNull();
  });
  test("batch add and remove each issue one full member update and preserve unselected members", async () => {
    const model = getModel();
    model.source.value = "local";
    model.selectAll(true);
    await model.batchMembership(true);
    expect(api.putOutboundConnections).toHaveBeenCalledTimes(1);
    expect(model.members.value.map((r) => r.name)).toEqual([
      "North",
      "West",
      "South",
    ]);
    await model.batchMembership(false);
    expect(api.putOutboundConnections).toHaveBeenCalledTimes(2);
    expect(model.members.value.map((r) => r.name)).toEqual(["West"]);
    expect(
      groupMembers(
        model.nodes.touch.value,
        model.store.connectedServer,
        "other",
      ).map((r) => r.name),
    ).toEqual(["South"]);
  });
  test("tests only listed rows over TCP and HTTP, exposing pending and returned latency", async () => {
    const model = getModel();
    model.source.value = response.touch.subscriptions[0].address;
    model.query.value = "West";
    const pending = model.testListed();
    expect(model.testing.value).toBe(true);
    expect(Number.parseFloat(model.listed.value[0].pingLatency)).toBeNaN();
    await pending;
    expect(api.getPingLatency).toHaveBeenCalledWith([
      { id: 1, _type: "subscriptionServer", sub: 0 },
    ]);
    expect(model.listed.value[0].pingLatency).toBe("20ms");
    expect(model.testing.value).toBe(false);
    await model.testListed(true);
    expect(api.getHttpLatency).toHaveBeenCalledWith([
      { id: 1, _type: "subscriptionServer", sub: 0 },
    ]);
    expect(model.listed.value[0].pingLatency).toBe("40ms");
  });
  test("drops hidden batch selection so a later delete cannot affect filtered-out rows", async () => {
    const model = getModel();
    model.selectAll(true);
    model.source.value = "local";
    await nextTick();
    expect(model.selected.value.map((r) => r.name)).toEqual(["North", "South"]);
    expect(model.canDelete.value).toBe(true);
    model.query.value = "South";
    await nextTick();
    model.query.value = "";
    expect(model.selected.value.map((r) => r.name)).toEqual(["South"]);
    expect(model.allSelected.value).toBe(false);
  });
});
