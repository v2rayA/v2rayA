// @vitest-environment happy-dom
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { enableAutoUnmount, flushPromises } from "@vue/test-utils";
import type { VueWrapper } from "@vue/test-utils";
import { mountWithApp } from "@/test/mount";
import {
  closeAllDialogs,
  closeDialog,
  dialogState,
} from "@/composables/useDialog";
import { closeAllNotices, noticeState } from "@/composables/useNotify";
import en from "@/locales/en";

const api = vi.hoisted(() => ({
  getCustomInbound: vi.fn(),
  getOutbounds: vi.fn(),
  postCustomInbound: vi.fn(),
  deleteCustomInbound: vi.fn(),
}));
vi.mock("@/api", () => api);

import CustomInboundDialog from "./CustomInbound.vue";

enableAutoUnmount(afterEach);
const existing = {
  tag: "work",
  protocol: "http",
  port: 10801,
  outbound: "PROXY",
  outboundType: "routingA",
  routingARules: "default: proxy",
  username: "user",
  password: "secret",
};
beforeEach(() => {
  vi.resetAllMocks();
  closeAllDialogs();
  closeAllNotices();
  api.getCustomInbound.mockResolvedValue({ inbounds: [existing] });
  api.getOutbounds.mockResolvedValue({ outbounds: ["PROXY", "OTHER"] });
  api.postCustomInbound.mockResolvedValue({ inbounds: [existing] });
  api.deleteCustomInbound.mockResolvedValue({ inbounds: [] });
});
afterEach(() => {
  closeAllDialogs();
  closeAllNotices();
});

function select(w: VueWrapper, label: string, value: string) {
  w.findAllComponents({ name: "VSelect" })
    .find((c) => c.props("label") === label)!
    .vm.$emit("update:modelValue", value);
}

describe("the custom inbound dialog", () => {
  test("loads inbounds and outbounds, then adds the legacy RoutingA body and refreshes the list", async () => {
    const w = mountWithApp(CustomInboundDialog);
    await flushPromises();
    expect(api.getCustomInbound).toHaveBeenCalledOnce();
    expect(api.getOutbounds).toHaveBeenCalledOnce();
    expect(w.get(".inbound-list").text()).toContain("work");
    expect(w.get(".inbound-list").text()).toContain("10801");
    expect(w.get(".inbound-list").text()).toContain("HTTP");
    expect(w.get(".inbound-list").text()).toContain("PROXY");
    const outbound = w
      .findAllComponents({ name: "VSelect" })
      .find((c) => c.props("label") === en.customInbound.outbound)!;
    expect(outbound.text()).toContain("PROXY");
    expect(outbound.props("items")).toEqual(["PROXY", "OTHER"]);
    await w.get('input[name="tag"]').setValue("  home  ");
    await w.get('input[name="port"]').setValue("10802");
    select(w, en.customInbound.protocol, "http");
    select(w, en.customInbound.outbound, "OTHER");
    select(w, en.customInbound.outboundType, "routingA");
    await flushPromises();
    await w.get('textarea[name="routingARules"]').setValue("default: proxy\n");
    await w.get('input[name="username"]').setValue(" alice ");
    await w.get('input[name="password"]').setValue(" secret ");
    const body = {
      tag: "home",
      protocol: "http",
      port: 10802,
      outbound: "OTHER",
      outboundType: "routingA",
      routingARules: "default: proxy\n",
      username: "alice",
      password: " secret ",
    };
    api.postCustomInbound.mockResolvedValue({ inbounds: [existing, body] });
    await w.get("form").trigger("submit");
    await flushPromises();
    expect(api.postCustomInbound).toHaveBeenCalledWith(body);
    expect(w.get(".inbound-list").text()).toContain("home");
    expect(w.get<HTMLInputElement>('input[name="tag"]').element.value).toBe("");
    expect(
      w.get<HTMLInputElement>('input[name="password"]').element.value,
    ).toBe("");
    expect(w.find('textarea[name="routingARules"]').exists()).toBe(false);
    expect(outbound.text()).toContain("PROXY");
    expect(w.emitted("close")).toBeUndefined();
  });

  test("does not submit hidden RoutingA rules in direct mode", async () => {
    const w = mountWithApp(CustomInboundDialog);
    await flushPromises();
    await w.get('input[name="tag"]').setValue("direct");
    await w.get('input[name="port"]').setValue("65535");
    select(w, en.customInbound.outboundType, "routingA");
    await flushPromises();
    await w.get('textarea[name="routingARules"]').setValue("default: proxy");
    select(w, en.customInbound.outboundType, "direct");
    await flushPromises();
    await w.get("form").trigger("submit");
    await flushPromises();
    expect(api.postCustomInbound).toHaveBeenCalledWith({
      tag: "direct",
      protocol: "socks",
      port: 65535,
      outbound: "PROXY",
      outboundType: "direct",
      routingARules: "",
      username: "",
      password: "",
    });
  });

  test("requires confirmation before deleting and replaces the list with the response", async () => {
    const w = mountWithApp(CustomInboundDialog);
    await flushPromises();
    const remove = () =>
      w
        .get(`button[aria-label="${en.operations.delete}: work"]`)
        .trigger("click");
    await remove();
    expect(api.deleteCustomInbound).not.toHaveBeenCalled();
    expect(dialogState.stack.at(-1)?.props.message).toContain("work");
    closeDialog(dialogState.stack.at(-1)!.id, false);
    await flushPromises();
    expect(api.deleteCustomInbound).not.toHaveBeenCalled();
    expect(w.get(".inbound-list").text()).toContain("work");
    await remove();
    closeDialog(dialogState.stack.at(-1)!.id, true);
    await flushPromises();
    expect(api.deleteCustomInbound).toHaveBeenCalledWith({ tag: "work" });
    expect(w.get(".inbound-list").text()).not.toContain("work");
    expect(w.get(".inbound-list").text()).toContain(en.customInbound.empty);
  });

  test("validates required fields and retains input after a failed save", async () => {
    api.getOutbounds.mockResolvedValue({ outbounds: [] });
    const w = mountWithApp(CustomInboundDialog);
    await flushPromises();
    await w.get("form").trigger("submit");
    expect(noticeState.current?.text).toBe(en.customInbound.fillAll);
    expect(api.postCustomInbound).not.toHaveBeenCalled();
    closeAllNotices();
    await w.get('input[name="tag"]').setValue("home");
    await w.get('input[name="port"]').setValue("10802");
    await w.get("form").trigger("submit");
    expect(noticeState.current?.text).toBe(en.customInbound.outboundRequired);
    expect(api.postCustomInbound).not.toHaveBeenCalled();
    closeAllNotices();
    select(w, en.customInbound.outbound, "PROXY");
    api.postCustomInbound.mockRejectedValue(new Error("port occupied"));
    await w.get("form").trigger("submit");
    await flushPromises();
    expect(noticeState.current?.text).toContain("port occupied");
    expect(w.get<HTMLInputElement>('input[name="tag"]').element.value).toBe(
      "home",
    );
    expect(w.get(".inbound-list").text()).toContain("work");
  });
});
