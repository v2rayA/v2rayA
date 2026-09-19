/* eslint-disable @typescript-eslint/no-explicit-any */
// Share-link codec: a share link in, the editor's form object out, and
// back. Moved from modalServer.vue without changes in logic; the form
// objects keep the shapes the protocol forms bind to. The editor is the
// only consumer: the import dialog posts links as typed, and the share
// dialog shows the backend's link.
import { Base64 } from "js-base64";
import { decodeSafe, generateURL as buildURL, parseURL } from "@/lib/url";

export type ShareForm = Record<string, any>;

/** parseShareLink turns a share link into the editor's form object, or null for an unknown scheme. */
export function parseShareLink(url: string): ShareForm | null {
  if (url.toLowerCase().startsWith("vmess://")) {
    const obj = JSON.parse(
      Base64.decode(url.substring(url.indexOf("://") + 3)),
    );
    // the backend stores ps as typed; only decode when it is valid
    // percent-encoding, a literal "%" in a name must survive
    obj.ps = decodeSafe(obj.ps);
    // the backend keys the uTLS fingerprint "fingerprint" in vmess JSON
    obj.fp = obj.fp || obj.fingerprint || "";
    obj.tls = obj.tls || "none";
    obj.type = obj.type || "none";
    obj.scy = obj.scy || "auto";
    obj.protocol = obj.protocol || "vmess";
    // xhttpHeaders arrives as a JSON-encoded string; convert it to the
    // array model used by the GUI so the editor round-trips correctly.
    if (typeof obj.xhttpHeaders === "string") {
      try {
        const hdrsObj = JSON.parse(obj.xhttpHeaders);
        obj.xhttpHeaders = Object.entries(hdrsObj || {}).map(
          ([key, value]) => ({ key, value }),
        );
      } catch {
        obj.xhttpHeaders = [];
      }
    }
    if (!Array.isArray(obj.xhttpHeaders)) obj.xhttpHeaders = [];
    // multiMode/permitWithoutStream arrive as strings; restore the booleans
    // used by the GUI model.
    if (typeof obj.multiMode === "string")
      obj.multiMode = obj.multiMode === "true" || obj.multiMode === "1";
    if (typeof obj.permitWithoutStream === "string")
      obj.permitWithoutStream =
        obj.permitWithoutStream === "true" || obj.permitWithoutStream === "1";
    return obj;
  } else if (url.toLowerCase().startsWith("vless://")) {
    const u = parseURL(url);
    const o: ShareForm = {
      ps: decodeSafe(u.hash),
      add: u.host,
      port: u.port,
      id: decodeSafe(u.username),
      flow: u.params.flow || "",
      net: u.params.type || "tcp",
      type: u.params.headerType || "none",
      host: u.params.host || "",
      path: u.params.path || u.params.serviceName || "",
      alpn: u.params.alpn || "",
      sni: u.params.sni || "",
      tls: u.params.security === "xtls" ? "tls" : u.params.security || "none",
      quicSecurity: u.params.quicSecurity || "none",
      fp: u.params.fp || "",
      pbk: u.params.pbk || "",
      sid: u.params.sid || "",
      spx: u.params.spx || "",
      pinnedPeerCertSha256: u.params.pinnedPeerCertSha256 || "",
      verifyPeerCertByName: u.params.verifyPeerCertByName || "",
      key: u.params.key,
      xhttpMode: u.params.xhttpMode || "auto",
      noGRPCHeader: u.params.noGRPCHeader === "true",
      noSSEHeader: u.params.noSSEHeader === "true",
      uplinkHTTPMethod: u.params.uplinkHTTPMethod || "",
      scMaxEachPostBytesFrom: u.params.scMaxEachPostBytesFrom || "",
      scMaxEachPostBytesTo: u.params.scMaxEachPostBytesTo || "",
      scMinPostsIntervalFrom: u.params.scMinPostsIntervalFrom || "",
      scMinPostsIntervalTo: u.params.scMinPostsIntervalTo || "",
      scMaxBufferedPosts: u.params.scMaxBufferedPosts || "",
      scStreamUpServerFrom: u.params.scStreamUpServerFrom || "",
      scStreamUpServerTo: u.params.scStreamUpServerTo || "",
      xPaddingBytesFrom: u.params.xPaddingBytesFrom || "",
      xPaddingBytesTo: u.params.xPaddingBytesTo || "",
      xmuxMaxConcurFrom: u.params.xmuxMaxConcurFrom || "",
      xmuxMaxConcurTo: u.params.xmuxMaxConcurTo || "",
      xmuxMaxConnFrom: u.params.xmuxMaxConnFrom || "",
      xmuxMaxConnTo: u.params.xmuxMaxConnTo || "",
      xmuxCMaxReuseFrom: u.params.xmuxCMaxReuseFrom || "",
      xmuxCMaxReuseTo: u.params.xmuxCMaxReuseTo || "",
      xmuxHMaxReqFrom: u.params.xmuxHMaxReqFrom || "",
      xmuxHMaxReqTo: u.params.xmuxHMaxReqTo || "",
      xmuxHMaxReusableFrom: u.params.xmuxHMaxReusableFrom || "",
      xmuxHMaxReusableTo: u.params.xmuxHMaxReusableTo || "",
      xmuxHKeepAlive: u.params.xmuxHKeepAlive || "",
      xhttpHeaders: (() => {
        try {
          const raw = u.params.xhttpHeaders;
          if (!raw) return [];
          const obj = JSON.parse(raw);
          return Object.entries(obj).map(([key, value]) => ({ key, value }));
        } catch {
          return [];
        }
      })(),
      maxEarlyData: u.params.maxEarlyData || "",
      earlyDataHeaderName: u.params.earlyDataHeaderName || "",
      multiMode: u.params.multiMode === "true" || u.params.multiMode === "1",
      idleTimeout: u.params.idleTimeout || "",
      healthCheckTimeout: u.params.healthCheckTimeout || "",
      permitWithoutStream:
        u.params.permitWithoutStream === "true" ||
        u.params.permitWithoutStream === "1",
      initialWindowsSize: u.params.initialWindowsSize || "",
      protocol: "vless",
    };
    if (o.alpn !== "") {
      o.alpn = decodeSafe(o.alpn);
    }
    if (o.net === "mkcp" || o.net === "kcp") {
      o.path = u.params.seed;
    }
    return o;
  } else if (url.toLowerCase().startsWith("ss://")) {
    const u = parseURL(url);
    let userinfo = u.username;
    // The backend percent-escapes the base64 userinfo ("=" becomes %3D),
    // and js-base64 decodes those escapes as data, which corrupted the
    // password of every shadowsocks node opened for editing.
    try {
      userinfo = decodeSafe(userinfo);
    } catch {
      // a literal "%" in the userinfo: use it as it came
    }
    // Handle SIP002 format: ss://BASE64URL@host:port vs legacy ss://BASE64
    let method = "",
      password = "";
    try {
      const decoded = Base64.decode(userinfo);
      const idx = decoded.indexOf(":");
      if (idx > -1) {
        method = decoded.substring(0, idx);
        password = decoded.substring(idx + 1);
      }
    } catch {
      method = userinfo;
    }
    // parseURL already decoded the query values; decoding again turns
    // %2F into a path separator and throws on a literal percent sign
    const ssPlugin = u.params.plugin || "";
    const opts = ssPlugin.split(";");
    const o: ShareForm = {
      method: method,
      password: password,
      server: u.host,
      port: u.port,
      name: decodeSafe(u.hash || ""),
      plugin: opts[0] || "",
      obfs: "http",
      tls: "",
      mode: "websocket",
      host: "",
      path: "",
      impl: "",
      protocol: "ss",
      backend: u.params["v2raya-backend"] || "",
    };
    switch (o.plugin) {
      case "obfs-local":
      case "simpleobfs":
        o.plugin = "simple-obfs";
        break;
    }
    // "obfs-local;obfs=tls;obfs-host=example.com" or
    // "v2ray-plugin;tls;mode=websocket;host=example.com;path=/ws"
    for (const opt of opts.slice(1)) {
      const [k, v = ""] = opt.split("=");
      switch (k) {
        case "obfs":
          o.obfs = v;
          break;
        case "host":
        case "obfs-host":
          o.host = v;
          break;
        case "path":
        case "obfs-path":
        case "obfs-uri":
          o.path = v;
          break;
        case "mode":
          o.mode = v;
          break;
        case "tls":
          o.tls = "tls";
          break;
        case "impl":
          o.impl = v;
          break;
      }
    }
    return o;
  } else if (url.toLowerCase().startsWith("ssr://")) {
    url = Base64.decode(url.substr(6));
    const arr = url.split("/?");
    const query = (arr[1] ?? "").split("&");
    const m: ShareForm = {};
    for (const param of query) {
      const [key, val] = param.split("=", 2);
      m[key] = Base64.decode(val);
    }
    let pre = arr[0].split(":");
    if (pre.length > 6) {
      //如果长度多于6，说明host中包含字符:，重新合并前几个分组到host去
      pre[pre.length - 6] = pre.slice(0, pre.length - 5).join(":");
      pre = pre.slice(pre.length - 6);
    }
    pre[5] = Base64.decode(pre[5]);
    return {
      method: pre[3],
      password: pre[5],
      server: pre[0],
      port: pre[1],
      name: m["remarks"],
      proto: pre[2],
      protoParam: m["protoparam"],
      obfs: pre[4],
      obfsParam: m["obfsparam"],
      protocol: "ssr",
    };
  } else if (
    url.toLowerCase().startsWith("trojan://") ||
    url.toLowerCase().startsWith("trojan-go://")
  ) {
    const u = parseURL(url);
    const o: ShareForm = {
      password: decodeSafe(u.username),
      server: u.host,
      port: u.port,
      name: decodeSafe(u.hash),
      peer: u.params.peer || u.params.sni || "",
      pinnedPeerCertSha256: u.params.pinnedPeerCertSha256 || "",
      verifyPeerCertByName: u.params.verifyPeerCertByName || "",
      method: "origin",
      net: u.params.type || "tcp",
      host: u.params.host || "",
      obfs: "none",
      ssCipher: "2022-blake3-aes-128-gcm",
      path: u.params.path || u.params.serviceName || "",
      protocol: "trojan",
      backend: u.params["v2raya-backend"] || "",
    };
    if (url.toLowerCase().startsWith("trojan-go://")) {
      if (u.params.encryption?.startsWith("ss;")) {
        o.method = "shadowsocks";
        const fields = u.params.encryption.split(";");
        o.ssCipher = fields[1];
        o.ssPassword = fields[2];
      }
      if (u.params.type === "ws") {
        o.obfs = "websocket";
        o.host = u.params.host || "";
        o.path = u.params.path || "/";
      }
    }
    return o;
  } else if (url.toLowerCase().startsWith("juicity://")) {
    const u = parseURL(url);
    return {
      name: decodeSafe(u.hash),
      uuid: decodeSafe(u.username),
      password: decodeSafe(u.password),
      server: u.host,
      port: u.port,
      sni: u.params.sni || "",
      pinnedCertchainSha256: u.params.pinned_certchain_sha256 || "",
      cc: u.params.congestion_control || "bbr",
      allowInsecure:
        u.params.allow_insecure === "true" || u.params.allowInsecure === "true",
      protocol: "juicity",
    };
  } else if (url.toLowerCase().startsWith("tuic://")) {
    const u = parseURL(url);
    return {
      name: decodeSafe(u.hash),
      uuid: decodeSafe(u.username),
      password: decodeSafe(u.password),
      server: u.host,
      port: u.port,
      sni: u.params.sni || "",
      pinnedPeerCertSha256:
        u.params.pinnedPeerCertSha256 || u.params.pinned_peer_cert_sha256 || "",
      verifyPeerCertByName:
        u.params.verifyPeerCertByName ||
        u.params.verify_peer_cert_by_name ||
        "",
      disableSni:
        u.params.disable_sni === "true" || u.params.disable_sni === "1",
      alpn: u.params.alpn,
      cc: u.params.congestion_control || "bbr",
      udpRelayMode: u.params.udp_relay_mode || "native",
      allowInsecure:
        u.params.allow_insecure === "true" || u.params.allow_insecure === "1",
      protocol: "tuic",
    };
  } else if (
    url.toLowerCase().startsWith("hysteria2://") ||
    url.toLowerCase().startsWith("hy2://")
  ) {
    const u = parseURL(url);
    let password = decodeSafe(u.username);
    const userInfoEnd = u.source.indexOf("@");
    if (
      userInfoEnd !== -1 &&
      u.source.slice(u.source.indexOf("://") + 3, userInfoEnd).includes(":")
    ) {
      password += `:${decodeSafe(u.password)}`;
    }
    return {
      name: decodeSafe(u.hash),
      password: password,
      server: u.host,
      port: u.port,
      sni: u.params.sni || "",
      pinnedPeerCertSha256:
        u.params.pinSHA256 ||
        u.params.pinSha256 ||
        u.params.pin_sha256 ||
        u.params.pinned_peer_cert_sha256 ||
        u.params.pinnedPeerCertSha256 ||
        "",
      verifyPeerCertByName:
        u.params.verifyPeerCertByName ||
        u.params.verify_peer_cert_by_name ||
        "",
      obfs: u.params.obfs || "none",
      obfsPassword: u.params["obfs-password"] || "",
      protocol: "hysteria2",
    };
  } else if (
    url.toLowerCase().startsWith("http://") ||
    url.toLowerCase().startsWith("https://")
  ) {
    const u = parseURL(url);
    return {
      username: decodeSafe(u.username),
      password: decodeSafe(u.password),
      host: u.host,
      port: u.port,
      protocol: u.protocol,
      name: decodeSafe(u.hash),
    };
  } else if (url.toLowerCase().startsWith("socks5://")) {
    const u = parseURL(url);
    return {
      username: decodeSafe(u.username),
      password: decodeSafe(u.password),
      host: u.host,
      port: u.port,
      protocol: u.protocol,
      name: decodeSafe(u.hash),
    };
  } else if (url.toLowerCase().startsWith("anytls://")) {
    const u = parseURL(url);
    const auth = u.username ? decodeSafe(u.username) : "";
    const sni = u.params.peer || u.params.sni || "";
    return {
      name: decodeSafe(u.hash),
      host: u.host,
      port: u.port,
      auth: auth,
      sni: sni,
      pinnedPeerCertSha256:
        u.params.pinnedPeerCertSha256 || u.params.pinned_peer_cert_sha256 || "",
      verifyPeerCertByName:
        u.params.verifyPeerCertByName ||
        u.params.verify_peer_cert_by_name ||
        "",
      allowInsecure:
        u.params.allow_insecure === "true" || u.params.allow_insecure === "1",
      minIdleSession: u.params.minIdleSession || "",
      protocol: "anytls",
    };
  } else if (url.toLowerCase().startsWith("wireguard://")) {
    // layout of kernel/serverObj/wireguard.go:
    // wireguard://PRIVATEKEY@server:port?publicKey=&address=&preSharedKey=&allowedIPs=&keepAlive=&mtu=#name
    const u = parseURL(url);
    return {
      protocol: "wireguard",
      name: decodeSafe(u.hash),
      address: u.host,
      port: u.port,
      privateKey: decodeSafe(u.username || ""),
      publicKey: u.params.publicKey || "",
      localAddress: u.params.address || "",
      dns: u.params.dns || "",
      mtu: u.params.mtu || "",
      allowedIPs: u.params.allowedIPs || "",
      persistentKeepalive: u.params.keepAlive || "",
      preSharedKey: u.params.preSharedKey || "",
      endpoint: u.params.endpoint || "",
    };
  }
  return null;
}

/** generateShareLink turns the editor's form object back into a share link, or null when the protocol is unknown. */
export function generateShareLink(srcObj: ShareForm): string | null {
  let query: ShareForm = {};
  let obj: ShareForm = {};
  let tmp;
  switch (srcObj.protocol) {
    case "vless":
      // todo: support reality and xhttp
      // https://github.com/XTLS/Xray-core/discussions/716
      query = {
        type: srcObj.net,
        flow: srcObj.flow || "",
        security: srcObj.tls,
        fp: srcObj.fp || "",
        path: srcObj.path,
        host: srcObj.host,
        headerType: srcObj.type,
        sni: srcObj.sni,
        pinnedPeerCertSha256: srcObj.pinnedPeerCertSha256,
        verifyPeerCertByName: srcObj.verifyPeerCertByName,
      };
      if (srcObj.alpn !== "") {
        query.alpn = srcObj.alpn;
      }
      if (srcObj.net === "ws") {
        if (srcObj.maxEarlyData) {
          query.maxEarlyData = srcObj.maxEarlyData;
        }
        if (srcObj.earlyDataHeaderName) {
          query.earlyDataHeaderName = srcObj.earlyDataHeaderName;
        }
      }
      if (srcObj.net === "grpc") {
        query.serviceName = srcObj.path;
        if (srcObj.multiMode) {
          query.multiMode = srcObj.multiMode;
        }
        if (srcObj.idleTimeout) {
          query.idleTimeout = srcObj.idleTimeout;
        }
        if (srcObj.healthCheckTimeout) {
          query.healthCheckTimeout = srcObj.healthCheckTimeout;
        }
        if (srcObj.permitWithoutStream) {
          query.permitWithoutStream = srcObj.permitWithoutStream;
        }
        if (srcObj.initialWindowsSize) {
          query.initialWindowsSize = srcObj.initialWindowsSize;
        }
      }
      if (srcObj.net === "mkcp" || srcObj.net === "kcp") {
        query.seed = srcObj.path;
      }
      if (srcObj.net === "quic") {
        query.key = srcObj.key;
        query.quicSecurity = srcObj.quicSecurity;
      }
      if (srcObj.net === "xhttp") {
        query.xhttpMode = srcObj.xhttpMode;
        if (srcObj.noGRPCHeader) query.noGRPCHeader = "true";
        if (srcObj.noSSEHeader) query.noSSEHeader = "true";
        if (srcObj.uplinkHTTPMethod)
          query.uplinkHTTPMethod = srcObj.uplinkHTTPMethod;
        if (srcObj.scMaxEachPostBytesFrom)
          query.scMaxEachPostBytesFrom = srcObj.scMaxEachPostBytesFrom;
        if (srcObj.scMaxEachPostBytesTo)
          query.scMaxEachPostBytesTo = srcObj.scMaxEachPostBytesTo;
        if (srcObj.scMinPostsIntervalFrom)
          query.scMinPostsIntervalFrom = srcObj.scMinPostsIntervalFrom;
        if (srcObj.scMinPostsIntervalTo)
          query.scMinPostsIntervalTo = srcObj.scMinPostsIntervalTo;
        if (srcObj.scMaxBufferedPosts)
          query.scMaxBufferedPosts = srcObj.scMaxBufferedPosts;
        if (srcObj.scStreamUpServerFrom)
          query.scStreamUpServerFrom = srcObj.scStreamUpServerFrom;
        if (srcObj.scStreamUpServerTo)
          query.scStreamUpServerTo = srcObj.scStreamUpServerTo;
        if (srcObj.xPaddingBytesFrom)
          query.xPaddingBytesFrom = srcObj.xPaddingBytesFrom;
        if (srcObj.xPaddingBytesTo)
          query.xPaddingBytesTo = srcObj.xPaddingBytesTo;
        if (srcObj.xmuxMaxConcurFrom)
          query.xmuxMaxConcurFrom = srcObj.xmuxMaxConcurFrom;
        if (srcObj.xmuxMaxConcurTo)
          query.xmuxMaxConcurTo = srcObj.xmuxMaxConcurTo;
        if (srcObj.xmuxMaxConnFrom)
          query.xmuxMaxConnFrom = srcObj.xmuxMaxConnFrom;
        if (srcObj.xmuxMaxConnTo) query.xmuxMaxConnTo = srcObj.xmuxMaxConnTo;
        if (srcObj.xmuxCMaxReuseFrom)
          query.xmuxCMaxReuseFrom = srcObj.xmuxCMaxReuseFrom;
        if (srcObj.xmuxCMaxReuseTo)
          query.xmuxCMaxReuseTo = srcObj.xmuxCMaxReuseTo;
        if (srcObj.xmuxHMaxReqFrom)
          query.xmuxHMaxReqFrom = srcObj.xmuxHMaxReqFrom;
        if (srcObj.xmuxHMaxReqTo) query.xmuxHMaxReqTo = srcObj.xmuxHMaxReqTo;
        if (srcObj.xmuxHMaxReusableFrom)
          query.xmuxHMaxReusableFrom = srcObj.xmuxHMaxReusableFrom;
        if (srcObj.xmuxHMaxReusableTo)
          query.xmuxHMaxReusableTo = srcObj.xmuxHMaxReusableTo;
        if (srcObj.xmuxHKeepAlive) query.xmuxHKeepAlive = srcObj.xmuxHKeepAlive;
        if (srcObj.xhttpHeaders && srcObj.xhttpHeaders.length > 0) {
          const hdrsObj: ShareForm = {};
          srcObj.xhttpHeaders.forEach((h: ShareForm) => {
            if (h.key) hdrsObj[h.key] = h.value;
          });
          if (Object.keys(hdrsObj).length > 0)
            query.xhttpHeaders = JSON.stringify(hdrsObj);
        }
      }
      if (query.security == "reality") {
        query.pbk = srcObj.pbk;
        query.sid = srcObj.sid;
        query.spx = srcObj.spx;
      }
      return buildURL({
        protocol: "vless",
        username: srcObj.id,
        host: srcObj.add,
        port: srcObj.port,
        hash: srcObj.ps,
        params: query,
      });
    case "vmess":
      //https://github.com/2dust/v2rayN/wiki/%E5%88%86%E4%BA%AB%E9%93%BE%E6%8E%A5%E6%A0%BC%E5%BC%8F%E8%AF%B4%E6%98%8E(ver-2)
      obj = Object.assign({}, srcObj);
      // xhttpHeaders is stored on the backend as a JSON-encoded string; the
      // GUI model uses an array of {key,value}. Serialize as a string so the
      // vmess payload can round-trip through the backend parser.
      if (Array.isArray(obj.xhttpHeaders)) {
        const hdrsObj: ShareForm = {};
        obj.xhttpHeaders.forEach((h: ShareForm) => {
          if (h && h.key) hdrsObj[h.key] = h.value;
        });
        obj.xhttpHeaders =
          Object.keys(hdrsObj).length > 0 ? JSON.stringify(hdrsObj) : "";
      }
      // multiMode/permitWithoutStream are stored as strings on the backend
      // but modeled as booleans in the GUI; serialize them as strings.
      if (typeof obj.multiMode === "boolean")
        obj.multiMode = obj.multiMode ? "true" : "false";
      if (typeof obj.permitWithoutStream === "boolean")
        obj.permitWithoutStream = obj.permitWithoutStream ? "true" : "false";
      // kernel/serverObj/v2ray.go tags the uTLS fingerprint "fingerprint"
      if (obj.fp) obj.fingerprint = obj.fp;
      delete obj.fp;
      switch (obj.net) {
        case "kcp":
        case "tcp":
        case "quic":
          break;
        default:
          obj.type = "";
      }
      switch (obj.net) {
        case "ws":
        case "h2":
        case "http":
        case "quic":
        case "grpc":
        case "kcp":
        case "mkcp":
          break;
        default:
          if (obj.net === "tcp" && obj.type === "http") {
            break;
          }
          obj.path = "";
      }
      return "vmess://" + Base64.encode(JSON.stringify(obj));
    case "ss":
      /* ss://BASE64URL(method:password)@server:port#name; SIP002 userinfo is
           base64url without padding and the backend decodes only that. */
      tmp = `ss://${Base64.encodeURI(`${srcObj.method}:${srcObj.password}`)}@${
        srcObj.server
      }:${srcObj.port}/`;
      if (srcObj.plugin) {
        const plugin = [srcObj.plugin];
        if (srcObj.plugin === "v2ray-plugin") {
          if (srcObj.tls) {
            plugin.push("tls");
          }
          if (srcObj.mode !== "websocket") {
            plugin.push("mode=" + srcObj.mode);
          }
          if (srcObj.host) {
            plugin.push("host=" + srcObj.host);
          }
          if (srcObj.path) {
            if (!srcObj.path.startsWith("/")) {
              srcObj.path = "/" + srcObj.path;
            }
            plugin.push("path=" + srcObj.path);
          }
          if (srcObj.impl) {
            plugin.push("impl=" + srcObj.impl);
          }
        } else {
          plugin.push("obfs=" + srcObj.obfs);
          plugin.push("obfs-host=" + srcObj.host);
          if (srcObj.obfs === "http") {
            plugin.push("obfs-uri=" + srcObj.path);
          }
          if (srcObj.impl) {
            plugin.push("impl=" + srcObj.impl);
          }
        }
        tmp += `?plugin=${encodeURIComponent(plugin.join(";"))}`;
      }
      if (srcObj.backend) {
        tmp += `${srcObj.plugin ? "&" : "?"}v2raya-backend=${encodeURIComponent(srcObj.backend)}`;
      }
      tmp += srcObj.name.length ? `#${encodeURIComponent(srcObj.name)}` : "";
      return tmp;

    case "ssr":
      /* ssr://server:port:proto:method:obfs:URLBASE64(password)/?remarks=URLBASE64(remarks)&protoparam=URLBASE64(protoparam)&obfsparam=URLBASE64(obfsparam)) */
      return `ssr://${Base64.encode(
        `${srcObj.server}:${srcObj.port}:${srcObj.proto}:${srcObj.method}:${
          srcObj.obfs
        }:${Base64.encodeURI(srcObj.password)}/?remarks=${Base64.encodeURI(
          srcObj.name,
        )}&protoparam=${Base64.encodeURI(
          srcObj.protoParam,
        )}&obfsparam=${Base64.encodeURI(srcObj.obfsParam)}`,
      )}`;
    case "trojan":
      /* trojan://password@server:port?allowInsecure=1&sni=sni#URIESCAPE(name) */
      query = {
        type: srcObj.net,
        pinnedPeerCertSha256: srcObj.pinnedPeerCertSha256,
        verifyPeerCertByName: srcObj.verifyPeerCertByName,
      };
      if (srcObj.peer !== "") {
        query.sni = srcObj.peer;
      }
      tmp = "trojan";
      if (srcObj.method !== "origin" || srcObj.obfs !== "none") {
        tmp = "trojan-go";
        query.type = srcObj.obfs === "none" ? "original" : "ws";
        if (srcObj.method === "shadowsocks") {
          query.encryption = `ss;${srcObj.ssCipher};${srcObj.ssPassword}`;
        }
        if (query.type === "ws") {
          query.host = srcObj.host || "";
          query.path = srcObj.path || "/";
        }
      }
      if (tmp === "trojan" && (srcObj.net === "ws" || srcObj.net === "h2")) {
        query.host = srcObj.host;
        query.path = srcObj.path;
      }

      if (srcObj.alpn !== "") {
        query.alpn = srcObj.alpn;
      }
      if (srcObj.net === "grpc") {
        query.serviceName = srcObj.path;
      }
      if (srcObj.net === "mkcp" || srcObj.net === "kcp") {
        query.seed = srcObj.path;
      }
      if (srcObj.backend) {
        query["v2raya-backend"] = srcObj.backend;
      }
      return buildURL({
        protocol: tmp,
        username: srcObj.password,
        host: srcObj.server,
        port: srcObj.port,
        hash: srcObj.name,
        params: query,
      });
    case "juicity":
      query = {
        congestion_control: srcObj.cc,
      };
      if (srcObj.sni !== "") {
        query.sni = srcObj.sni;
      }
      if (srcObj.pinnedCertchainSha256 !== "") {
        query.pinned_certchain_sha256 = srcObj.pinnedCertchainSha256;
      }
      if (srcObj.allowInsecure) {
        query.allow_insecure = "true";
      }
      return buildURL({
        protocol: "juicity",
        username: srcObj.uuid,
        password: srcObj.password,
        host: srcObj.server,
        port: srcObj.port,
        hash: srcObj.name,
        params: query,
      });
    case "tuic":
      query = {
        pinnedPeerCertSha256: srcObj.pinnedPeerCertSha256,
        verifyPeerCertByName: srcObj.verifyPeerCertByName,
        congestion_control: srcObj.cc,
        disable_sni: srcObj.disableSni,
        alpn: srcObj.alpn,
        udp_relay_mode: srcObj.udpRelayMode,
      };
      if (srcObj.sni !== "") {
        query.sni = srcObj.sni;
      }
      if (srcObj.allowInsecure) {
        query.allow_insecure = "true";
      }
      return buildURL({
        protocol: "tuic",
        username: srcObj.uuid,
        password: srcObj.password,
        host: srcObj.server,
        port: srcObj.port,
        hash: srcObj.ps || srcObj.name,
        params: query,
      });
    case "hysteria2":
      query = {
        pinned_peer_cert_sha256: srcObj.pinnedPeerCertSha256,
        verify_peer_cert_by_name: srcObj.verifyPeerCertByName,
      };
      if (srcObj.sni !== "") {
        query.sni = srcObj.sni;
      }
      if (srcObj.obfs !== "none") {
        query.obfs = srcObj.obfs;
        query["obfs-password"] = srcObj.obfsPassword;
      }
      return buildURL({
        protocol: "hysteria2",
        username: srcObj.password,
        host: srcObj.server,
        port: srcObj.port,
        hash: srcObj.ps || srcObj.name,
        params: query,
      });
    case "http":
    case "https":
      tmp = {
        protocol: srcObj.protocol + "-proxy",
        host: srcObj.host,
        port: srcObj.port,
        hash: srcObj.name,
      };
      if (srcObj.username && srcObj.password) {
        Object.assign(tmp, {
          username: srcObj.username,
          password: srcObj.password,
        });
      }
      return buildURL(tmp);
    case "socks5":
      tmp = {
        protocol: "socks5",
        host: srcObj.host,
        port: srcObj.port,
        hash: srcObj.name,
      };
      if (srcObj.username && srcObj.password) {
        Object.assign(tmp, {
          username: srcObj.username,
          password: srcObj.password,
        });
      }
      return buildURL(tmp);
    case "anytls":
      if (srcObj.sni) {
        // the backend (kernel/serverObj/anytls.go) reads sni, not peer
        query.sni = srcObj.sni;
      }
      if (srcObj.pinnedPeerCertSha256) {
        query.pinnedPeerCertSha256 = srcObj.pinnedPeerCertSha256;
      }
      if (srcObj.verifyPeerCertByName) {
        query.verifyPeerCertByName = srcObj.verifyPeerCertByName;
      }
      if (srcObj.allowInsecure) {
        query.allow_insecure = "true";
      }
      if (srcObj.minIdleSession) {
        query.minIdleSession = srcObj.minIdleSession;
      }
      return buildURL({
        protocol: "anytls",
        username: srcObj.auth,
        host: srcObj.host,
        port: srcObj.port,
        hash: srcObj.name,
        params: query,
      });
    case "wireguard":
      // parameter names of kernel/serverObj/wireguard.go; the private key
      // travels as the userinfo
      query = {};
      if (srcObj.publicKey) query.publicKey = srcObj.publicKey;
      if (srcObj.localAddress) query.address = srcObj.localAddress;
      if (srcObj.mtu) query.mtu = srcObj.mtu;
      if (srcObj.allowedIPs) query.allowedIPs = srcObj.allowedIPs;
      if (srcObj.persistentKeepalive)
        query.keepAlive = srcObj.persistentKeepalive;
      if (srcObj.preSharedKey) query.preSharedKey = srcObj.preSharedKey;
      return buildURL({
        protocol: "wireguard",
        username: srcObj.privateKey,
        host: srcObj.address,
        port: srcObj.port,
        hash: srcObj.name,
        params: query,
      });
  }
  return null;
}
