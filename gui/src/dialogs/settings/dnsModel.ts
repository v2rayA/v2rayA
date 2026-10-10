import { putDnsRules } from "@/api";
import { errorText } from "@/api/errors";
import type {
  DnsMode,
  DnsRule,
  DnsSettingsRequest,
  Setting,
} from "@/api/types";

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
  { status: "saved" } | { status: "failed"; error: string };

export async function saveDnsSettings(input: {
  mode: DnsMode;
  storedMode: DnsMode;
  rules: DnsRule[];
  rulesDirty: boolean;
  nodeDns: string;
  storedNodeDns: string;
}): Promise<DnsSaveResult> {
  const { mode, storedMode, rules, rulesDirty, nodeDns, storedNodeDns } = input;
  const body: DnsSettingsRequest = {};
  if (mode !== storedMode) body.dnsMode = mode;
  // Hidden edits do not replace the configuration an off mode keeps.
  if (mode !== "off") {
    if (rulesDirty) body.rules = dnsRulesPayload(rules);
    if (nodeDns !== storedNodeDns) body.nodeDns = nodeDns;
  }
  try {
    if (Object.keys(body).length) await putDnsRules(body);
    return { status: "saved" };
  } catch (err) {
    return { status: "failed", error: errorText(err) };
  }
}
