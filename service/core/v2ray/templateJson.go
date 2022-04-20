package v2ray

const TemplateJson = `
{
    "inbounds": [
        {
            "port": 20170,
            "listen": "0.0.0.0",
            "protocol": "socks",
            "sniffing": {
                "enabled": false,
                "destOverride": [
                    "http",
                    "tls"
                ]
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
                "enabled": false,
                "destOverride": [
                    "http",
                    "tls"
                ]
            },
            "tag": "http"
        },
        {
            "port": 0,
            "listen": "0.0.0.0",
            "protocol": "socks",
            "sniffing": {
                "enabled": false,
                "destOverride": [
                    "http",
                    "tls"
                ]
            },
            "settings": {
                "auth": "noauth",
                "udp": true,
                "ip": null,
                "clients": null
            },
            "streamSettings": null,
            "tag": "rule-socks"
        },
        {
            "port": 20172,
            "listen": "0.0.0.0",
            "protocol": "http",
            "sniffing": {
                "enabled": false,
                "destOverride": [
                    "http",
                    "tls"
                ]
            },
            "tag": "rule-http"
        },
        {
            "listen": "0.0.0.0",
            "port": 0,
            "protocol": "vmess",
            "settings": {
                "clients": [
                    {
                        "id": ""
                    }
                ]
            },
            "tag": "vmess"
        }
    ],
    "outbounds": [],
    "routing": {
        "domainStrategy": "IPOnDemand",
        "rules": []
    }
}`
