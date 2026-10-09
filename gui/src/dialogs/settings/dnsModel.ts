// What the DNS dialog owns: the mode, which decides both whether the module
// runs and whether the system's queries are pointed at it, and the rules the
// module follows. The two are saved together but through two requests the
// service applies separately, so saveDnsSettings reports which of them
// reached the store instead of pretending the pair is one write.
import { putDnsRules, putSetting } from "@/api";
import { errorText } from "@/api/errors";
import type { DnsMode, DnsRule, Setting } from "@/api/types";

export const dnsModeItems = (t: (key: string) => string) => [
  { value: "off", title: t("dns.modeOff") },
  { value: "service", title: t("dns.modeService") },
  { value: "hijack", title: t("dns.modeHijack") },
];

/** what each mode means, shown under the choice */
export const dnsModeHelp = (t: (key: string) => string) => ({
  off: t("dns.modeOffHelp"),
  service: t("dns.modeServiceHelp"),
  hijack: t("dns.modeHijackHelp"),
});

/** The mode a service stores, mirroring configure.ResolveDnsMode: a mode it does
 *  not know is read as off rather than silently turning interception on, and a
 *  service that predates the mode answers through the opt-out it does have. */
export function resolveDnsMode(setting: Setting | undefined): DnsMode {
  const mode = setting?.dnsMode;
  if (mode === "off" || mode === "service" || mode === "hijack") return mode;
  if (mode) return "off";
  return setting?.dnsHijack === "no" ? "off" : "hijack";
}

export const defaultDnsRules: DnsRule[] = [
  { server: "localhost", domains: "geosite:private", outbound: "direct" },
  { server: "223.5.5.5", domains: "geosite:cn", outbound: "direct" },
  { server: "8.8.8.8", domains: "", outbound: "proxy" },
];

export const cloneDnsRules = (rules: DnsRule[]): DnsRule[] =>
  rules.map((rule) => ({ ...rule }));

/** The form shape. Fields set outside the dialog (the DNS module's matchers)
 *  survive a save; `upstream` and `domain` are the migrated aliases of the two
 *  edited fields and the generator prefers them, so they fold into the form and
 *  are not sent back. */
export function normalizeDnsRules(
  stored: Partial<DnsRule>[] | null | undefined,
): DnsRule[] {
  if (!stored?.length) return cloneDnsRules(defaultDnsRules);
  return stored.map(({ upstream, domain, ...rule }) => ({
    ...rule,
    server: (upstream as string) || rule.server || "",
    domains: (domain as string) || rule.domains || "",
    outbound: rule.outbound || "direct",
  }));
}

/** a rule without a server is a row the user has not filled in yet */
export const dnsRulesPayload = (rules: DnsRule[]): DnsRule[] =>
  rules.filter((rule) => rule.server.trim() !== "");

export type DnsSaveResult =
  /** both writes the dialog owed succeeded; the flags say which those were */
  | { status: "saved"; rules: boolean; mode: boolean }
  /** the rules are stored, the mode was refused: the retry owes only the mode */
  | { status: "partial"; error: string }
  /** nothing was stored; the form stays as it was */
  | { status: "failed"; error: string };

/** The rules go first and the mode last: the write that makes DNS active is the
 *  one that must not be left standing over a rejected one. A service that
 *  refuses the mode has then already stored the rules, which the caller records
 *  so the retry does not send them again. */
export async function saveDnsSettings(input: {
  mode: DnsMode;
  storedMode: DnsMode;
  rules: DnsRule[];
  rulesDirty: boolean;
}): Promise<DnsSaveResult> {
  const { mode, storedMode, rules, rulesDirty } = input;
  // An off mode keeps the rules: nothing would read them, and dropping them
  // would make turning the module back on a different configuration.
  const writeRules = mode !== "off" && rulesDirty;
  if (writeRules) {
    try {
      await putDnsRules(dnsRulesPayload(rules));
    } catch (err) {
      return { status: "failed", error: errorText(err) };
    }
  }
  if (mode !== storedMode) {
    try {
      await putSetting({ dnsMode: mode });
    } catch (err) {
      return writeRules
        ? { status: "partial", error: errorText(err) }
        : { status: "failed", error: errorText(err) };
    }
  }
  return { status: "saved", rules: writeRules, mode: mode !== storedMode };
}
