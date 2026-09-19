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
  });
  vi.mocked(putSetting).mockResolvedValue(undefined);
  wrapper = mountWithApp(SubscriptionSettings);
  await flushPromises();
});
afterEach(() => wrapper.unmount());
test("loads subscription values and saves all three changes without losing other settings", async () => {
  const selects = wrapper.findAllComponents(VSelect);
  expect(selects[0].props("modelValue")).toBe("auto_update");
  expect(selects[1].props("modelValue")).toBe("pac");
  selects[0].vm.$emit("update:modelValue", "auto_update_at_intervals");
  await flushPromises();
  expect(
    (wrapper.get('input[type="number"]').element as HTMLInputElement).value,
  ).toBe("6");
  await wrapper.get('input[type="number"]').setValue("12");
  selects[1].vm.$emit("update:modelValue", "proxy");
  await wrapper
    .findAll("button")
    .find((b) => b.text() === "Save")!
    .trigger("click");
  await flushPromises();
  expect(putSetting).toHaveBeenCalledWith(
    {
      ...loaded,
      subscriptionAutoUpdateMode: "auto_update_at_intervals",
      subscriptionAutoUpdateIntervalHour: 12,
      proxyModeWhenSubscribe: "proxy",
    },
    expect.objectContaining({ signal: expect.any(AbortSignal) }),
  );
  expect(wrapper.emitted("close")).toEqual([[true]]);
});
test("does not save an empty or nonpositive update interval", async () => {
  wrapper
    .findAllComponents(VSelect)[0]
    .vm.$emit("update:modelValue", "auto_update_at_intervals");
  await flushPromises();
  await wrapper.get('input[type="number"]').setValue("0");
  await wrapper
    .findAll("button")
    .find((b) => b.text() === "Save")!
    .trigger("click");
  await flushPromises();
  expect(putSetting).not.toHaveBeenCalled();
  expect(wrapper.emitted("close")).toBeUndefined();
});
