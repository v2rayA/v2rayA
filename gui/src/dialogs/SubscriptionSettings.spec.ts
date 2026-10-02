// @vitest-environment happy-dom
import { afterEach, beforeEach, expect, test, vi } from "vitest";
import { flushPromises, type VueWrapper } from "@vue/test-utils";
import { VSelect } from "vuetify/components";
import type * as Api from "@/api";
import { getSetting, putSetting } from "@/api";
import { mountWithApp } from "@/test/mount";
import { defaultForm } from "@/views/settings/model";
import SubscriptionSettings from "./SubscriptionSettings.vue";

vi.mock("@/api", async (original) => ({
  ...(await original<typeof Api>()),
  getSetting: vi.fn(),
  putSetting: vi.fn(),
}));
vi.mock("@/api/connect", () => ({
  watchConnected: vi.fn((request: Promise<unknown>) => request),
}));
const loaded = {
  ...defaultForm(),
  transparent: "proxy",
  pacMode: "routingA",
  mux: 12,
  subscriptionAutoUpdateMode: "auto_update",
  subscriptionAutoUpdateIntervalHour: 6,
  proxyModeWhenSubscribe: "pac",
};
let wrapper: VueWrapper;
beforeEach(async () => {
  vi.clearAllMocks();
  vi.mocked(getSetting).mockResolvedValue({
    setting: { ...loaded },
    localGFWListVersion: "",
    localGeositeVersion: "",
  });
  vi.mocked(putSetting).mockResolvedValue(undefined);
  wrapper = mountWithApp(SubscriptionSettings);
  await flushPromises();
});
afterEach(() => wrapper.unmount());
test("only configures download transport; schedules belong to individual subscriptions", async () => {
  const selects = wrapper.findAllComponents(VSelect);
  expect(selects).toHaveLength(1);
  expect(selects[0].props("modelValue")).toBe("pac");
  expect(wrapper.findAll('input[type="number"]')).toHaveLength(0);
  selects[0].vm.$emit("update:modelValue", "proxy");
  await wrapper
    .findAll("button")
    .find((b) => b.text() === "Save")!
    .trigger("click");
  await flushPromises();
  expect(putSetting).toHaveBeenCalledWith(
    { ...loaded, proxyModeWhenSubscribe: "proxy" },
    expect.objectContaining({ signal: expect.any(AbortSignal) }),
  );
});
