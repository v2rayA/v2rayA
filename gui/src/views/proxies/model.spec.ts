// @vitest-environment happy-dom
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { defineComponent, h, nextTick } from "vue";
import type { VueWrapper } from "@vue/test-utils";
import type * as Api from "@/api";
import type * as Download from "@/lib/download";
import type { OutboundStatus, TouchResponse, Which } from "@/api/types";
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";
import timezone from "dayjs/plugin/timezone";
import { mountWithApp } from "@/test/mount";
import { groupMembers, useProxies } from "./model";
import { fixture } from "./fixture";

const api = vi.hoisted(() => ({
  getTouch: vi.fn(),
  getPingLatency: vi.fn(),
  getHttpLatency: vi.fn(),
  putOutboundConnections: vi.fn(),
  getSharingAddress: vi.fn(),
}));
const lib = vi.hoisted(() => ({ copyText: vi.fn(), saveText: vi.fn() }));
vi.mock("@/api", async (original) => ({
  ...(await original<typeof Api>()),
  ...api,
}));
vi.mock("@/lib/clipboard", () => ({ copyText: lib.copyText }));
vi.mock("@/lib/download", async (original) => ({
  ...(await original<typeof Download>()),
  saveText: lib.saveText,
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
  test("combines source, name/address/protocol search, ignoring group membership", () => {
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
    model.source.value = "local";
    model.query.value = "";
    expect(model.listed.value.map((r) => r.name)).toEqual(["North", "South"]);
  });
  test("setMembership changes one row and leaves the other members and groups alone", async () => {
    const model = getModel();
    await model.setMembership(model.rows.value[1], "media", true);
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
    await model.setMembership(model.rows.value[0], "media", false);
    expect(model.members.value.map((r) => r.name)).toEqual(["West", "South"]);
  });
  test("adds a node to the group picked from the menu and removes it from one it is in", async () => {
    const model = getModel();
    model.store.outbounds = ["proxy", "media", "other"];
    const north = model.rows.value[0];
    const east = model.rows.value[3];
    expect(model.memberGroups(north)).toEqual(["media"]);
    expect(model.memberGroups(east)).toEqual(["proxy"]);
    await model.nodeAction(north, "addToGroup", "other");
    expect(model.memberGroups(north)).toEqual(["media", "other"]);
    expect(model.members.value.map((r) => r.name)).toEqual(["North", "West"]);
    await model.nodeAction(east, "addToGroup", "other");
    expect(model.memberGroups(east)).toEqual(["proxy", "other"]);
    expect(model.memberGroups(north)).toEqual(["media", "other"]);
    await model.nodeAction(north, "removeFromGroup", "media");
    expect(model.memberGroups(north)).toEqual(["other"]);
    expect(model.members.value.map((r) => r.name)).toEqual(["West"]);
    // a group the node is already in, or one it is not in, changes nothing
    await model.nodeAction(north, "addToGroup", "other");
    await model.nodeAction(north, "removeFromGroup", "proxy");
    expect(api.putOutboundConnections).toHaveBeenCalledTimes(3);
  });
  test("observatory ignores dead and nonmember probes", async () => {
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
    expect(model.inUse("other")).toBeNull();
  });
  test("leaving the group in view clears the batch selection", async () => {
    const model = getModel();
    model.selectAll(true);
    expect(model.selected.value).toHaveLength(4);
    model.store.outboundName = "other";
    await nextTick();
    expect(model.selected.value).toEqual([]);
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
  test("exports the selected nodes to the clipboard or to a timestamped TXT file", async () => {
    const model = getModel();
    api.getSharingAddress.mockResolvedValue({
      sharingAddress: "vmess://north",
    });
    model.selectRow(model.rows.value[0], true);
    await model.exportSelected("file");
    expect(api.getSharingAddress).toHaveBeenCalledWith({
      _type: "server",
      id: 1,
    });
    expect(lib.saveText).toHaveBeenCalledWith(
      expect.stringMatching(/^export-\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}\.txt$/),
      "vmess://north",
    );
    expect(lib.copyText).not.toHaveBeenCalled();
    await model.exportSelected();
    expect(lib.copyText).toHaveBeenCalledWith("vmess://north");
    // nothing shareable: neither destination is written
    api.getSharingAddress.mockResolvedValue({ sharingAddress: "" });
    await model.exportSelected("file");
    expect(lib.saveText).toHaveBeenCalledTimes(1);
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
