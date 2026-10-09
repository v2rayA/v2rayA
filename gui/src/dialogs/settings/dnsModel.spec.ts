// @vitest-environment happy-dom
import { beforeEach, describe, expect, test, vi } from "vitest";
import type { Setting } from "@/api/types";

const api = vi.hoisted(() => ({ putDnsRules: vi.fn(), putSetting: vi.fn() }));
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

  beforeEach(() => {
    api.putDnsRules.mockReset().mockResolvedValue({});
    api.putSetting.mockReset().mockResolvedValue({});
  });

  test("writes only what changed", async () => {
    expect(
      await saveDnsSettings({
        mode: "hijack",
        storedMode: "hijack",
        rules,
        rulesDirty: false,
      }),
    ).toEqual({ status: "saved", rules: false, mode: false });
    expect(api.putDnsRules).not.toHaveBeenCalled();
    expect(api.putSetting).not.toHaveBeenCalled();
  });

  test("stores the rules before the mode that activates them", async () => {
    const order: string[] = [];
    api.putDnsRules.mockImplementation(async () => void order.push("rules"));
    api.putSetting.mockImplementation(async () => void order.push("mode"));
    expect(
      await saveDnsSettings({
        mode: "service",
        storedMode: "hijack",
        rules,
        rulesDirty: true,
      }),
    ).toEqual({ status: "saved", rules: true, mode: true });
    expect(api.putSetting).toHaveBeenCalledExactlyOnceWith({
      dnsMode: "service",
    });
    expect(order).toEqual(["rules", "mode"]);
  });

  // dropping them would make switching the mode back on a different
  // configuration, and nothing would have read them in the meantime
  test("keeps the rules an off mode puts out of reach", async () => {
    expect(
      await saveDnsSettings({
        mode: "off",
        storedMode: "hijack",
        rules,
        rulesDirty: true,
      }),
    ).toEqual({ status: "saved", rules: false, mode: true });
    expect(api.putDnsRules).not.toHaveBeenCalled();
    expect(api.putSetting).toHaveBeenCalledExactlyOnceWith({ dnsMode: "off" });
  });

  test("stores nothing and keeps the form as it was when the rules are refused", async () => {
    api.putDnsRules.mockRejectedValueOnce(new Error("bad upstream"));
    const result = await saveDnsSettings({
      mode: "service",
      storedMode: "hijack",
      rules,
      rulesDirty: true,
    });
    expect(result.status).toBe("failed");
    expect(api.putSetting).not.toHaveBeenCalled();
  });

  test("reports the half that landed so the retry owes only the mode", async () => {
    api.putSetting.mockRejectedValueOnce(new Error("mode refused"));
    const result = await saveDnsSettings({
      mode: "service",
      storedMode: "hijack",
      rules,
      rulesDirty: true,
    });
    expect(result).toEqual({ status: "partial", error: "mode refused" });
  });

  test("reports a refused mode on its own as nothing stored", async () => {
    api.putSetting.mockRejectedValueOnce(new Error("mode refused"));
    expect(
      await saveDnsSettings({
        mode: "off",
        storedMode: "hijack",
        rules,
        rulesDirty: false,
      }),
    ).toEqual({ status: "failed", error: "mode refused" });
  });
});
