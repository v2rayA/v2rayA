// @vitest-environment happy-dom
import { afterEach, beforeEach, expect, test, vi } from "vitest";
import { flushPromises, type VueWrapper } from "@vue/test-utils";
import { VSelect, VSwitch, VTextField } from "vuetify/components";
import type * as Api from "@/api";
import { getOutbound, putOutbound } from "@/api";
import { mountWithApp } from "@/test/mount";
import OutboundGroup from "./OutboundGroup.vue";

vi.mock("@/api", async (original) => ({
  ...(await original<typeof Api>()),
  getOutbound: vi.fn(),
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
  expect(wrapper.find(".md3-body-large").text()).toContain(
    "entire Proxies list",
  );
  expect(wrapper.find(".md3-body-small").classes()).toContain(
    "text-on-surface-variant",
  );

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

test("offers four connection strategies and saves keep-current", async () => {
  const select = wrapper.getComponent(VSelect);
  expect(select.props("label")).toBe("Connection Strategy");
  expect(select.props("items")).toEqual([
    { value: "leastping", title: "Lowest latency" },
    { value: "keepcurrent", title: "Keep current until failure" },
    { value: "roundrobin", title: "Round robin" },
    { value: "random", title: "Random" },
  ]);

  select.vm.$emit("update:modelValue", "keepcurrent");
  await flushPromises();
  const details = wrapper.findAll(".md3-body-small").at(-1)!;
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
