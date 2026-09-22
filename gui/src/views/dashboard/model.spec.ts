// @vitest-environment happy-dom
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { DOMWrapper, flushPromises } from "@vue/test-utils";
import type { VueWrapper } from "@vue/test-utils";
import type * as api from "@/api";
import {
  deleteV2ray,
  getSetting,
  getTouch,
  postV2ray,
  getPorts,
  putSetting,
  putOutboundConnections,
  putOutboundSelection,
  getPingLatency,
  putSubscription,
} from "@/api";
import { watchConnected } from "@/api/connect";
import type { TouchResponse, TouchServer } from "@/api/types";
import { loadingState } from "@/composables/useLoading";
import { closeAllNotices, noticeState } from "@/composables/useNotify";
import { closeAllDialogs } from "@/composables/useDialog";
import DialogHost from "@/components/hosts/DialogHost.vue";
import GroupMembersDialog from "./GroupMembers.vue";
import { useAppStore } from "@/stores/app";
import { mountWithApp } from "@/test/mount";
import DashboardView from "../DashboardView.vue";

vi.mock("@/api", async (original) => ({
  ...(await original<typeof api>()),
  getTouch: vi.fn(),
  getSetting: vi.fn(),
  postV2ray: vi.fn(),
  deleteV2ray: vi.fn(),
  getPorts: vi.fn(),
  putSetting: vi.fn(),
  putOutboundSelection: vi.fn(),
  putOutboundConnections: vi.fn(),
  getPingLatency: vi.fn(),
  putSubscription: vi.fn(),
}));
vi.mock("@/api/connect", () => ({
  watchConnected: vi.fn((request: Promise<TouchResponse>) => request),
}));

const server = (name: string): TouchServer => ({
  id: 1,
  _type: "server",
  name,
  address: `${name}.example:443`,
  net: "vmess",
  pingLatency: "24ms",
});
function response(running = false): TouchResponse {
  return {
    running,
    networkPaused: false,
    touch: {
      servers: [server("Standalone")],
      subscriptions: [
        {
          id: 1,
          _type: "subscription",
          host: "subscription.example",
          address: "https://subscription.example",
          status: "",
          info: "",
          autoSelect: false,
          servers: [{ ...server("Subscribed"), _type: "subscriptionServer" }],
        },
      ],
      connectedServer: [
        { _type: "server", id: 1 },
        { _type: "subscriptionServer", id: 1, sub: 0, outbound: "work" },
      ],
    },
  };
}

let wrapper: VueWrapper;
const button = (text: string) =>
  wrapper.findAll("button").find((item) => item.text() === text)!;
/** menus render in the body, so their entries are found there */
const listItem = (text: string) =>
  [...document.body.querySelectorAll<HTMLElement>(".v-list-item")].find(
    (item) => item.textContent?.trim() === text,
  )!;
const control = () =>
  wrapper
    .getComponent(".dashboard-status")
    .findAll("button")
    .find((b) => /Start|Stop/.test(b.text()))!;

beforeEach(() => {
  vi.clearAllMocks();
  closeAllNotices();
  closeAllDialogs();
  vi.mocked(getPorts).mockResolvedValue({
    socks5: 20170,
    http: 20171,
    httpWithPac: 20172,
    socks5WithPac: 0,
    vmess: 0,
    api: { port: 0, services: [] },
  });
  vi.mocked(putSetting).mockResolvedValue(undefined);
  vi.mocked(putSubscription).mockResolvedValue(response());
  vi.mocked(putOutboundSelection).mockResolvedValue(response());
  vi.mocked(putOutboundConnections).mockResolvedValue(response());
  // the tile pings the members once on arrival; a test that wants a
  // result arranges it after the mount
  vi.mocked(getPingLatency).mockResolvedValue({ whiches: [] });
  vi.mocked(getTouch).mockResolvedValue(response());
  vi.mocked(getSetting).mockResolvedValue({
    setting: {
      transparent: "close",
      transparentType: "tproxy",
      pacMode: "routingA",
      logLevel: "info",
    },
    localGFWListVersion: "",
    localGeositeVersion: "",
  });
  vi.mocked(postV2ray).mockResolvedValue(response(true));
  vi.mocked(deleteV2ray).mockResolvedValue(response());
});
afterEach(() => {
  wrapper?.unmount();
  closeAllDialogs();
});

describe("dashboard", () => {
  test("shows the current group node, then starts and stops the core", async () => {
    wrapper = mountWithApp(DashboardView);
    await flushPromises();
    expect(wrapper.get('[role="status"]').text()).toBe("Ready");
    expect(wrapper.get(".dashboard-connection").text()).toContain("Standalone");
    await control().trigger("click");
    await flushPromises();
    expect(postV2ray).toHaveBeenCalledOnce();
    expect(wrapper.get('[role="status"]').text()).toBe("Running");
    await control().trigger("click");
    await flushPromises();
    expect(deleteV2ray).toHaveBeenCalledOnce();
    expect(wrapper.get('[role="status"]').text()).toBe("Ready");
    expect(getTouch).toHaveBeenCalledOnce();

    useAppStore().outboundName = "work";
    await flushPromises();
    expect(wrapper.get(".dashboard-connection").text()).toContain("Subscribed");
  });

  test("keeps the confirmed state while starting and prevents duplicate requests", async () => {
    // The project's ES library target predates Promise.withResolvers.
    let finish!: (value: TouchResponse) => void;
    const promise = new Promise<TouchResponse>((resolve) => {
      finish = resolve;
    });
    vi.mocked(postV2ray).mockReturnValue(promise);
    wrapper = mountWithApp(DashboardView);
    await flushPromises();
    await control().trigger("click");
    await control().trigger("click");
    expect(postV2ray).toHaveBeenCalledOnce();
    expect(wrapper.get('[role="status"]').text()).toBe("Ready");
    expect(control().attributes("disabled")).toBeDefined();
    expect(loadingState.open.size).toBe(1);
    finish(response(true));
    await flushPromises();
    expect(wrapper.get('[role="status"]').text()).toBe("Running");
    expect(loadingState.open.size).toBe(0);
  });

  test("refreshes confirmed state when the connection watcher wins", async () => {
    vi.mocked(watchConnected).mockResolvedValueOnce(undefined);
    vi.mocked(getTouch)
      .mockResolvedValueOnce(response())
      .mockResolvedValueOnce(response(true));
    wrapper = mountWithApp(DashboardView);
    await flushPromises();
    await control().trigger("click");
    await flushPromises();
    expect(wrapper.get('[role="status"]').text()).toBe("Running");
    expect(getTouch).toHaveBeenCalledTimes(2);
  });

  test("resumes a paused core and preserves state on a failed stop", async () => {
    vi.mocked(getTouch).mockResolvedValue({
      ...response(true),
      networkPaused: true,
    });
    wrapper = mountWithApp(DashboardView);
    await flushPromises();
    expect(wrapper.get('[role="status"]').text()).toBe("Waiting for network");
    await control().trigger("click");
    await flushPromises();
    expect(postV2ray).toHaveBeenCalledOnce();
    vi.mocked(deleteV2ray).mockRejectedValueOnce(new Error("Core is busy"));
    await control().trigger("click");
    await flushPromises();
    expect(wrapper.get('[role="status"]').text()).toBe("Running");
    expect(noticeState.current?.text).toContain("Core is busy");
    expect(control().attributes("disabled")).toBeUndefined();
    expect(loadingState.open.size).toBe(0);
  });

  test("keeps start disabled and reports a failed load", async () => {
    vi.mocked(getTouch).mockRejectedValueOnce(new Error("Backend unreachable"));
    wrapper = mountWithApp(DashboardView);
    expect(control().attributes("disabled")).toBeDefined();
    await flushPromises();
    expect(wrapper.get('[role="alert"]').text()).toContain(
      "Backend unreachable",
    );
    expect(noticeState.current).toBeNull();
    expect(control().attributes("disabled")).toBeDefined();
  });

  test("offers the group editor for an empty group and saves its members", async () => {
    const dialogs = mountWithApp(DialogHost);
    vi.mocked(getTouch).mockResolvedValueOnce({
      ...response(),
      touch: { ...response().touch, connectedServer: [] },
    });
    wrapper = mountWithApp(DashboardView);
    await flushPromises();
    expect(wrapper.get(".dashboard-connection").text()).toContain(
      "This group has no nodes",
    );
    await button("Add or remove nodes").trigger("click");
    await flushPromises();
    const dialog = dialogs.getComponent(GroupMembersDialog);
    await dialog.findAll(".v-list-item")[0].trigger("click");
    await dialog
      .findAll("button")
      .find((b) => b.text() === "Save")!
      .trigger("click");
    await flushPromises();
    expect(putOutboundConnections).toHaveBeenCalledWith({
      outbound: "proxy",
      touches: [{ _type: "server", id: 1, sub: 0, outbound: "proxy" }],
    });
    dialogs.unmount();
  });

  test("names the proxy group in view on the card and switches it from there", async () => {
    wrapper = mountWithApp(DashboardView);
    useAppStore().setOutbounds(["proxy", "work"]);
    await flushPromises();
    const card = wrapper.get(".dashboard-connection");
    expect(card.text()).toContain("Proxy group");
    expect(card.text()).toContain("PROXY");
    await card
      .findAll("button")
      .find((b) => b.text() === "PROXY")!
      .trigger("click");
    await flushPromises();
    await listItem("WORK").dispatchEvent(
      new MouseEvent("click", { bubbles: true }),
    );
    await flushPromises();
    expect(useAppStore().outboundName).toBe("work");
    expect(wrapper.get(".dashboard-connection").text()).toContain(
      "Subscribed",
    );
  });
  test("prefers a pinned member, then the best alive probe, then a single member", async () => {
    const data = response();
    data.touch.servers.push(
      { ...server("Fast"), id: 2 },
      { ...server("Dead"), id: 3 },
    );
    data.touch.connectedServer = [
      { _type: "server", id: 1, selected: true },
      { _type: "server", id: 2 },
      { _type: "server", id: 3 },
    ];
    vi.mocked(getTouch).mockResolvedValue(data);
    wrapper = mountWithApp(DashboardView);
    await flushPromises();
    const store = useAppStore();
    store.observatory.proxy = [1, 2, 3].map((id) => ({
      which: { _type: "server", id },
      alive: id !== 3,
      delay: id === 1 ? 90 : id === 2 ? 20 : 1,
      outbound_tag: "",
      last_seen_time: 0,
      last_try_time: 0,
    }));
    await flushPromises();
    const connection = () => wrapper.get(".dashboard-connection").text();
    expect(connection()).toContain("Standalone");
    expect(connection()).toContain("90 ms");
    store.connectedServer[0].selected = false;
    await flushPromises();
    expect(connection()).toContain("Fast");
    expect(connection()).toContain("20 ms");
    expect(
      wrapper
        .get(".dashboard-latency")
        .findAll(".v-list-item")
        .map((item) => item.text()),
    ).toEqual(["Fast20 ms", "Standalone90 ms", "Dead1 ms"]);
    store.observatory = {};
    await flushPromises();
    expect(connection()).not.toContain("Fast");
    store.connectedServer = [{ _type: "server", id: 1 }];
    await flushPromises();
    expect(connection()).toContain("Standalone");
    // the latency tile tests every member, the node in use included
    vi.mocked(getPingLatency).mockResolvedValueOnce({
      whiches: [{ _type: "server", id: 1, pingLatency: "17ms" }],
    });
    await wrapper
      .get(".dashboard-latency")
      .findAll("button")
      .find((b) => b.text() === "Test latency")!
      .trigger("click");
    await flushPromises();
    expect(getPingLatency).toHaveBeenLastCalledWith([
      { _type: "server", id: 1 },
    ]);
    expect(connection()).toContain("17ms");
  });

  test("chooses a member and returns to automatic routing through the picker", async () => {
    wrapper = mountWithApp(DashboardView);
    await flushPromises();
    const pinned = response();
    pinned.touch.connectedServer![0].selected = true;
    vi.mocked(putOutboundSelection)
      .mockResolvedValueOnce(pinned)
      .mockResolvedValueOnce(response());
    const menuItem = async (index: number) => {
      await wrapper
        .get(".dashboard-connection .dashboard-node-name")
        .trigger("click");
      await flushPromises();
      const items = [...document.querySelectorAll('[role="menuitemradio"]')];
      await new DOMWrapper(items[index]).trigger("click");
      await flushPromises();
    };
    await menuItem(1);
    expect(putOutboundSelection).toHaveBeenNthCalledWith(1, {
      outbound: "proxy",
      which: { _type: "server", id: 1 },
    });
    expect(wrapper.get(".dashboard-connection").text()).toContain("Pinned");
    await menuItem(0);
    expect(putOutboundSelection).toHaveBeenNthCalledWith(2, {
      outbound: "proxy",
      which: null,
    });
    expect(wrapper.get(".dashboard-connection").text()).not.toContain("Pinned");
  });

  test("autosaves the whole form from a quick control", async () => {
    wrapper = mountWithApp(DashboardView);
    await flushPromises();
    await wrapper
      .get('.dashboard-transparent input[type="checkbox"]')
      .setValue(true);
    await flushPromises();
    const saved = vi.mocked(putSetting).mock.calls[0][0];
    expect(saved).toMatchObject({
      transparent: "close",
      transparentType: "tproxy",
      pacMode: "routingA",
      logLevel: "info",
      portSharing: true,
      mux: 8,
      tunAutoRoute: true,
      subscriptionAutoUpdateIntervalHour: 0,
    });
    expect(putSetting).toHaveBeenCalledOnce();
  });

  test("updates all subscriptions sequentially and continues after an update fails", async () => {
    const data = response();
    data.touch.subscriptions[0].info =
      "Used 1 GiB / 10 GiB · Expires 2026-10-01";
    data.touch.subscriptions.push({
      ...data.touch.subscriptions[0],
      id: 2,
      host: "second.example",
    });
    vi.mocked(getTouch).mockResolvedValue(data);
    let reject!: (err: Error) => void;
    vi.mocked(putSubscription)
      .mockReturnValueOnce(
        new Promise((_resolve, fail) => {
          reject = fail;
        }),
      )
      .mockResolvedValueOnce(data);
    wrapper = mountWithApp(DashboardView);
    await flushPromises();
    expect(wrapper.get(".dashboard-subscriptions").text()).toContain(
      "1 GiB / 10 GiB",
    );
    // the card's loading bar is indeterminate; the usage bar carries the value
    expect(
      wrapper
        .findAll(".dashboard-subscriptions [role='progressbar']")
        .map((bar) => bar.attributes("aria-valuenow"))
        .filter((value) => value !== undefined),
    ).toContain("10");
    await button("Update all").trigger("click");
    expect(putSubscription).toHaveBeenCalledTimes(1);
    reject(new Error("Subscription unavailable"));
    await flushPromises();
    expect(
      vi.mocked(putSubscription).mock.calls.map(([which]) => which),
    ).toEqual([
      { _type: "subscription", id: 1 },
      { _type: "subscription", id: 2 },
    ]);
    expect(noticeState.current?.text).toContain("Subscription unavailable");
    expect(button("Update all").attributes("disabled")).toBeUndefined();
  });
});
