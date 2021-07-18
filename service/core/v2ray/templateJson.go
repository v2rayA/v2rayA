package v2ray

const TemplateJson = `{
  "template": {
    "log": {
      "access": "/var/log/v2ray.log",
      "error": "/var/log/v2ray.err.log",
      "loglevel": "none"
    },
    "inbounds": [
      {
        "port": 20170,
        "listen": "0.0.0.0",
        "protocol": "socks",
        "sniffing": {
          "enabled": true,
          "destOverride": ["http", "tls"]
        },
        "settings": {
          "auth": "noauth",
          "udp": true,
          "ip": null,
          "clients": null
        },
        "streamSettings": null,
        "tag": "socks"
      },
      {
        "port": 20171,
        "listen": "0.0.0.0",
        "protocol": "http",
        "sniffing": {
          "enabled": true,
          "destOverride": ["http", "tls"]
        },
        "tag": "http"
      },
      {
        "port": 20172,
        "listen": "0.0.0.0",
        "protocol": "http",
        "sniffing": {
          "enabled": true,
          "destOverride": ["http", "tls"]
        },
        "tag": "rule"
      }
    ],
    "outbounds": [
      {
        "tag": "proxy",
        "protocol": "vmess",
        "settings": {
          "vnext": null,
          "servers": null
        },
        "streamSettings": null,
        "mux": null
      },
      {
        "protocol": "freedom",
        "settings": {},
        "tag": "direct"
      },
      {
        "protocol": "blackhole",
        "settings": {},
        "tag": "block"
      }
    ],
    "routing": {
      "domainStrategy": "IPOnDemand",
      "rules": []
    }
  },
  "tcpSettings": {
    "connectionReuse": true,
    "header": {
      "type": "http",
      "request": {
        "version": "1.1",
        "method": "GET",
        "path": ["/"],
        "headers": {
          "Host": ["host"],
          "User-Agent": [
            "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.75 Safari/537.36",
            "Mozilla/5.0 (iPhone; CPU iPhone OS 10_0_2 like Mac OS X) AppleWebKit/601.1 (KHTML, like Gecko) CriOS/53.0.2785.109 Mobile/14A456 Safari/601.1.46"
          ],
          "Accept-Encoding": ["gzip, deflate"],
          "Connection": ["keep-alive"],
          "Pragma": "no-cache"
        }
      },
      "response": {
        "version": "1.1",
        "status": "200",
        "reason": "OK",
        "headers": {
          "Content-Type": ["application/octet-stream", "video/mpeg"],
          "Transfer-Encoding": ["chunked"],
          "Connection": ["keep-alive"],
          "Pragma": "no-cache"
        }
      }
    }
  },
  "wsSettings": {
    "connectionReuse": true,
    "path": "",
    "headers": {
      "Host": "host"
    }
  },
  "tlsSettings": {
    "allowInsecure": false,
    "allowInsecureCiphers": false,
    "serverName": null
  },
  "kcpSettings": {
    "mtu": 1350,
    "tti": 50,
    "uplinkCapacity": 12,
    "downlinkCapacity": 100,
    "congestion": false,
    "readBufferSize": 2,
    "writeBufferSize": 2,
    "header": {
      "type": "none",
      "request": null,
      "response": null
    }
  },
  "httpSettings": {
    "path": "path",
    "host": ["host"]
  },
  "streamSettings": {
    "network": "ws",
    "security": "",
    "tlsSettings": null,
    "tcpSettings": null,
    "kcpSettings": null,
    "wsSettings": null,
    "httpSettings": null
  },
  "mux": {
    "enabled": false,
    "concurrency": 8
  }
}`