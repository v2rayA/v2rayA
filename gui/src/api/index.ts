// One function per backend operation (method + path), 40 in all, named
// <method><Path>. Bodies and queries are what the old components sent, so
// the recorded requests match byte for byte; types narrow them where the
// backend's Go types are known (types.ts) and stay open elsewhere.
import { apiRoot, call, client, timeouts } from "./client";

/** What a caller may attach to a request: the connect watcher aborts through `signal`, its poll shortens `timeout`. */
export interface RequestOptions {
  signal?: AbortSignal;
  timeout?: number;
}
import type {
  CustomInbound,
  DnsRule,
  DnsRulesResponse,
  Ports,
  Setting,
  SettingResponse,
  TouchResponse,
  VersionResponse,
  Which,
} from "./types";

// ---- account and session ----------------------------------------------
export const getAccount = () =>
  call<{ hasAnyAccounts: boolean }>({ url: "account", method: "get" });
export const postAccount = (body: { username: string; password: string }) =>
  call<{ token: string }>({ url: "account", method: "post", data: body });
export const postLogin = (body: { username: string; password: string }) =>
  call<{ token: string }>({ url: "login", method: "post", data: body });
export const getVersion = () =>
  call<VersionResponse>({ url: "version", method: "get" });

// ---- nodes and subscriptions ---------------------------------------------
export const getTouch = (o: RequestOptions = {}) =>
  call<TouchResponse>({ url: "touch", method: "get", ...o });
export const deleteTouch = (touches: Which[]) =>
  call<TouchResponse>({ url: "touch", method: "delete", data: { touches } });
// The editor sends `which: null` for a new node, as the old page did; the
// key stays in the body. Saving one node has no time limit, a batch has.
export const postImport = (
  body: { url: string; kind?: string; which?: Which | null },
  timeout: number = timeouts.import,
) =>
  call<TouchResponse>({ url: "import", method: "post", data: body, timeout });
export const getSharingAddress = (touch: Which) =>
  call<{ sharingAddress: string }>({
    url: "sharingAddress",
    method: "get",
    params: { touch },
  });
export const putSubscription = (which: Which) =>
  call<TouchResponse>({ url: "subscription", method: "put", data: which });
export const patchSubscription = (body: Record<string, unknown>) =>
  call<TouchResponse>({ url: "subscription", method: "patch", data: body });
export const getPingLatency = (whiches: Which[]) =>
  call<{ whiches: Which[] }>({
    url: "pingLatency",
    method: "get",
    params: { whiches },
    timeout: timeouts.none,
  });
export const getHttpLatency = (whiches: Which[]) =>
  call<{ whiches: Which[] }>({
    url: "httpLatency",
    method: "get",
    params: { whiches },
    timeout: timeouts.none,
  });

// ---- connections and the core ---------------------------------------------
export const putOutboundConnections = (body: {
  outbound: string;
  touches: Which[];
}) =>
  call<TouchResponse>({
    url: "outboundConnections",
    method: "put",
    data: body,
  });
/** the member a group routes through alone; null returns it to balancing */
export const putOutboundSelection = (body: {
  outbound: string;
  which: Which | null;
}) =>
  call<TouchResponse>({
    url: "outboundSelection",
    method: "put",
    data: body,
  });
export const postV2ray = (o: RequestOptions = {}) =>
  call<TouchResponse>({ url: "v2ray", method: "post", ...o });
export const deleteV2ray = () =>
  call<TouchResponse>({ url: "v2ray", method: "delete" });

// ---- outbound groups ---------------------------------------------------------
export const getOutbounds = () =>
  call<{ outbounds: string[] }>({ url: "outbounds", method: "get" });
export interface OutboundSetting {
  probeURL: string;
  probeInterval: string;
  type: string;
  /** the share link of the member routed through alone; empty balances */
  selected?: string;
}
export const getOutbound = (outbound: string) =>
  call<{ setting: OutboundSetting }>({
    url: "outbound",
    method: "get",
    params: { outbound },
  });
export const postOutbound = (
  body: Record<string, unknown>,
  o: RequestOptions = {},
) => call<unknown>({ url: "outbound", method: "post", data: body, ...o });
export const putOutbound = (body: {
  outbound: string;
  setting: OutboundSetting;
}) => call<unknown>({ url: "outbound", method: "put", data: body });
export const deleteOutbound = (
  body: Record<string, unknown>,
  o: RequestOptions = {},
) => call<unknown>({ url: "outbound", method: "delete", data: body, ...o });

// ---- settings ------------------------------------------------------------------
export const getSetting = () =>
  call<SettingResponse>({ url: "setting", method: "get" });
export const putSetting = (setting: Setting, o: RequestOptions = {}) =>
  call<unknown>({ url: "setting", method: "put", data: setting, ...o });
export const getPorts = () => call<Ports>({ url: "ports", method: "get" });
export const putPorts = (ports: Ports) =>
  call<unknown>({ url: "ports", method: "put", data: ports });
export const getCustomInbound = () =>
  call<{ inbounds: CustomInbound[] }>({ url: "customInbound", method: "get" });
export const postCustomInbound = (body: CustomInbound) =>
  call<unknown>({ url: "customInbound", method: "post", data: body });
export const deleteCustomInbound = (body: { tag: string }) =>
  call<unknown>({ url: "customInbound", method: "delete", data: body });
export const getRoutingA = () =>
  call<{ routingA: string }>({ url: "routingA", method: "get" });
export const putRoutingA = (body: { routingA: string }) =>
  call<unknown>({ url: "routingA", method: "put", data: body });
export const getDnsRules = () =>
  call<DnsRulesResponse>({ url: "dnsRules", method: "get" });
export const putDnsRules = (body: DnsRule[]) =>
  call<unknown>({ url: "dnsRules", method: "put", data: body });
export const getDomainsExcluded = () =>
  call<{ domains: string }>({ url: "domainsExcluded", method: "get" });
export const putDomainsExcluded = (body: { domains: string }) =>
  call<unknown>({ url: "domainsExcluded", method: "put", data: body });
export const getTproxyWhiteIpGroups = () =>
  call<{ countryCodes: string[]; customIps: string[] }>({
    url: "tproxyWhiteIpGroups",
    method: "get",
  });
export const putTproxyWhiteIpGroups = (body: {
  countryCodes: string[];
  customIps: string[];
}) => call<unknown>({ url: "tproxyWhiteIpGroups", method: "put", data: body });
export const getRemoteGFWListVersion = () =>
  call<{ remoteGFWListVersion: string }>({
    url: "remoteGFWListVersion",
    method: "get",
  });
export const putGfwList = (body: { downloadLink: string }) =>
  call<{ localGFWListVersion: string; alreadyUpToDate?: boolean }>({
    url: "gfwList",
    method: "put",
    data: body,
    timeout: timeouts.none,
  });
export const deleteGfwList = () =>
  call<unknown>({ url: "gfwList", method: "delete" });

// ---- logs ------------------------------------------------------------------------
/** GET /logger answers plain text from the byte offset `skip`, not the envelope. */
export const getLogger = async (params: { skip: number }): Promise<string> => {
  const res = await client.get<string>(`${apiRoot()}/logger`, {
    params,
    responseType: "text",
    transformResponse: (data) => data,
  });
  return typeof res.data === "string" ? res.data : "";
};
