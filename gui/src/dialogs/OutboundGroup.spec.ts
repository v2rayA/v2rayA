// @vitest-environment happy-dom
import { afterEach, beforeEach, expect, test, vi } from "vitest";
import { flushPromises, type VueWrapper } from "@vue/test-utils";
import { VSelect, VSwitch, VTextField } from "vuetify/components";
import type * as Api from "@/api";
import { getOutbound, postOutboundRefresh, putOutbound } from "@/api";
import { mountWithApp } from "@/test/mount";
import OutboundGroup from "./OutboundGroup.vue";

vi.mock("@/api", async (original) => ({
  ...(await original<typeof Api>()),
  getOutbound: vi.fn(),
  postOutboundRefresh: vi.fn(),
  putOutbound: vi.fn(),
}));

let wrapper: VueWrapper;
beforeEach(async () => {
  vi.clearAllMocks();
  vi.mocked(getOutbound).mockResolvedValue({
    setting: {
      autoAdd: false,
      probeURL: "https://www.gstatic.com/generate_204",
      probeInterval: "60s",
      type: "leastping",
    },
  });
  vi.mocked(putOutbound).mockResolvedValue(undefined);
  vi.mocked(postOutboundRefresh).mockResolvedValue(undefined);
  wrapper = mountWithApp(OutboundGroup, { props: { outbound: "proxy" } });
  await flushPromises();
});
afterEach(() => wrapper.unmount());

test("enabling automatic membership uses the five-minute default and saves the group policy", async () => {
  const toggle = wrapper.getComponent(VSwitch);
  toggle.vm.$emit("update:modelValue", true);
  await flushPromises();
  const interval = wrapper
    .findAllComponents(VTextField)
    .find((field) => field.props("label") === "Probe Interval");
  expect(interval?.props("modelValue")).toBe("300s");
  const automaticHelp = wrapper
    .findAll(".md3-body-large")
    .find((paragraph) => paragraph.text().includes("entire Proxies list"));
  expect(automaticHelp?.exists()).toBe(true);
  expect(
    wrapper
      .findAll(".md3-body-small")
      .find((paragraph) => paragraph.text().includes("catalog updates"))
      ?.classes(),
  ).toContain("text-on-surface-variant");

  await wrapper
    .findAll("button")
    .find((button) => button.text() === "Save")!
    .trigger("click");
  await flushPromises();
  expect(putOutbound).toHaveBeenCalledWith({
    outbound: "proxy",
    setting: {
      autoAdd: true,
      probeURL: "https://www.gstatic.com/generate_204",
      probeInterval: "300s",
      type: "leastping",
    },
  });
});

test("offers connection strategies, explains them on demand, and saves keep-current", async () => {
  const select = wrapper.getComponent(VSelect);
  expect(select.props("label")).toBe("Connection Strategy");
  expect(select.props("items")).toEqual([
    { value: "leastping", title: "Lowest latency" },
    { value: "keepcurrent", title: "Keep current until failure" },
    { value: "roundrobin", title: "Round robin" },
    { value: "random", title: "Random" },
    { value: "firstavailable", title: "First available" },
  ]);

  select.vm.$emit("update:modelValue", "keepcurrent");
  await flushPromises();
  const help = wrapper.get('[aria-label="About connection strategies"]');
  expect(help.attributes("aria-expanded")).toBe("false");
  await help.trigger("click");
  const details = wrapper.get("#strategy-help");
  expect(details.classes()).toContain("text-on-surface-variant");
  expect(details.text()).toContain("healthy current server");

  await wrapper
    .findAll("button")
    .find((button) => button.text() === "Save")!
    .trigger("click");
  await flushPromises();
  expect(putOutbound).toHaveBeenCalledWith({
    outbound: "proxy",
    setting: {
      autoAdd: false,
      probeURL: "https://www.gstatic.com/generate_204",
      probeInterval: "60s",
      type: "keepcurrent",
    },
  });
});

test("orders fields and refreshes membership with the edited policy", async () => {
  const html = wrapper.html();
  const labels = [
    "Connection Strategy",
    "Probe URL",
    "Probe Interval",
    "Automatically add available servers",
  ];
  expect(labels.every((label) => html.includes(label))).toBe(true);
  expect(labels.map((label) => html.indexOf(label))).toEqual(
    [...labels]
      .map((label) => html.indexOf(label))
      .sort((left, right) => left - right),
  );

  const refresh = () =>
    wrapper
      .findAll("button")
      .find((button) => button.text() === "Update server list")!;
  expect(refresh().attributes("disabled")).toBeDefined();
  wrapper.getComponent(VSwitch).vm.$emit("update:modelValue", true);
  await flushPromises();
  expect(refresh().attributes("disabled")).toBeUndefined();

  await refresh().trigger("click");
  await flushPromises();
  expect(putOutbound).toHaveBeenCalledWith({
    outbound: "proxy",
    setting: {
      autoAdd: true,
      probeURL: "https://www.gstatic.com/generate_204",
      probeInterval: "300s",
      type: "leastping",
    },
  });
  expect(postOutboundRefresh).toHaveBeenCalledWith("proxy");
});
