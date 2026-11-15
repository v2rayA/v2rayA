// @vitest-environment happy-dom
import { afterEach, expect, test, vi } from "vitest";
import { flushPromises, type VueWrapper } from "@vue/test-utils";
import { VSelect } from "vuetify/components";
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
afterEach(() => wrapper?.unmount());

test.each(["leastping", "leastload", "roundrobin", "random"])(
  "saves and restores the native %s strategy",
  async (strategy) => {
    vi.clearAllMocks();
    const setting = {
      probeURL: "https://www.gstatic.com/generate_204",
      probeInterval: "60s",
      type: "leastping",
    };
    vi.mocked(getOutbound).mockResolvedValue({ setting });
    vi.mocked(putOutbound).mockResolvedValue(undefined);
    wrapper = mountWithApp(OutboundGroup, { props: { outbound: "proxy" } });
    await flushPromises();
    const select = wrapper.getComponent(VSelect);
    expect(select.props("items")).toEqual(
      expect.arrayContaining([expect.objectContaining({ value: strategy })]),
    );
    select.vm.$emit("update:modelValue", strategy);
    await wrapper
      .findAll("button")
      .find((button) => button.text() === "Save")!
      .trigger("click");
    await flushPromises();
    expect(putOutbound).toHaveBeenCalledWith({
      outbound: "proxy",
      setting: { ...setting, type: strategy },
    });
    expect(wrapper.emitted("close")).toEqual([[true]]);

    wrapper.unmount();
    vi.mocked(getOutbound).mockResolvedValue({
      setting: { ...setting, type: strategy },
    });
    wrapper = mountWithApp(OutboundGroup, { props: { outbound: "proxy" } });
    await flushPromises();
    expect(wrapper.getComponent(VSelect).props("modelValue")).toBe(strategy);
  },
);
