// The editor's form models, one per protocol, with the field sets the
// old editor had (the codec reads and writes exactly these keys). vmess
// and vless share the v2ray model; the tab sets its protocol.

export const v2rayModel = () => ({
  ps: "",
  add: "",
  port: "",
  id: "",
  flow: "",
  aid: "",
  net: "tcp",
  type: "none",
  host: "",
  path: "",
  tls: "none",
  sni: "",
  quicSecurity: "none",
  fp: "",
  pbk: "",
  sid: "",
  spx: "",
  alpn: "",
  scy: "auto",
  v: "",
  pinnedPeerCertSha256: "",
  verifyPeerCertByName: "",
  protocol: "vmess",
  key: "none",
  xhttpMode: "auto",
  xhttpHeaders: [] as { key: string; value: string }[],
  noGRPCHeader: false,
  noSSEHeader: false,
  uplinkHTTPMethod: "",
  scMaxEachPostBytesFrom: "",
  scMaxEachPostBytesTo: "",
  scMinPostsIntervalFrom: "",
  scMinPostsIntervalTo: "",
  scMaxBufferedPosts: "",
  scStreamUpServerFrom: "",
  scStreamUpServerTo: "",
  xPaddingBytesFrom: "",
  xPaddingBytesTo: "",
  xmuxMaxConcurFrom: "",
  xmuxMaxConcurTo: "",
  xmuxMaxConnFrom: "",
  xmuxMaxConnTo: "",
  xmuxCMaxReuseFrom: "",
  xmuxCMaxReuseTo: "",
  xmuxHMaxReqFrom: "",
  xmuxHMaxReqTo: "",
  xmuxHMaxReusableFrom: "",
  xmuxHMaxReusableTo: "",
  xmuxHKeepAlive: "",
  maxEarlyData: "",
  earlyDataHeaderName: "",
  multiMode: false,
  idleTimeout: "",
  healthCheckTimeout: "",
  permitWithoutStream: false,
  initialWindowsSize: "",
});
export type V2rayModel = ReturnType<typeof v2rayModel>;

export const ssModel = () => ({
  method: "2022-blake3-aes-128-gcm",
  plugin: "",
  obfs: "http",
  tls: "",
  path: "/",
  mode: "websocket",
  host: "",
  password: "",
  server: "",
  port: "",
  name: "",
  protocol: "ss",
  impl: "",
  backend: "",
});
export type SsModel = ReturnType<typeof ssModel>;

export const ssrModel = () => ({
  method: "aes-128-cfb",
  password: "",
  server: "",
  port: "",
  name: "",
  proto: "origin",
  protoParam: "",
  obfs: "plain",
  obfsParam: "",
  protocol: "ssr",
});
export type SsrModel = ReturnType<typeof ssrModel>;

export const trojanModel = () => ({
  name: "",
  server: "",
  peer: "", // tls sni
  host: "", // websocket host
  path: "", // websocket path
  pinnedPeerCertSha256: "",
  verifyPeerCertByName: "",
  port: "",
  password: "",
  method: "origin", // shadowsocks
  ssCipher: "aes-128-gcm",
  ssPassword: "",
  net: "tcp",
  obfs: "none", // websocket
  protocol: "trojan",
  backend: "",
});
export type TrojanModel = ReturnType<typeof trojanModel>;

export const juicityModel = () => ({
  name: "",
  server: "",
  port: "",
  sni: "",
  cc: "bbr",
  uuid: "",
  password: "",
  pinnedCertchainSha256: "",
  allowInsecure: false,
  protocol: "juicity",
});
export type JuicityModel = ReturnType<typeof juicityModel>;

export const tuicModel = () => ({
  name: "",
  server: "",
  port: "",
  sni: "",
  cc: "bbr",
  uuid: "",
  password: "",
  pinnedPeerCertSha256: "",
  verifyPeerCertByName: "",
  allowInsecure: false,
  disableSni: false,
  alpn: "h3",
  udpRelayMode: "native",
  protocol: "tuic",
});
export type TuicModel = ReturnType<typeof tuicModel>;

export const hysteria2Model = () => ({
  name: "",
  server: "",
  port: "",
  password: "",
  sni: "",
  obfs: "none",
  obfsPassword: "",
  pinnedPeerCertSha256: "",
  verifyPeerCertByName: "",
  protocol: "hysteria2",
});
export type Hysteria2Model = ReturnType<typeof hysteria2Model>;

export const httpModel = () => ({
  username: "",
  password: "",
  host: "",
  port: "",
  protocol: "http",
  name: "",
});
export type HttpModel = ReturnType<typeof httpModel>;

export const socks5Model = () => ({
  username: "",
  password: "",
  host: "",
  port: "",
  protocol: "socks5",
  name: "",
});
export type Socks5Model = ReturnType<typeof socks5Model>;

export const anytlsModel = () => ({
  name: "",
  host: "",
  port: "",
  auth: "",
  sni: "",
  pinnedPeerCertSha256: "",
  verifyPeerCertByName: "",
  allowInsecure: false,
  minIdleSession: "",
  protocol: "anytls",
});
export type AnytlsModel = ReturnType<typeof anytlsModel>;

export const wireguardModel = () => ({
  protocol: "wireguard",
  name: "",
  address: "",
  port: "",
  publicKey: "",
  privateKey: "",
  localAddress: "",
  dns: "",
  mtu: "",
  allowedIPs: "",
  persistentKeepalive: "",
  preSharedKey: "",
  endpoint: "",
});
export type WireguardModel = ReturnType<typeof wireguardModel>;

/** defaultModels gives a fresh set, the editor's state for one dialog. */
export function defaultModels() {
  return {
    v2ray: v2rayModel(),
    ss: ssModel(),
    ssr: ssrModel(),
    trojan: trojanModel(),
    juicity: juicityModel(),
    tuic: tuicModel(),
    hysteria2: hysteria2Model(),
    http: httpModel(),
    socks5: socks5Model(),
    anytls: anytlsModel(),
    wireguard: wireguardModel(),
  };
}
export type Models = ReturnType<typeof defaultModels>;
