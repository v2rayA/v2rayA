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
test("old subscriptions default to working-server selection and monitoring off; both switches save", async () => {
  const subscription = {
    id: 1,
    _type: "subscription" as const,
    host: "example.test",
    status: "",
    info: "",
    address: "https://example.test/sub",
    remarks: "keep",
    autoSelect: true,
    servers: [],
  };
  wrapper = mountWithApp(Subscription, { props: { subscription } });
  const switches = wrapper.findAllComponents(VSwitch);
  expect(switches.map((s) => s.props("modelValue"))).toEqual([
    true,
    false,
    false,
  ]);
  switches[1].vm.$emit("update:modelValue", true);
  switches[2].vm.$emit("update:modelValue", true);
  await flushPromises();
  await wrapper
    .findAll("button")
    .find((b) => b.text() === "Save and Apply")!
    .trigger("click");
  await flushPromises();
  expect(patchSubscription).toHaveBeenCalledWith({
    subscription: { ...subscription, monitor: true, preferFirst: true },
  });
  expect(wrapper.emitted("close")).toEqual([[true]]);
});
test("saved policies reopen without being reset", () => {
  wrapper = mountWithApp(Subscription, {
    props: {
      subscription: {
        id: 1,
        _type: "subscription" as const,
        host: "example.test",
        status: "",
        info: "",
        address: "https://example.test/sub",
        autoSelect: false,
        monitor: true,
        preferFirst: true,
        servers: [],
      },
    },
  });
  expect(
    wrapper.findAllComponents(VSwitch).map((s) => s.props("modelValue")),
  ).toEqual([false, true, true]);
});
