import URI from "urijs";

export interface URLParts {
  username?: string;
  password?: string;
  protocol?: string;
  host?: string;
  port?: number | string;
  params?: Record<string, unknown>;
  hash?: string;
  path?: string;
}

export function parseURL(source: string) {
  let url = source;
  let protocol = "";
  let fakeProtocol = false;
  const separator = url.indexOf("://");
  if (separator === -1) {
    url = "http://" + url;
  } else {
    protocol = url.slice(0, separator);
    if (!["http", "https", "ws", "wss"].includes(protocol)) {
      // Parse custom schemes as HTTP so credentials, host and port follow browser rules.
      url = "http" + url.slice(separator);
      fakeProtocol = true;
    }
  }
  const anchor = document.createElement("a");
  anchor.href = url;
  const params: Record<string, string> = {};
  for (const segment of anchor.search.replace(/^\?/, "").split("&")) {
    if (!segment) continue;
    const [key, raw] = segment.split("=");
    params[key] = decodeSafe(raw);
  }
  return {
    source,
    username: anchor.username,
    password: anchor.password,
    protocol: fakeProtocol ? protocol : anchor.protocol.replace(":", ""),
    host: anchor.hostname,
    port: anchor.port
      ? parseInt(anchor.port)
      : protocol === "https" || protocol === "wss"
        ? 443
        : 80,
    query: anchor.search,
    params,
    file: (anchor.pathname.match(/\/([^/?#]+)$/i) || [null, ""])[1]!,
    hash: anchor.hash.replace("#", ""),
    path: anchor.pathname.replace(/^([^/])/, "/$1"),
    relative: (anchor.href.match(/tps?:\/\/[^/]+(.+)/) || [null, ""])[1]!,
    segments: anchor.pathname.replace(/^\//, "").split("/"),
  };
}

export function generateURL({
  username,
  password,
  protocol,
  host,
  port,
  params,
  hash,
  path,
}: URLParts): string {
  // URI.js treats undefined arguments as getters, which would break the chain.
  return URI()
    .protocol(protocol || "http")
    .username(username || "")
    .password(password || "")
    .host(host || "")
    .port(port || 80)
    .path(path || "")
    .query(params || {})
    .hash(hash || "")
    .toString();
}

/** decodeSafe decodes a URI component, or returns the text as it is when it is not encoded (a name with a literal %). */
export function decodeSafe(text: string): string {
  try {
    return decodeURIComponent(text);
  } catch {
    return text;
  }
}
