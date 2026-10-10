// @vitest-environment happy-dom
import { beforeEach, describe, expect, test, vi } from "vitest";
import type { Setting } from "@/api/types";

const api = vi.hoisted(() => ({ putDnsRules: vi.fn() }));
vi.mock("@/api", () => api);

import { resolveDnsMode, saveDnsSettings } from "./dnsModel";

const setting = (fields: Partial<Setting>) => fields as Setting;

describe("the mode a service stores", () => {
  test("reads the mode it knows as it is", () => {
    for (const mode of ["off", "service", "hijack"] as const)
      expect(resolveDnsMode(setting({ dnsMode: mode }))).toBe(mode);
  });

  // a service that predates the mode carries the decision in the opt-out, and
  // anything but an explicit refusal there is the historical interception
  test("falls back to the opt-out when no mode is stored", () => {
    expect(resolveDnsMode(setting({ dnsHijack: "no" }))).toBe("off");
    expect(resolveDnsMode(setting({ dnsMode: "", dnsHijack: "no" }))).toBe(
      "off",
    );
    expect(resolveDnsMode(setting({ dnsHijack: "yes" }))).toBe("hijack");
    expect(resolveDnsMode(setting({ dnsHijack: "default" }))).toBe("hijack");
    expect(resolveDnsMode(setting({}))).toBe("hijack");
  });

  // an unrecognised value must never read as interception; configure.ResolveDnsMode
  // refuses one on a write and resolves one to off on a read
  test("reads a mode this build does not know as off", () => {
    expect(resolveDnsMode(setting({ dnsMode: "hijackk" as never }))).toBe(
      "off",
    );
  });
});

describe("saving the DNS dialog", () => {
  const rules = [{ server: "1.1.1.1", domains: "", outbound: "direct" }];
  const input = {
    mode: "service" as const,
    storedMode: "hijack" as const,
    rules,
    rulesDirty: true,
    nodeDns: "udp://1.1.1.1:53",
    storedNodeDns: "auto",
  };
  beforeEach(() => api.putDnsRules.mockReset().mockResolvedValue({}));

  test("saves rules, mode and the new node source in one request", async () => {
    expect(await saveDnsSettings(input)).toEqual({ status: "saved" });
    expect(api.putDnsRules).toHaveBeenCalledExactlyOnceWith({
      rules,
      dnsMode: "service",
      nodeDns: input.nodeDns,
    });
  });
  test("writes nothing when nothing changed", async () => {
    await saveDnsSettings({
      ...input,
      storedMode: "service",
      rulesDirty: false,
      storedNodeDns: input.nodeDns,
    });
    expect(api.putDnsRules).not.toHaveBeenCalled();
  });
  test("keeps hidden rules and node edits out of an off-mode save", async () => {
    await saveDnsSettings({ ...input, mode: "off" });
    expect(api.putDnsRules).toHaveBeenCalledExactlyOnceWith({ dnsMode: "off" });
  });
  test("saves only the node source if that is the only edit", async () => {
    await saveDnsSettings({
      ...input,
      storedMode: "service",
      rulesDirty: false,
    });
    expect(api.putDnsRules).toHaveBeenCalledExactlyOnceWith({
      nodeDns: input.nodeDns,
    });
  });
  test("reports a refused combined save without partial success", async () => {
    api.putDnsRules.mockRejectedValueOnce(new Error("bad upstream"));
    expect(await saveDnsSettings(input)).toEqual({
      status: "failed",
      error: "bad upstream",
    });
  });
});
