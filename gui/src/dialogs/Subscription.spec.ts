// @vitest-environment happy-dom
import { afterEach, expect, test, vi } from "vitest";
import { flushPromises, type VueWrapper } from "@vue/test-utils";
import { VSwitch } from "vuetify/components";
import type * as Api from "@/api";
import { patchSubscription } from "@/api";
import { mountWithApp } from "@/test/mount";
import Subscription from "./Subscription.vue";

vi.mock("@/api", async (original) => ({
  ...(await original<typeof Api>()),
  patchSubscription: vi.fn().mockResolvedValue(undefined),
}));
let wrapper: VueWrapper;
afterEach(() => {
  wrapper?.unmount();
  vi.clearAllMocks();
});

const subscription = {
  id: 1,
  _type: "subscription" as const,
  host: "example.test",
  status: "",
  info: "",
  address: "https://example.test/sub",
  remarks: "keep",
  servers: [],
};
test("only exposes updating and saves both minute timers", async () => {
  wrapper = mountWithApp(Subscription, { props: { subscription } });
  const switches = wrapper.findAllComponents(VSwitch);
  expect(switches).toHaveLength(1);
  expect(switches[0].props("modelValue")).toBe(false);
  switches[0].vm.$emit("update:modelValue", true);
  await flushPromises();
  const inputs = wrapper.findAll('input[type="number"]');
  expect(inputs).toHaveLength(2);
  await inputs[0].setValue("0");
  await inputs[1].setValue("2");
  await wrapper
    .findAll("button")
    .find((b) => b.text() === "Save and Apply")!
    .trigger("click");
  await flushPromises();
  expect(patchSubscription).toHaveBeenCalledWith({
    subscription: {
      ...subscription,
      autoUpdate: true,
      updateIntervalMinutes: 0,
      failureIntervalMinutes: 2,
    },
  });
});
test.each(["", "-1", "0.5"])(
  "rejects invalid regular interval %s",
  async (value) => {
    wrapper = mountWithApp(Subscription, {
      props: { subscription: { ...subscription, autoUpdate: true } },
    });
    await wrapper.findAll('input[type="number"]')[0].setValue(value);
    await wrapper
      .findAll("button")
      .find((b) => b.text() === "Save and Apply")!
      .trigger("click");
    await flushPromises();
    expect(patchSubscription).not.toHaveBeenCalled();
  },
);
test("rejects zero failure interval", async () => {
  wrapper = mountWithApp(Subscription, {
    props: { subscription: { ...subscription, autoUpdate: true } },
  });
  await wrapper.findAll('input[type="number"]')[1].setValue("0");
  await wrapper
    .findAll("button")
    .find((b) => b.text() === "Save and Apply")!
    .trigger("click");
  await flushPromises();
  expect(patchSubscription).not.toHaveBeenCalled();
});
