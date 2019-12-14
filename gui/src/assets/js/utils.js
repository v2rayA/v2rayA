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

function handleResponse(res, that, suc, err) {
  if (res.data.code === "SUCCESS") {
    suc.apply(that);
  } else {
    if (err && err instanceof Function) {
      err.apply(that);
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

/*
var myURL = parseURL('http://abc.com:8080/dir/index.html?id=255&m=hello#top');
myURL.file;     // = 'index.html'
myURL.hash;     // = 'top'
myURL.host;     // = 'abc.com'
myURL.query;    // = '?id=255&m=hello'
myURL.params;   // = Object = { id: 255, m: hello }
myURL.path;     // = '/dir/index.html'
myURL.segments; // = Array = ['dir', 'index.html']
myURL.port;     // = '8080'
myURL.protocol; // = 'http'
myURL.source;   // = 'http://abc.com:8080/dir/index.html?id=255&m=hello#top'
*/
function parseURL(url) {
  if (url.indexOf("://") === -1) {
    url = "http://" + url;
  }
  var a = document.createElement("a");
  a.href = url;
  return {
    source: url,
    protocol: a.protocol.replace(":", ""),
    host: a.hostname,
    port: a.port,
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
        ret[s[0]] = s[1];
      }
      return ret;
    })(),
    file: (a.pathname.match(/\/([^\/?#]+)$/i) || [, ""])[1],
    hash: a.hash.replace("#", ""),
    path: a.pathname.replace(/^([^\/])/, "/$1"),
    relative: (a.href.match(/tps?:\/\/[^\/]+(.+)/) || [, ""])[1],
    segments: a.pathname.replace(/^\//, "").split("/")
  };
}

/*判断一个IPv4的地址是否是内网地址*/
function isIntranet(url) {
  let host = parseURL(url).host;
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
    "240.0.0.0/4",
    "255.255.255.255/32"
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
  if (va.toLowerCase() === "debug") {
    return true;
  }
  va = va.trim();
  vb = vb.trim();
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

export {
  locateServer,
  handleResponse,
  parseURL,
  isIntranet,
  isVersionGreaterEqual
};
