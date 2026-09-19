// Shapes the backend sends and accepts, named after the Go types they come
// from (service/kernel/touch, service/db/configure). Fields the redo does
// not read are left out on purpose; a widening is a one-line change here.

export type TouchType = "server" | "subscription" | "subscriptionServer";

/** configure.Which — how the backend addresses one node or subscription. */
export interface Which {
  _type: TouchType;
  id: number;
  /** subscription index, only for subscriptionServer */
  sub?: number;
  pingLatency?: string;
  Link?: string;
  outbound?: string;
  /** in a touch's connectedServer: the member its group routes through alone */
  selected?: boolean;
}

/** touch.Server; `sub` and `connected` are the page's own marks on a row */
export interface TouchServer {
  id: number;
  _type: TouchType;
  name: string;
  address: string;
  net: string;
  pingLatency: string;
  sub?: number;
  connected?: boolean;
}

/** touch.Subscription */
export interface TouchSubscription {
  id: number;
  _type: TouchType;
  remarks?: string;
  host: string;
  address: string;
  status: string;
  info: string;
  servers: TouchServer[];
  autoSelect: boolean;
}

/** touch.Touch */
export interface Touch {
  servers: TouchServer[];
  subscriptions: TouchSubscription[];
  connectedServer: Which[] | null;
}

/** GET /touch */
export interface TouchResponse {
  running: boolean;
  networkPaused: boolean;
  touch: Touch;
}

/** GET /version */
export interface VersionResponse {
  version: string;
  foundNew: boolean;
  remoteVersion: string;
  serviceValid: boolean;
  v5: boolean;
  /** 0 or 1: the backend sends its lite flag as a number */
  lite: number;
  loadBalanceValid: boolean;
  variant: string;
  /** the core binary's version; empty when it could not be asked */
  coreVersion?: string;
  os: string;
  isRoot: boolean;
  tunSupported: boolean;
  coreVersionValid: boolean;
  coreVersionErr: string;
  hasAccounts: boolean;
  lastKernelExit: unknown;
  docker?: boolean;
}

/** configure.Setting — kept as the backend's own keys; the settings dialog owns the full list. */
export type Setting = Record<string, unknown> & {
  transparent: string;
  transparentType: string;
  pacMode: string;
  logLevel: string;
};

/** GET /setting */
export interface SettingResponse {
  setting: Setting;
  localGFWListVersion: string;
}

/** configure.Ports */
export interface Ports {
  socks5: number;
  http: number;
  socks5WithPac: number;
  httpWithPac: number;
  vmess: number;
  vmessLink?: string;
  api: { port: number; services: string[] };
}

/** configure.CustomInbound */
export interface CustomInbound {
  tag: string;
  protocol: "socks" | "http";
  port: number;
  outbound: string;
  outboundType: "direct" | "routingA" | string;
  routingARules: string;
  username: string;
  password: string;
}

export interface DnsRule {
  server: string;
  domains: string;
  outbound: string;
  // matchers the DNS module reads and only the API sets; kept on save
  [extra: string]: unknown;
}

export interface DnsRulesResponse {
  rules?: Partial<DnsRule>[] | null;
}

/** The WebSocket frames on /api/message: the ones the interface reads, and any other the backend may add. */
export interface RunningStateMessage {
  type: "running_state";
  body: { running: boolean; networkPaused?: boolean };
}
/** kernel/v2ray OutboundStatus: what the core's observatory saw of one connected node */
export interface OutboundStatus {
  alive: boolean;
  delay: number;
  outbound_tag: string;
  which: Which;
  last_seen_time: number;
  last_try_time: number;
}
export interface ObservatoryMessage {
  type: "observatory";
  body: { outboundName: string; outboundStatus: OutboundStatus[] };
}
export interface TrafficMessage {
  type: "traffic";
  body: {
    up: number;
    down: number;
    upTotal: number;
    downTotal: number;
  };
}
export type WsMessage = { type: string; body?: unknown };
