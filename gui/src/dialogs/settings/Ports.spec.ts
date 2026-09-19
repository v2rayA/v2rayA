// @vitest-environment happy-dom
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { enableAutoUnmount, flushPromises } from "@vue/test-utils";
import type * as ApiClient from "@/api/client";
import { mountWithApp } from "@/test/mount";
import { closeAllDialogs, dialogState } from "@/composables/useDialog";
import { closeAllNotices, noticeState } from "@/composables/useNotify";
import { useAppStore } from "@/stores/app";
import en from "@/locales/en";

const api = vi.hoisted(() => ({
  getPorts: vi.fn(),
  putPorts: vi.fn(),
  probe: vi.fn(),
  resetSession: vi.fn(),
}));
vi.mock("@/api", () => api);
vi.mock("@/api/client", async (importOriginal) => ({
  ...(await importOriginal<typeof ApiClient>()),
  probe: api.probe,
}));
vi.mock("@/session", () => ({ resetSession: api.resetSession }));

import PortsDialog from "./Ports.vue";

enableAutoUnmount(afterEach);
const loaded = {
  socks5: 20170,
  http: 20171,
  socks5WithPac: 1080,
  httpWithPac: 20172,
  vmess: 12345,
  vmessLink: "vmess://existing",
  api: { port: 23456, services: ["LoggerService", "StatsService"] },
};

beforeEach(() => {
  vi.resetAllMocks();
  localStorage.clear();
  localStorage.setItem("backendAddress", "http://localhost:2017");
  closeAllDialogs();
  closeAllNotices();
  api.getPorts.mockResolvedValue(structuredClone(loaded));
  api.putPorts.mockResolvedValue({});
  api.probe.mockResolvedValue({ code: "SUCCESS", data: { version: "2" } });
});
afterEach(() => {
  closeAllDialogs();
  closeAllNotices();
  localStorage.clear();
});

describe("the address and ports dialog", () => {
  test("loads the fields and saves the legacy integer body without response-only fields", async () => {
    const w = mountWithApp(PortsDialog);
    await flushPromises();
    expect(api.getPorts).toHaveBeenCalledOnce();
    expect(
      w.get<HTMLInputElement>('input[name="backendAddress"]').element.value,
    ).toBe("http://localhost:2017");
    const expected = {
      socks5: 20170,
      http: 20171,
      socks5WithPac: 1080,
      httpWithPac: 20172,
      vmess: 12345,
      api: 23456,
    };
    for (const [name, value] of Object.entries(expected)) {
      expect(
        w.get<HTMLInputElement>(`input[name="${name}"]`).element.value,
      ).toBe(String(value));
      await w.get(`input[name="${name}"]`).setValue(String(value + 1));
    }
    expect(w.findAll(".v-chip").map((chip) => chip.text())).toEqual([
      "LoggerService",
      "StatsService",
    ]);
    w.getComponent({ name: "VSelect" }).vm.$emit("update:modelValue", [
      "HandlerService",
    ]);
    await w
      .get('input[name="backendAddress"]')
      .setValue("http://localhost:2017/");
    await w.get("form").trigger("submit");
    await flushPromises();
    expect(api.putPorts).toHaveBeenCalledWith({
      socks5: 20171,
      http: 20172,
      socks5WithPac: 1081,
      httpWithPac: 20173,
      vmess: 12346,
      api: { port: 23457, services: ["HandlerService"] },
    });
    expect(api.probe).not.toHaveBeenCalled();
    expect(w.emitted("close")).toEqual([[true]]);
    expect(w.find("code").exists()).toBe(false);
    expect(
      w
        .findAll(".v-alert")
        .map((alert) => alert.text())
        .join(" "),
    ).toContain("V2RAYA_ADDRESS");
  });

  test("hides ports for a changed address and switches the session only after probing", async () => {
    let finish!: () => void;
    api.probe.mockImplementation(
      () =>
        new Promise<void>((resolve) => {
          finish = resolve;
        }),
    );
    const w = mountWithApp(PortsDialog);
    await flushPromises();
    await w.get('input[name="backendAddress"]').setValue("/proxy/");
    expect(w.find('input[name="socks5"]').exists()).toBe(false);
    await w.get("form").trigger("submit");
    expect(api.probe).toHaveBeenCalledWith("/proxy");
    expect(useAppStore().backendAddress).toBe("http://localhost:2017");
    expect(api.resetSession).not.toHaveBeenCalled();
    finish();
    await flushPromises();
    expect(useAppStore().backendAddress).toBe("/proxy");
    expect(localStorage.getItem("backendAddress")).toBe("/proxy");
    expect(api.resetSession).toHaveBeenCalledWith({ backendAddress: "/proxy" });
    expect(api.putPorts).not.toHaveBeenCalled();
    expect(w.emitted("close")).toEqual([[true]]);
  });

  test("rejects an invalid address and preserves the session when the probe fails", async () => {
    const w = mountWithApp(PortsDialog);
    await flushPromises();
    await w.get('input[name="backendAddress"]').setValue("localhost:2017");
    await w.get("form").trigger("submit");
    expect(api.probe).not.toHaveBeenCalled();
    api.probe.mockRejectedValue(new Error("offline"));
    await w
      .get('input[name="backendAddress"]')
      .setValue("https://other.example");
    await w.get("form").trigger("submit");
    await flushPromises();
    expect(api.probe).toHaveBeenCalledWith("https://other.example");
    expect(noticeState.current?.text).toContain("offline");
    expect(useAppStore().backendAddress).toBe("http://localhost:2017");
    expect(api.resetSession).not.toHaveBeenCalled();
    expect(w.emitted("close")).toBeUndefined();
  });

  test("opens the VMess link, custom inbound dialog, and a link returned by saving", async () => {
    const w = mountWithApp(PortsDialog);
    await flushPromises();
    await w
      .findAll("button")
      .find((b) => b.text() === en.customAddressPort.portVmessLink)!
      .trigger("click");
    expect(dialogState.stack.at(-1)?.props).toEqual({
      title: en.customAddressPort.portVmessLink,
      link: loaded.vmessLink,
      name: "VMess | v2rayA",
      type: "server",
    });
    closeAllDialogs();
    await w
      .findAll("button")
      .find((b) => b.text() === en.customInbound.title)!
      .trigger("click");
    expect(dialogState.stack.at(-1)?.component.name).toBe(
      "CustomInboundDialog",
    );
    closeAllDialogs();
    api.putPorts.mockResolvedValue({ vmessLink: "vmess://new" });
    await w.get("form").trigger("submit");
    await flushPromises();
    expect(dialogState.stack.at(-1)?.props.link).toBe("vmess://new");
  });

  test("can select the same-origin backend when loading the old backend fails", async () => {
    api.getPorts.mockRejectedValue(new Error("unreachable"));
    const w = mountWithApp(PortsDialog);
    await flushPromises();
    expect(w.find('input[name="socks5"]').exists()).toBe(false);
    expect(w.get(".v-alert").text()).toContain("unreachable");
    await w.get('input[name="backendAddress"]').setValue("");
    await w.get("form").trigger("submit");
    await flushPromises();
    expect(api.probe).toHaveBeenCalledWith("");
    expect(api.resetSession).toHaveBeenCalledWith({ backendAddress: "" });
  });
});
