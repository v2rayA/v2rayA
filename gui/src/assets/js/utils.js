import CONST from "./const.js";

function locateServer(touch, whichServer) {
  let ind = whichServer.id - 1;
  let sub = whichServer.sub;
  if (whichServer._type === CONST.ServerType) {
    return touch.servers[ind];
  } else if (whichServer._type === CONST.SubscriptionServerType) {
    return touch.subscriptions[sub].servers[ind];
  }
  return null;
}

function handleResponse(res, that, suc, err, fail) {
  if (!res.data) {
    if (err && err instanceof Function) {
      err.apply(that);
    }
    return;
  }
  if (res.data.code === "SUCCESS") {
    suc.apply(that);
  } else {
    if (err && err instanceof Function) {
      err.apply(that);
    } else {
      if (fail && fail instanceof Function) {
        fail.apply(that);
      } else {
        that.$buefy.toast.open({
          message: res.data.message,
          type: "is-warning",
          position: "is-top",
          queue: false,
          duration: 5000
        });
      }
    }
  }
}

/*
var myURL = parseURL('http://user:pass@abc.com:8080/dir/index.html?id=255&m=hello#top');
myURL.username;     // = 'user'
myURL.password;     // = 'pass'
myURL.file;     // = 'index.html'
myURL.hash;     // = 'top'
myURL.host;     // = 'abc.com'
myURL.query;    // = '?id=255&m=hello'
myURL.params;   // = Object = { id: 255, m: hello }
myURL.path;     // = '/dir/index.html'
myURL.segments; // = Array = ['dir', 'index.html']
myURL.port;     // = '8080'
myURL.protocol; // = 'http'
myURL.source;   // = 'http://user:pass@abc.com:8080/dir/index.html?id=255&m=hello#top'
*/
function parseURL(u) {
  let url = u;
  let protocol = "";
  let fakeProto = false;
  if (url.indexOf("://") === -1) {
    url = "http://" + url;
  } else {
    protocol = url.substr(0, url.indexOf("://"));
    switch (protocol) {
      case "http":
      case "https":
      case "ws":
      case "wss":
        break;
      default:
        url = "http" + url.substr(url.indexOf("://"));
        fakeProto = true;
    }
  }
  const a = document.createElement("a");
  a.href = url;
  const r = {
    source: u,
    username: a.username,
    password: a.password,
    protocol: fakeProto ? protocol : a.protocol.replace(":", ""),
    host: a.hostname,
    port: a.port ? parseInt(a.port) : 80,
    query: a.search,
    params: (function() {
      var ret = {},
        seg = a.search.replace(/^\?/, "").split("&"),
        len = seg.length,
        i = 0,
        s;
      for (; i < len; i++) {
        if (!seg[i]) {
          continue;
        }
        s = seg[i].split("=");
        ret[s[0]] = decodeURIComponent(s[1]);
      }
      return ret;
    })(),
    file: (a.pathname.match(/\/([^/?#]+)$/i) || [null, ""])[1],
    hash: a.hash.replace("#", ""),
    path: a.pathname.replace(/^([^/])/, "/$1"),
    relative: (a.href.match(/tps?:\/\/[^/]+(.+)/) || [null, ""])[1],
    segments: a.pathname.replace(/^\//, "").split("/")
  };
  a.remove();
  return r;
}

function generateURL({
  username,
  password,
  protocol,
  host,
  port,
  params,
  hash,
  path
}) {
  const a = document.createElement("a");
  if (protocol) {
    if (protocol.indexOf("://") === -1) {
      protocol = protocol + "://";
    }
  } else {
    protocol = "http://";
  }
  let user = "";
  if (username || password) {
    console.log(username, password, protocol ? protocol : "http://");
    if (username && password) {
      user = `${username}:${password}@`;
    } else {
      user = `${username || password}@`;
    }
  }
  let query = "";
  console.log(params);
  if (params) {
    let first = true;
    for (const k in params) {
      if (!params.hasOwnProperty(k)) {
        continue;
      }
      if (!first) {
        query += "&";
      } else {
        first = false;
      }
      query += `${k}=${encodeURIComponent(params[k])}`;
    }
  }
  console.log(query);
  path = path || "";
  if (path && path.length > 0 && path[0] !== "/") {
    path = "/" + path;
  }
  a.href = `http://${user}${host}${port ? `:${port}` : ""}${path ? path : ""}`;
  console.log(
    `http://${user}${host}${port ? `:${port}` : ""}${path ? path : ""}`
  );
  a.search = query.length ? `?${query}` : "";
  a.hash = hash;
  const r = (protocol ? protocol : "http://") + a.href.substr(7);
  a.remove();
  console.log(r, a.href);
  return r;
}

/*判断一个IPv4的地址是否是内网地址*/
function isIntranet(url) {
  if (!url.trim() || url.startsWith("/")) {
    url = location.protocol + "//" + location.host + url;
  }
  let u = parseURL(url);
  if (u.host === "") {
    u = parseURL(location.protocol + "//" + location.host + url);
  }
  let host = u.host;
  let arr = host.split(".");
  if (arr.length !== 4) {
    return host === "localhost" || host === "local";
  }
  if (arr.some(p => parseInt(p) < 0 || parseInt(p) > 255)) {
    //每一位必须是[0,255]
    return false;
  }
  let bin = ""; //传入的IP的二进制表示
  arr.forEach(p => {
    let t = parseInt(p).toString(2);
    bin += "0".repeat(8 - t.length) + t;
  });
  const list = [
    "0.0.0.0/32",
    "10.0.0.0/8",
    "127.0.0.0/8",
    "169.254.0.0/16",
    "172.16.0.0/12",
    "192.168.0.0/16",
    "224.0.0.0/4",
    "240.0.0.0/4"
  ];
  return list.some(mask => {
    let arr = mask.split("/");
    let prefix = arr[0],
      suffix = arr[1];
    arr = prefix.split(".");
    let b = "";
    arr.forEach(p => {
      let t = parseInt(p).toString(2);
      b += "0".repeat(8 - t.length) + t;
    });
    suffix = parseInt(suffix);
    let is = true;
    for (let i = 0; i < suffix; i++) {
      if (b[i] !== bin[i]) {
        is = false;
        break;
      }
    }
    return is;
  });
}

function isVersionGreaterEqual(va, vb) {
  if (va === "debug" || va === "unstable") {
    return true;
  }
  if (vb === "debug" || vb === "unstable") {
    return false;
  }
  va = va.trim();
  vb = vb.trim();
  if (va.length > 0 && va[0] === "v") {
    va = va.substr(1);
  }
  if (vb.length > 0 && vb[0] === "v") {
    vb = vb.substr(1);
  }
  va.replace("-", ".");
  vb.replace("-", ".");
  let a = va.split(".");
  let b = vb.split(".");
  let minlen = Math.min(a.length, b.length);
  for (let i = 0; i < minlen; i++) {
    if (parseInt(a[i]) < parseInt(b[i])) {
      return false;
    }
    if (parseInt(a[i]) > parseInt(b[i])) {
      return true;
    }
  }
  return a.length >= b.length;
}

function toInt(s) {
  if (typeof s === "string") {
    return parseInt(s);
  } else if (typeof s === "number") {
    return parseInt(s);
  } else if (typeof s === "boolean") {
    return s ? 1 : 0;
  }
  return s;
}

export {
  locateServer,
  handleResponse,
  parseURL,
  generateURL,
  isIntranet,
  isVersionGreaterEqual,
  toInt
};
