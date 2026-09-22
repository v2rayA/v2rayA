// @vitest-environment happy-dom
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { flushPromises, type VueWrapper } from "@vue/test-utils";
import { VCheckboxBtn } from "vuetify/components";
import type * as Api from "@/api";
import type * as Download from "@/lib/download";
import { getSharingAddress, getTouch, putOutboundConnections } from "@/api";
import { copyText } from "@/lib/clipboard";
import { saveText } from "@/lib/download";
import { mountWithApp } from "@/test/mount";
import { useAppStore } from "@/stores/app";
import { dialogState, closeAllDialogs } from "@/composables/useDialog";
import ImportDialog from "@/dialogs/Import.vue";
import ServerDialog from "@/dialogs/Server/index.vue";
import ProxiesView from "./ProxiesView.vue";
import NodeCard from "./proxies/NodeCard.vue";
import NodeListItem from "./proxies/NodeListItem.vue";
import { fixture } from "./proxies/fixture";
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";
import timezone from "dayjs/plugin/timezone";

vi.mock("@/api", async (original) => ({
  ...(await original<typeof Api>()),
  getTouch: vi.fn(),
  putOutboundConnections: vi.fn(),
  getSharingAddress: vi.fn(),
}));
vi.mock("@/lib/download", async (original) => ({
  ...(await original<typeof Download>()),
  saveText: vi.fn(),
}));
vi.mock("@/lib/clipboard", () => ({ copyText: vi.fn() }));
dayjs.extend(utc);
dayjs.extend(timezone);
let wrapper: VueWrapper;
const originalWidth = window.innerWidth;
/** phones the batch bar collapses into a single menu */
function asPhone() {
  window.innerWidth = 500;
  window.dispatchEvent(new Event("resize"));
}
const button = (text: string) =>
  wrapper
    .findAll("button, .v-chip")
    .find((b) => b.text() === text || b.attributes("aria-label") === text)!;
/** menus render in the body, so their entries are found there */
const listItem = (text: string) =>
  [...document.body.querySelectorAll<HTMLElement>(".v-list-item")].find(
    (item) => item.textContent?.trim() === text,
  )!;
beforeEach(async () => {
  localStorage.clear();
  vi.clearAllMocks();
  vi.mocked(getTouch).mockImplementation(async () => fixture());
  vi.mocked(putOutboundConnections).mockResolvedValue(fixture());
  wrapper = mountWithApp(ProxiesView);
  useAppStore().outboundName = "media";
  await flushPromises();
});
afterEach(() => {
  wrapper.unmount();
  window.innerWidth = originalWidth;
  window.dispatchEvent(new Event("resize"));
  closeAllDialogs();
});

describe("unified proxies page", () => {
  test("shows node cards, switches to a real list and remembers it", async () => {
    expect(wrapper.findAllComponents(NodeCard)).toHaveLength(4);
    await button("List").trigger("click");
    await flushPromises();
    expect(wrapper.get(".proxies__list").classes()).toContain("v-list");
    expect(wrapper.findAllComponents(NodeListItem)).toHaveLength(4);
    expect(wrapper.findComponent(NodeCard).exists()).toBe(false);
    expect(localStorage.getItem("proxiesView")).toBe("list");
    wrapper.unmount();
    wrapper = mountWithApp(ProxiesView);
    await flushPromises();
    expect(wrapper.findAllComponents(NodeListItem)).toHaveLength(4);
    await button("Cards").trigger("click");
    expect(wrapper.findAllComponents(NodeCard)).toHaveLength(4);
  });
  test("select-all selects listed rows only, shows partial state and never toggles membership", async () => {
    await button("List").trigger("click");
    const all = wrapper.get(".proxies__batch").getComponent(VCheckboxBtn);
    await all.get("input").setValue(true);
    expect(wrapper.get(".proxies__batch").text()).toContain("4 selected");
    expect(button("Delete").attributes("disabled")).toBeDefined();
    const first = wrapper
      .findAllComponents(NodeListItem)[0]
      .getComponent(VCheckboxBtn);
    await first.get("input").setValue(false);
    expect(all.props("indeterminate")).toBe(true);
    expect(wrapper.get(".proxies__batch").text()).toContain("3 selected");
    expect(putOutboundConnections).not.toHaveBeenCalled();
    await wrapper.get(".proxies__search input").setValue("South");
    await all.get("input").setValue(true);
    expect(wrapper.get(".proxies__batch").text()).toContain("1 selected");
    expect(button("Delete").attributes("disabled")).toBeUndefined();
    await all.get("input").setValue(false);
    expect(wrapper.get(".proxies__batch").text()).toContain("none selected");
  });
  test("toolbar buttons open their dialogs", async () => {
    await button("New node").trigger("click");
    expect(dialogState.stack.at(-1)?.component).toBe(ServerDialog);
    expect(dialogState.stack.at(-1)?.props).toEqual({ which: null });
    closeAllDialogs();
    await flushPromises();
    await button("Import").trigger("click");
    expect(dialogState.stack.at(-1)?.component).toBe(ImportDialog);
  });
  test("Ctrl+A selects every listed node in the list view and Escape clears", async () => {
    window.dispatchEvent(
      new KeyboardEvent("keydown", { key: "a", ctrlKey: true }),
    );
    await flushPromises();
    expect(wrapper.get(".proxies__batch").text()).toContain("none selected");
    await button("List").trigger("click");
    await flushPromises();
    window.dispatchEvent(
      new KeyboardEvent("keydown", { key: "a", ctrlKey: true }),
    );
    await flushPromises();
    expect(wrapper.get(".proxies__batch").text()).toContain("4 selected");
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
    await flushPromises();
    expect(wrapper.get(".proxies__batch").text()).toContain("none selected");
  });
  test("clicking a card or a list row changes no membership", async () => {
    const card = wrapper.findAllComponents(NodeCard)[0];
    await card.get(".node-card").trigger("click");
    await card.get(".node-card").trigger("keydown", { key: "Enter" });
    await flushPromises();
    expect(putOutboundConnections).not.toHaveBeenCalled();
    await button("List").trigger("click");
    await flushPromises();
    await wrapper.findAllComponents(NodeListItem)[0].trigger("click");
    await flushPromises();
    expect(putOutboundConnections).not.toHaveBeenCalled();
  });
  test("cards are checkable and the batch bar exports to the clipboard or a file", async () => {
    vi.mocked(getSharingAddress).mockResolvedValue({
      sharingAddress: "vmess://north",
    });
    const card = wrapper.findAllComponents(NodeCard)[0];
    await card.getComponent(VCheckboxBtn).get("input").setValue(true);
    expect(wrapper.get(".proxies__batch").text()).toContain("1 selected");
    await button("Export").trigger("click");
    await flushPromises();
    await listItem("Export to clipboard").dispatchEvent(
      new MouseEvent("click", { bubbles: true }),
    );
    await flushPromises();
    expect(copyText).toHaveBeenCalledWith("vmess://north");
    await button("Export").trigger("click");
    await flushPromises();
    await listItem("Export to TXT file").dispatchEvent(
      new MouseEvent("click", { bubbles: true }),
    );
    await flushPromises();
    expect(saveText).toHaveBeenCalledWith(
      expect.stringMatching(/^export-\d{4}-\d{2}-\d{2}_\d{2}-\d{2}-\d{2}\.txt$/),
      "vmess://north",
    );
  });
  test("on a phone the batch bar folds everything but the test button", async () => {
    wrapper.unmount();
    asPhone();
    wrapper = mountWithApp(ProxiesView);
    useAppStore().outboundName = "media";
    await flushPromises();
    const bar = wrapper.get(".proxies__batch");
    await wrapper
      .findAllComponents(NodeCard)[0]
      .getComponent(VCheckboxBtn)
      .get("input")
      .setValue(true);
    expect(bar.text()).toContain("Test latency");
    expect(bar.text()).toContain("More actions");
    expect(bar.text()).not.toContain("Delete");
    expect(bar.text()).not.toContain("Export");
    await bar.findAll("button").find((b) => b.text() === "More actions")!.trigger("click");
    await flushPromises();
    expect(listItem("Add to proxy group")).toBeTruthy();
    expect(listItem("Remove from proxy group")).toBeTruthy();
    expect(listItem("Delete")).toBeTruthy();
    expect(listItem("Export to clipboard")).toBeTruthy();
    expect(listItem("Export to TXT file")).toBeTruthy();
  });
  test("the node menu adds the node to the group picked from its submenu", async () => {
    const card = wrapper.findAllComponents(NodeCard)[0];
    await card.get(".node-menu").trigger("click");
    await flushPromises();
    await listItem("Add to proxy group").dispatchEvent(
      new MouseEvent("click", { bubbles: true }),
    );
    await flushPromises();
    await listItem("PROXY").dispatchEvent(
      new MouseEvent("click", { bubbles: true }),
    );
    await flushPromises();
    expect(putOutboundConnections).toHaveBeenCalledWith({
      outbound: "proxy",
      touches: [
        { id: 1, _type: "subscriptionServer", sub: 1, outbound: "proxy" },
        { id: 1, _type: "server", sub: 0, outbound: "proxy" },
      ],
    });
  });
  test("keeps a failed load distinct from empty results and retries", async () => {
    wrapper.unmount();
    vi.mocked(getTouch).mockRejectedValueOnce(new Error("backend unavailable"));
    wrapper = mountWithApp(ProxiesView);
    await flushPromises();
    expect(wrapper.get('[role="alert"]').text()).toContain(
      "backend unavailable",
    );
    expect(wrapper.findComponent(NodeCard).exists()).toBe(false);
    await wrapper.get('[role="alert"] button').trigger("click");
    await flushPromises();
    expect(wrapper.findAllComponents(NodeCard)).toHaveLength(4);
    expect(wrapper.find('[role="alert"]').exists()).toBe(false);
  });
});
