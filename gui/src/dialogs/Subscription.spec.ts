// @vitest-environment happy-dom
import { afterEach, expect, test, vi } from "vitest";
import { flushPromises, type VueWrapper } from "@vue/test-utils";
import { VSelect, VSwitch } from "vuetify/components";
import type * as Api from "@/api";
import type { TouchSubscription } from "@/api/types";
import { patchSubscription } from "@/api";
import { mountWithApp } from "@/test/mount";
import Subscription from "./Subscription.vue";

vi.mock("@/api", async (original) => ({
  ...(await original<typeof Api>()),
  patchSubscription: vi.fn().mockResolvedValue(undefined),
  getSetting: vi.fn().mockResolvedValue({ setting: { sendHWID: false } }),
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
  allowDirectRecovery: false,
  updateMode: "disabled" as const,
  updateIntervalMinutes: 0,
  failureIntervalMinutes: 1,
};

test("offers four modes and only shows fields required by the mode", async () => {
  wrapper = mountWithApp(Subscription, { props: { subscription } });
  const mode = wrapper.findComponent(VSelect);
  expect(mode.props("items")).toHaveLength(4);
  expect(wrapper.findAll('input[type="number"]')).toHaveLength(0);
  expect(
    wrapper.findAllComponents(VSwitch).map((s) => s.props("label")),
  ).toContain("Send HWID to this subscription");
  expect(
    wrapper
      .findAllComponents(VSwitch)
      .some(
        (control) =>
          control.props("label") ===
          "Ignore subscription download routing during recovery",
      ),
  ).toBe(false);

  mode.vm.$emit("update:modelValue", "at_interval");
  await flushPromises();
  expect(wrapper.findAll('input[type="number"]')).toHaveLength(1);
  expect(
    wrapper
      .findAllComponents(VSwitch)
      .some(
        (control) =>
          control.props("label") ===
          "Ignore subscription download routing during recovery",
      ),
  ).toBe(false);

  mode.vm.$emit("update:modelValue", "interval_failsafe");
  await flushPromises();
  const inputs = wrapper.findAll('input[type="number"]');
  expect(inputs).toHaveLength(2);
  expect(
    wrapper
      .findAllComponents(VSwitch)
      .find(
        (control) =>
          control.props("label") ===
          "Ignore subscription download routing during recovery",
      )!
      .props("modelValue"),
  ).toBe(false);
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
      allowDirectRecovery: false,
      nodeDns: "auto",
    },
  });
});

test("persists explicit direct-recovery consent and restores it when reopening", async () => {
  wrapper = mountWithApp(Subscription, {
    props: {
      subscription: {
        ...subscription,
        updateMode: "interval_failsafe",
        updateIntervalMinutes: 15,
      },
    },
  });
  wrapper
    .findAllComponents(VSwitch)
    .find(
      (control) =>
        control.props("label") ===
        "Ignore subscription download routing during recovery",
    )!
    .vm.$emit("update:modelValue", true);
  await flushPromises();
  await wrapper
    .findAll("button")
    .find((button) => button.text() === "Save and Apply")!
    .trigger("click");
  await flushPromises();
  const saved = vi.mocked(patchSubscription).mock.calls[0][0]
    .subscription as TouchSubscription;
  expect(saved.allowDirectRecovery).toBe(true);
  wrapper.unmount();
  wrapper = mountWithApp(Subscription, { props: { subscription: saved } });
  expect(
    wrapper
      .findAllComponents(VSwitch)
      .find(
        (control) =>
          control.props("label") ===
          "Ignore subscription download routing during recovery",
      )!
      .props("modelValue"),
  ).toBe(true);
  wrapper
    .findAllComponents(VSwitch)
    .find(
      (control) =>
        control.props("label") ===
        "Ignore subscription download routing during recovery",
    )!
    .vm.$emit("update:modelValue", false);
  await flushPromises();
  await wrapper
    .findAll("button")
    .find((button) => button.text() === "Save and Apply")!
    .trigger("click");
  await flushPromises();
  expect(vi.mocked(patchSubscription).mock.calls[1][0].subscription).toEqual(
    expect.objectContaining({ allowDirectRecovery: false }),
  );
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

async function save() {
  await wrapper
    .findAll("button")
    .find((b) => b.text() === "Save and Apply")!
    .trigger("click");
  await flushPromises();
}

test("old subscriptions default to auto without changing the incoming row", async () => {
  wrapper = mountWithApp(Subscription, { props: { subscription } });
  await save();
  expect(patchSubscription).toHaveBeenCalledWith({
    subscription: { ...subscription, nodeDns: "auto" },
  });
  expect(wrapper.emitted("close")).toEqual([[true]]);
  expect(subscription).not.toHaveProperty("nodeDns");
});

test("edits an IPv6 node DNS endpoint", async () => {
  const endpoint = "udp://[2001:db8::1]:5353";
  wrapper = mountWithApp(Subscription, {
    props: { subscription: { ...subscription, nodeDns: endpoint } },
  });
  const input = wrapper.get(".node-dns input");
  expect((input.element as HTMLInputElement).value).toBe(endpoint);
  expect(input.attributes("dir")).toBe("ltr");
  await input.setValue(" tls://1.1.1.1:853 ");
  await save();
  expect(patchSubscription).toHaveBeenCalledWith({
    subscription: { ...subscription, nodeDns: "tls://1.1.1.1:853" },
  });
});

test("an empty node DNS resets the subscription to auto", async () => {
  wrapper = mountWithApp(Subscription, {
    props: { subscription: { ...subscription, nodeDns: "tcp://1.1.1.1:53" } },
  });
  await wrapper.get(".node-dns input").setValue(" ");
  await save();
  expect(patchSubscription).toHaveBeenCalledWith({
    subscription: { ...subscription, nodeDns: "auto" },
  });
});

test("a rejected node DNS leaves the dialog open and retryable", async () => {
  wrapper = mountWithApp(Subscription, {
    props: { subscription: { ...subscription, nodeDns: "tls://dns.example" } },
  });
  vi.mocked(patchSubscription).mockRejectedValueOnce(new Error("Invalid DNS"));
  await save();
  expect(wrapper.emitted("close")).toBeUndefined();
  await wrapper.get(".node-dns input").setValue("auto");
  await save();
  expect(wrapper.emitted("close")).toEqual([[true]]);
});
