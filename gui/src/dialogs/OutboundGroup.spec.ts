// @vitest-environment happy-dom
import { afterEach, expect, test, vi } from "vitest";
import { enableAutoUnmount, flushPromises } from "@vue/test-utils";
import { VSelect } from "vuetify/components";
import { mountWithApp } from "@/test/mount";
import { closeAllNotices } from "@/composables/useNotify";
import en from "@/locales/en";
import OutboundGroup from "./OutboundGroup.vue";

const api = vi.hoisted(() => ({ getOutbound: vi.fn(), putOutbound: vi.fn() }));
vi.mock("@/api", () => api);
enableAutoUnmount(afterEach);
afterEach(closeAllNotices);

test("saves all-catalog membership, reopens it, and can turn it off without changing probes", async () => {
  let saved = {
    autoAdd: false,
    probeURL: "https://example.com/",
    probeInterval: "60s",
    type: "leastping",
  };
  api.getOutbound.mockImplementation(async () => ({ setting: { ...saved } }));
  api.putOutbound.mockImplementation(async ({ setting }) => {
    saved = { ...setting };
  });
  let wrapper = mountWithApp(OutboundGroup, { props: { outbound: "proxy" } });
  await flushPromises();
  expect(wrapper.text()).toContain(en.outbound.autoAddHelp);
  await wrapper.get('input[type="checkbox"]').setValue(true);
  await wrapper
    .findAll("button")
    .find((b) => b.text() === en.operations.save)!
    .trigger("click");
  await flushPromises();
  expect(saved).toEqual({
    autoAdd: true,
    probeURL: "https://example.com/",
    probeInterval: "60s",
    type: "leastping",
  });
  expect(wrapper.emitted("close")).toEqual([[true]]);
  wrapper.unmount();
  wrapper = mountWithApp(OutboundGroup, { props: { outbound: "proxy" } });
  await flushPromises();
  expect(
    (wrapper.get('input[type="checkbox"]').element as HTMLInputElement).checked,
  ).toBe(true);
  await wrapper.get('input[type="checkbox"]').setValue(false);
  await wrapper
    .findAll("button")
    .find((b) => b.text() === en.operations.save)!
    .trigger("click");
  await flushPromises();
  expect(saved.autoAdd).toBe(false);
  expect(saved.probeInterval).toBe("60s");
});

test.each(["leastping", "leastload", "roundrobin", "random"])(
  "saves and restores the native %s strategy",
  async (strategy) => {
    vi.clearAllMocks();
    const setting = {
      autoAdd: false,
      probeURL: "https://www.gstatic.com/generate_204",
      probeInterval: "60s",
      type: "leastping",
    };
    api.getOutbound.mockResolvedValue({ setting });
    api.putOutbound.mockResolvedValue(undefined);
    let wrapper = mountWithApp(OutboundGroup, { props: { outbound: "proxy" } });
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
    expect(api.putOutbound).toHaveBeenCalledWith({
      outbound: "proxy",
      setting: { ...setting, type: strategy },
    });
    expect(wrapper.emitted("close")).toEqual([[true]]);

    wrapper.unmount();
    api.getOutbound.mockResolvedValue({
      setting: { ...setting, type: strategy },
    });
    wrapper = mountWithApp(OutboundGroup, { props: { outbound: "proxy" } });
    await flushPromises();
    expect(wrapper.getComponent(VSelect).props("modelValue")).toBe(strategy);
  },
);
