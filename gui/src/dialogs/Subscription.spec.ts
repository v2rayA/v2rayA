// @vitest-environment happy-dom
import { afterEach, beforeEach, expect, test, vi } from "vitest";
import { flushPromises, type VueWrapper } from "@vue/test-utils";
import type * as Api from "@/api";
import { patchSubscription } from "@/api";
import { mountWithApp } from "@/test/mount";
import Subscription from "./Subscription.vue";

vi.mock("@/api", async (original) => ({
  ...(await original<typeof Api>()),
  patchSubscription: vi.fn(),
}));
const subscription = {
  id: 1,
  _type: "subscription" as const,
  address: "https://example.test/sub",
  host: "example.test",
  remarks: "Test",
  status: "",
  info: "",
  servers: [],
  autoSelect: false,
};
let wrapper: VueWrapper;
beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(patchSubscription).mockResolvedValue({} as never);
});
afterEach(() => wrapper?.unmount());

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
