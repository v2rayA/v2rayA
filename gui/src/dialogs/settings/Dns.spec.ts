// @vitest-environment happy-dom
import {
  afterAll,
  afterEach,
  beforeEach,
  describe,
  expect,
  test,
  vi,
} from "vitest";
import {
  enableAutoUnmount,
  flushPromises,
  type VueWrapper,
} from "@vue/test-utils";
import { mountWithApp } from "@/test/mount";
import { closeAllNotices, noticeState } from "@/composables/useNotify";
import en from "@/locales/en";
import type { Setting } from "@/api/types";

const api = vi.hoisted(() => ({
  getDnsRules: vi.fn(),
  getOutbounds: vi.fn(),
  getSetting: vi.fn(),
  putDnsRules: vi.fn(),
  postNodeDnsOptions: vi.fn(),
}));
vi.mock("@/api", () => api);

import Dns from "./Dns.vue";

enableAutoUnmount(afterEach);
afterAll(() => vi.unstubAllGlobals());

function button(w: VueWrapper, text: string) {
  return w.findAll("button").find((b) => b.text() === text)!;
}

function servers(w: VueWrapper) {
  return w.findAll<HTMLInputElement>(".v-text-field:not(.v-select) input");
}

/** the rule editor's outbounds follow the mode select, which comes first */
function outbounds(w: VueWrapper) {
  return w
    .findAllComponents({ name: "VSelect" })
    .filter((select) => select.props("label") === en.dns.colOutbound);
}

async function chooseMode(w: VueWrapper, title: string) {
  await w
    .findAllComponents({ name: "VSelect" })[0]
    .vm.$emit(
      "update:modelValue",
      title === en.dns.modeOff
        ? "off"
        : title === en.dns.modeService
          ? "service"
          : "hijack",
    );
  await flushPromises();
}

const stored = [
  {
    server: "https://dns.example/dns-query",
    domains: "geosite:cn\ndomain:example.com",
    outbound: "regional",
  },
  { server: "9.9.9.9", domains: "", outbound: "direct" },
];

const setting = (dnsMode: Setting["dnsMode"]): { setting: Setting } => ({
  setting: { transparent: "proxy", dnsMode } as Setting,
});

describe("the DNS settings dialog", () => {
  beforeEach(() => {
    // Happy DOM omits the optional viewport API used by Vuetify's menus.
    vi.stubGlobal("visualViewport", undefined);
    closeAllNotices();
    api.getDnsRules.mockReset().mockResolvedValue({ rules: stored });
    api.getOutbounds
      .mockReset()
      .mockResolvedValue({ outbounds: ["proxy", "regional"] });
    api.getSetting.mockReset().mockResolvedValue(setting("hijack"));
    api.putDnsRules.mockReset().mockResolvedValue({});
    api.postNodeDnsOptions.mockReset().mockResolvedValue({
      options: [
        { value: "auto", url: "udp://9.9.9.9:53", category: "auto" },
        {
          value: "udp://9.9.9.9:53",
          url: "udp://9.9.9.9:53",
          category: "direct",
        },
      ],
    });
  });

  test("loads the fields and outbound choices, then saves the combined rules body", async () => {
    const w = mountWithApp(Dns);
    await flushPromises();
    expect(api.getDnsRules).toHaveBeenCalledOnce();
    expect(api.getOutbounds).toHaveBeenCalledOnce();
    expect(servers(w).map((input) => input.element.value)).toEqual(
      stored.map((r) => r.server),
    );
    expect(
      w
        .findAll<HTMLTextAreaElement>('textarea:not([aria-hidden="true"])')
        .map((input) => input.element.value),
    ).toEqual(stored.map((r) => r.domains));
    const selects = outbounds(w);
    expect(selects.map((select) => select.props("modelValue"))).toEqual([
      "regional",
      "direct",
    ]);
    await selects[0].find(".v-field").trigger("mousedown");
    await flushPromises();
    expect(
      Array.from(document.querySelectorAll('[role="option"]')).map((option) =>
        option.textContent?.trim(),
      ),
    ).toEqual(["direct", "proxy", "regional"]);
    await servers(w)[0].setValue("1.1.1.1");
    await w.get("form").trigger("submit");
    await flushPromises();
    expect(api.putDnsRules).toHaveBeenCalledExactlyOnceWith({
      rules: [{ ...stored[0], server: "1.1.1.1" }, stored[1]],
    });
    expect(w.emitted("close")).toEqual([[true]]);
  });

  test("opens on the stored mode and sends it as the only field when it changes", async () => {
    const w = mountWithApp(Dns);
    await flushPromises();
    expect(
      w.findAllComponents({ name: "VSelect" })[0].props("modelValue"),
    ).toBe("hijack");
    expect(w.text()).toContain(en.dns.modeHijackHelp);
    await chooseMode(w, en.dns.modeService);
    expect(w.text()).toContain(en.dns.modeServiceHelp);
    await w.get("form").trigger("submit");
    await flushPromises();
    // the rules did not change, so the dialog owes the service the mode alone
    expect(api.putDnsRules).toHaveBeenCalledExactlyOnceWith({
      dnsMode: "service",
    });
    expect(w.emitted("close")).toEqual([[true]]);
  });

  test("reads a mode a service that predates it answers through the opt-out", async () => {
    api.getSetting.mockResolvedValue({
      setting: { transparent: "proxy", dnsHijack: "no" } as Setting,
    });
    const w = mountWithApp(Dns);
    await flushPromises();
    expect(
      w.findAllComponents({ name: "VSelect" })[0].props("modelValue"),
    ).toBe("off");
  });

  test("keeps an untouched configuration out of the service entirely", async () => {
    const w = mountWithApp(Dns);
    await flushPromises();
    expect(button(w, en.operations.save).attributes("disabled")).toBeDefined();
    await button(w, en.operations.save).trigger("click");
    await flushPromises();
    expect(api.putDnsRules).not.toHaveBeenCalled();
    expect(w.emitted("close")).toBeUndefined();
  });

  test("hides the rules when the mode is off and keeps the stored ones", async () => {
    const w = mountWithApp(Dns);
    await flushPromises();
    await chooseMode(w, en.dns.modeOff);
    expect(w.text()).toContain(
      en.dns.offKeepsRules.replace("{n}", String(stored.length)),
    );
    expect(servers(w)).toHaveLength(0);
    expect(w.find(".node-dns").exists()).toBe(false);
    expect(w.text()).not.toContain(en.dns.addRule);
    await w.get("form").trigger("submit");
    await flushPromises();
    // nothing would read the rules, so dropping them would lose the
    // configuration the mode comes back to
    expect(api.putDnsRules).toHaveBeenCalledExactlyOnceWith({ dnsMode: "off" });
    api.getSetting.mockResolvedValue(setting("off"));
    const reopened = mountWithApp(Dns);
    await flushPromises();
    expect(reopened.text()).toContain(
      en.dns.offKeepsRules.replace("{n}", String(stored.length)),
    );
    await chooseMode(reopened, en.dns.modeHijack);
    expect(servers(reopened).map((input) => input.element.value)).toEqual(
      stored.map((r) => r.server),
    );
  });

  test("does not send rules edits that an off mode has put out of reach", async () => {
    const w = mountWithApp(Dns);
    await flushPromises();
    await servers(w)[0].setValue("1.1.1.1");
    await chooseMode(w, en.dns.modeOff);
    await w.get("form").trigger("submit");
    await flushPromises();
    expect(api.putDnsRules).toHaveBeenCalledExactlyOnceWith({ dnsMode: "off" });
  });

  test("keeps edits after a failed save and allows another attempt", async () => {
    api.putDnsRules.mockRejectedValueOnce(new Error("DNS rejected"));
    const w = mountWithApp(Dns);
    await flushPromises();
    await servers(w)[0].setValue("1.1.1.1");
    await w.get("form").trigger("submit");
    await flushPromises();
    expect(w.emitted("close")).toBeUndefined();
    expect(noticeState.current?.text).toContain("DNS rejected");
    expect(servers(w)[0].element.value).toBe("1.1.1.1");
    await w.get("form").trigger("submit");
    await flushPromises();
    expect(api.putDnsRules).toHaveBeenLastCalledWith({
      rules: [{ ...stored[0], server: "1.1.1.1" }, stored[1]],
    });
    expect(w.emitted("close")).toEqual([[true]]);
  });

  test("saves edits in order, omits blank servers and preserves nonblank whitespace", async () => {
    const w = mountWithApp(Dns);
    await flushPromises();
    await servers(w)[0].setValue(" 1.1.1.1 ");
    await w
      .findAll('textarea:not([aria-hidden="true"])')[0]
      .setValue("geosite:private\nexample.org");
    outbounds(w)[0].vm.$emit("update:modelValue", "proxy");
    await button(w, en.dns.addRule).trigger("click");
    await servers(w)[2].setValue("   ");
    await w.findAll('button[aria-label="Delete"]')[1].trigger("click");
    await w.get("form").trigger("submit");
    await flushPromises();
    expect(api.putDnsRules).toHaveBeenCalledExactlyOnceWith({
      rules: [
        {
          server: " 1.1.1.1 ",
          domains: "geosite:private\nexample.org",
          outbound: "proxy",
        },
      ],
    });
  });

  test("uses the three legacy defaults when no rules are stored", async () => {
    api.getDnsRules.mockResolvedValue({ rules: [] });
    const w = mountWithApp(Dns);
    await flushPromises();
    expect(servers(w).map((input) => input.element.value)).toEqual([
      "localhost",
      "223.5.5.5",
      "8.8.8.8",
    ]);
  });

  test("restores the defaults over the stored rules", async () => {
    const w = mountWithApp(Dns);
    await flushPromises();
    await servers(w)[0].setValue("changed");
    await w.findAll('button[aria-label="Delete"]')[1].trigger("click");
    await button(w, en.dns.resetDefault).trigger("click");
    await w.get("form").trigger("submit");
    await flushPromises();
    expect(api.putDnsRules).toHaveBeenCalledExactlyOnceWith({
      rules: [
        { server: "localhost", domains: "geosite:private", outbound: "direct" },
        { server: "223.5.5.5", domains: "geosite:cn", outbound: "direct" },
        { server: "8.8.8.8", domains: "", outbound: "proxy" },
      ],
    });
  });

  test("normalizes missing fields and refuses to save without a DNS server", async () => {
    api.getDnsRules.mockResolvedValue({ rules: [{}] });
    const w = mountWithApp(Dns);
    await flushPromises();
    await servers(w)[0].setValue(" \t ");
    await w.get("form").trigger("submit");
    expect(api.putDnsRules).not.toHaveBeenCalled();
    expect(noticeState.current?.kind).toBe("warning");
    expect(w.emitted("close")).toBeUndefined();
    await servers(w)[0].setValue("localhost");
    await w.get("form").trigger("submit");
    await flushPromises();
    expect(api.putDnsRules).toHaveBeenCalledExactlyOnceWith({
      rules: [{ server: "localhost", domains: "", outbound: "direct" }],
    });
  });

  test("does not overwrite settings after a failed load and cancel sends nothing", async () => {
    api.getDnsRules.mockRejectedValueOnce(new Error("DNS unavailable"));
    const w = mountWithApp(Dns);
    await flushPromises();
    expect(w.get(".v-alert").text()).toContain("DNS unavailable");
    expect(button(w, en.operations.save).attributes("disabled")).toBeDefined();
    await w.get("form").trigger("submit");
    await button(w, en.operations.cancel).trigger("click");
    expect(api.putDnsRules).not.toHaveBeenCalled();
    expect(w.emitted("close")).toEqual([[]]);
  });

  test("cancel discards a chosen mode as well as the edits", async () => {
    const w = mountWithApp(Dns);
    await flushPromises();
    await chooseMode(w, en.dns.modeService);
    await servers(w)[0].setValue("1.1.1.1");
    await button(w, en.operations.cancel).trigger("click");
    await flushPromises();
    expect(api.putDnsRules).not.toHaveBeenCalled();
    expect(w.emitted("close")).toEqual([[]]);
  });
  function nodeSelect(w: VueWrapper) {
    return w
      .getComponent({ name: "NodeDnsChoice" })
      .getComponent({ name: "VSelect" });
  }
  async function refresh(w: VueWrapper) {
    await w.get(".node-dns button").trigger("click");
    await flushPromises();
  }
  test("places node DNS last and previews a new rule before selecting and saving it", async () => {
    const w = mountWithApp(Dns);
    await flushPromises();
    expect(
      w.findAllComponents({ name: "VSelect" }).at(-1)!.props("label"),
    ).toBe(en.nodeDns.title);
    await servers(w)[1].setValue("1.1.1.1");
    api.postNodeDnsOptions.mockResolvedValue({
      options: [
        { value: "auto", url: "udp://1.1.1.1:53", category: "auto" },
        {
          value: "udp://1.1.1.1:53",
          url: "udp://1.1.1.1:53",
          category: "direct",
        },
      ],
    });
    await refresh(w);
    expect(api.postNodeDnsOptions).toHaveBeenLastCalledWith({
      rules: [stored[0], { ...stored[1], server: "1.1.1.1" }],
    });
    await nodeSelect(w).vm.$emit("update:modelValue", "udp://1.1.1.1:53");
    await w.get("form").trigger("submit");
    await flushPromises();
    expect(api.putDnsRules).toHaveBeenCalledExactlyOnceWith({
      rules: [stored[0], { ...stored[1], server: "1.1.1.1" }],
      nodeDns: "udp://1.1.1.1:53",
    });
  });
  test("hides the node choice when off, restores draft choice when enabled, and omits hidden edits", async () => {
    const w = mountWithApp(Dns);
    await flushPromises();
    await nodeSelect(w).vm.$emit("update:modelValue", "udp://9.9.9.9:53");
    await chooseMode(w, en.dns.modeOff);
    expect(w.find(".node-dns").exists()).toBe(false);
    await chooseMode(w, en.dns.modeService);
    expect(nodeSelect(w).props("modelValue")).toBe("udp://9.9.9.9:53");
    await chooseMode(w, en.dns.modeOff);
    await w.get("form").trigger("submit");
    await flushPromises();
    expect(api.putDnsRules).toHaveBeenCalledExactlyOnceWith({ dnsMode: "off" });
  });
  test("does not load node options while opening in off mode", async () => {
    api.getSetting.mockResolvedValue(setting("off"));
    const w = mountWithApp(Dns);
    await flushPromises();
    expect(w.find(".node-dns").exists()).toBe(false);
    expect(api.postNodeDnsOptions).not.toHaveBeenCalled();
  });
  test("keeps complete IPv6 DoH URLs in auto display and saves auto as the value", async () => {
    api.getSetting.mockResolvedValue({
      setting: { dnsMode: "service", nodeDns: "udp://9.9.9.9:53" },
    });
    const url = "https://[2001:db8::53]:443/dns-query?token=test";
    api.postNodeDnsOptions.mockResolvedValue({
      options: [{ value: "auto", url, category: "auto" }],
    });
    const w = mountWithApp(Dns);
    await flushPromises();
    await nodeSelect(w).vm.$emit("update:modelValue", "auto");
    expect(w.get(".node-dns__value").text()).toContain(url);
    await w.get("form").trigger("submit");
    await flushPromises();
    expect(api.putDnsRules).toHaveBeenCalledExactlyOnceWith({
      nodeDns: "auto",
    });
  });
  test("retries preview errors before displaying and saving auto for a removed source", async () => {
    api.getSetting.mockResolvedValue({
      setting: { dnsMode: "service", nodeDns: "tls://192.0.2.53:853" },
    });
    api.postNodeDnsOptions.mockRejectedValueOnce(new Error("offline"));
    const w = mountWithApp(Dns);
    await flushPromises();
    expect(w.get(".node-dns").text()).toContain("offline");
    expect(nodeSelect(w).props("modelValue")).toBe("tls://192.0.2.53:853");
    expect(w.find(".node-dns__fallback").exists()).toBe(false);
    await refresh(w);
    expect(nodeSelect(w).props("modelValue")).toBe("auto");
    expect(w.get(".node-dns__fallback").text()).toContain(
      "tls://192.0.2.53:853",
    );
    expect(api.putDnsRules).not.toHaveBeenCalled();
    await chooseMode(w, en.dns.modeHijack);
    await w.get("form").trigger("submit");
    await flushPromises();
    expect(api.putDnsRules).toHaveBeenCalledExactlyOnceWith({
      dnsMode: "hijack",
      nodeDns: "auto",
    });
  });
  test("discards stale preview responses after the rules change", async () => {
    const w = mountWithApp(Dns);
    await flushPromises();
    let resolve!: (result: unknown) => void;
    api.postNodeDnsOptions.mockReturnValueOnce(
      new Promise((done) => {
        resolve = done;
      }),
    );
    await w.get(".node-dns button").trigger("click");
    await servers(w)[1].setValue("1.1.1.1");
    api.postNodeDnsOptions.mockResolvedValue({
      options: [{ value: "auto", url: "udp://1.1.1.1:53", category: "auto" }],
    });
    await refresh(w);
    resolve({
      options: [
        { value: "auto", url: "udp://192.0.2.53:53", category: "auto" },
      ],
    });
    await flushPromises();
    expect(w.get(".node-dns__value").text()).toContain("udp://1.1.1.1:53");
  });
  test("shows auto and the missing explicit address below it, then clears the warning on selection", async () => {
    const address = "udp://192.0.2.53:53";
    api.getSetting.mockResolvedValue({
      setting: {
        dnsMode: "service",
        nodeDns: address,
      },
    });
    const w = mountWithApp(Dns);
    await flushPromises();
    expect(nodeSelect(w).props("modelValue")).toBe("auto");
    expect(w.get(".node-dns__fallback").text()).toBe(
      en.nodeDns.resetToAuto.replace("{address}", address),
    );
    expect(api.putDnsRules).not.toHaveBeenCalled();
    await nodeSelect(w).vm.$emit("update:modelValue", "udp://9.9.9.9:53");
    expect(w.text()).not.toContain(
      en.nodeDns.resetToAuto.replace("{address}", address),
    );
    await w.get("form").trigger("submit");
    await flushPromises();
    expect(api.putDnsRules).toHaveBeenCalledExactlyOnceWith({
      nodeDns: "udp://9.9.9.9:53",
    });
  });
  test("restores the explicit display when its source returns, and cancel never writes the fallback", async () => {
    const address = "udp://192.0.2.53:53";
    api.getSetting.mockResolvedValue({
      setting: { dnsMode: "service", nodeDns: address },
    });
    const w = mountWithApp(Dns);
    await flushPromises();
    expect(nodeSelect(w).props("modelValue")).toBe("auto");
    api.postNodeDnsOptions.mockResolvedValue({
      options: [
        { value: "auto", url: "udp://9.9.9.9:53", category: "auto" },
        { value: address, url: address, category: "localhost" },
      ],
    });
    await refresh(w);
    expect(nodeSelect(w).props("modelValue")).toBe(address);
    expect(w.find(".node-dns__fallback").exists()).toBe(false);
    expect(button(w, en.operations.save).attributes("disabled")).toBeDefined();
    api.postNodeDnsOptions.mockResolvedValue({
      options: [{ value: "auto", url: "udp://9.9.9.9:53", category: "auto" }],
    });
    await refresh(w);
    await button(w, en.operations.cancel).trigger("click");
    expect(w.emitted("close")).toEqual([[]]);
    expect(api.putDnsRules).not.toHaveBeenCalled();
  });
  test("rejects a missing source when no auto candidate exists", async () => {
    const address = "udp://192.0.2.53:53";
    api.getSetting.mockResolvedValue({
      setting: { dnsMode: "service", nodeDns: address },
    });
    api.postNodeDnsOptions.mockResolvedValue({ options: [] });
    const w = mountWithApp(Dns);
    await flushPromises();
    expect(nodeSelect(w).props("modelValue")).toBe(address);
    expect(w.find(".node-dns__fallback").exists()).toBe(false);
    await chooseMode(w, en.dns.modeHijack);
    await w.get("form").trigger("submit");
    await flushPromises();
    expect(api.putDnsRules).not.toHaveBeenCalled();
    expect(w.get(".node-dns").text()).toContain(en.nodeDns.unavailable);
  });
});
