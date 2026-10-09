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
  putSetting: vi.fn(),
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
  return w.findAllComponents({ name: "VSelect" }).slice(1);
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
    api.putSetting.mockReset().mockResolvedValue({});
  });

  test("loads the fields and outbound choices, then saves the legacy array body", async () => {
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
    expect(api.putDnsRules).toHaveBeenCalledExactlyOnceWith([
      { ...stored[0], server: "1.1.1.1" },
      stored[1],
    ]);
    expect(api.putSetting).not.toHaveBeenCalled();
    expect(w.emitted("close")).toEqual([[]]);
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
    expect(api.putDnsRules).not.toHaveBeenCalled();
    expect(api.putSetting).toHaveBeenCalledExactlyOnceWith({
      dnsMode: "service",
    });
    expect(w.emitted("close")).toEqual([[]]);
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
    expect(api.putSetting).not.toHaveBeenCalled();
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
    expect(w.text()).not.toContain(en.dns.addRule);
    await w.get("form").trigger("submit");
    await flushPromises();
    // nothing would read the rules, so dropping them would lose the
    // configuration the mode comes back to
    expect(api.putDnsRules).not.toHaveBeenCalled();
    expect(api.putSetting).toHaveBeenCalledExactlyOnceWith({ dnsMode: "off" });
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
    expect(api.putDnsRules).not.toHaveBeenCalled();
    expect(api.putSetting).toHaveBeenCalledExactlyOnceWith({ dnsMode: "off" });
  });

  test("names the refusal when the mode follows rules the service has stored", async () => {
    api.putSetting.mockRejectedValue(new Error("mode refused"));
    const w = mountWithApp(Dns);
    await flushPromises();
    await servers(w)[0].setValue("1.1.1.1");
    await chooseMode(w, en.dns.modeService);
    await w.get("form").trigger("submit");
    await flushPromises();
    expect(api.putDnsRules).toHaveBeenCalledOnce();
    expect(noticeState.current?.kind).toBe("warning");
    expect(noticeState.current?.text).toContain("mode refused");
    expect(noticeState.current?.text).toContain(
      en.dns.modeSaveFailed.split("{message}")[0],
    );
    expect(w.emitted("close")).toBeUndefined();
    // the choice returns to the mode that is in force, and the stored rules
    // are no longer owed
    expect(
      w.findAllComponents({ name: "VSelect" })[0].props("modelValue"),
    ).toBe("hijack");
    await chooseMode(w, en.dns.modeService);
    await w.get("form").trigger("submit");
    await flushPromises();
    expect(api.putDnsRules).toHaveBeenCalledOnce();
    expect(api.putSetting).toHaveBeenCalledTimes(2);
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
    expect(api.putDnsRules).toHaveBeenLastCalledWith([
      { ...stored[0], server: "1.1.1.1" },
      stored[1],
    ]);
    expect(w.emitted("close")).toEqual([[]]);
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
    expect(api.putDnsRules).toHaveBeenCalledExactlyOnceWith([
      {
        server: " 1.1.1.1 ",
        domains: "geosite:private\nexample.org",
        outbound: "proxy",
      },
    ]);
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
    expect(api.putDnsRules).toHaveBeenCalledExactlyOnceWith([
      { server: "localhost", domains: "geosite:private", outbound: "direct" },
      { server: "223.5.5.5", domains: "geosite:cn", outbound: "direct" },
      { server: "8.8.8.8", domains: "", outbound: "proxy" },
    ]);
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
    expect(api.putDnsRules).toHaveBeenCalledExactlyOnceWith([
      { server: "localhost", domains: "", outbound: "direct" },
    ]);
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
    expect(api.putSetting).not.toHaveBeenCalled();
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
    expect(api.putSetting).not.toHaveBeenCalled();
    expect(w.emitted("close")).toEqual([[]]);
  });
});
