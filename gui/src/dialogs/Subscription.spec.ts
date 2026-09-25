// @vitest-environment happy-dom
import { afterEach, expect, test, vi } from "vitest";
import { flushPromises, type VueWrapper } from "@vue/test-utils";
import { VSelect, VSwitch } from "vuetify/components";
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
  autoSelect: false,
  updateMode: "disabled" as const,
  updateIntervalMinutes: 0,
  failureIntervalMinutes: 1,
};

test("offers four modes and only shows fields required by the mode", async () => {
  wrapper = mountWithApp(Subscription, { props: { subscription } });
  const mode = wrapper.findComponent(VSelect);
  expect(mode.props("items")).toHaveLength(4);
  expect(wrapper.findAll('input[type="number"]')).toHaveLength(0);
  expect(wrapper.findAllComponents(VSwitch)).toHaveLength(1);

  mode.vm.$emit("update:modelValue", "at_interval");
  await flushPromises();
  expect(wrapper.findAll('input[type="number"]')).toHaveLength(1);

  mode.vm.$emit("update:modelValue", "interval_failsafe");
  await flushPromises();
  const inputs = wrapper.findAll('input[type="number"]');
  expect(inputs).toHaveLength(2);
  await inputs[0].setValue("15");
  await inputs[1].setValue("2");
  await wrapper
    .findAll("button")
    .find((button) => button.text() === "Save and Apply")!
    .trigger("click");
  await flushPromises();
  expect(patchSubscription).toHaveBeenCalledWith({
    subscription: {
      ...subscription,
      updateMode: "interval_failsafe",
      updateIntervalMinutes: 15,
      failureIntervalMinutes: 2,
    },
  });
});

test.each(["", "0", "0.5"])(
  "rejects invalid regular interval %s",
  async (value) => {
    wrapper = mountWithApp(Subscription, {
      props: {
        subscription: {
          ...subscription,
          updateMode: "at_interval" as const,
        },
      },
    });
    await wrapper.find('input[type="number"]').setValue(value);
    await wrapper
      .findAll("button")
      .find((button) => button.text() === "Save and Apply")!
      .trigger("click");
    await flushPromises();
    expect(patchSubscription).not.toHaveBeenCalled();
  },
);
