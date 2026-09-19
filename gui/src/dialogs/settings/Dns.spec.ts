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

const api = vi.hoisted(() => ({
  getDnsRules: vi.fn(),
  getOutbounds: vi.fn(),
  putDnsRules: vi.fn(),
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

const stored = [
  {
    server: "https://dns.example/dns-query",
    domains: "geosite:cn\ndomain:example.com",
    outbound: "regional",
  },
  { server: "9.9.9.9", domains: "", outbound: "direct" },
];

describe("the DNS settings dialog", () => {
  beforeEach(() => {
    // Happy DOM omits the optional viewport API used by Vuetify's menus.
    vi.stubGlobal("visualViewport", undefined);
    closeAllNotices();
    api.getDnsRules.mockReset().mockResolvedValue({ rules: stored });
    api.getOutbounds
      .mockReset()
      .mockResolvedValue({ outbounds: ["proxy", "regional"] });
    api.putDnsRules.mockReset().mockResolvedValue({});
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
    const outbounds = w.findAllComponents({ name: "VSelect" });
    expect(outbounds.map((select) => select.props("modelValue"))).toEqual([
      "regional",
      "direct",
    ]);
    await outbounds[0].find(".v-field").trigger("mousedown");
    await flushPromises();
    expect(
      Array.from(document.querySelectorAll('[role="option"]')).map((option) =>
        option.textContent?.trim(),
      ),
    ).toEqual(["direct", "proxy", "regional"]);
    await w.get("form").trigger("submit");
    await flushPromises();
    expect(api.putDnsRules).toHaveBeenCalledExactlyOnceWith(stored);
    expect(w.emitted("close")).toEqual([[]]);
  });

  test("saves edits in order, omits blank servers and preserves nonblank whitespace", async () => {
    const w = mountWithApp(Dns);
    await flushPromises();
    await servers(w)[0].setValue(" 1.1.1.1 ");
    await w
      .findAll('textarea:not([aria-hidden="true"])')[0]
      .setValue("geosite:private\nexample.org");
    w.findAllComponents({ name: "VSelect" })[0].vm.$emit(
      "update:modelValue",
      "proxy",
    );
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

  test("uses and restores the three legacy defaults when no rules are stored", async () => {
    api.getDnsRules.mockResolvedValue({ rules: [] });
    const w = mountWithApp(Dns);
    await flushPromises();
    expect(servers(w).map((input) => input.element.value)).toEqual([
      "localhost",
      "223.5.5.5",
      "8.8.8.8",
    ]);
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
});
