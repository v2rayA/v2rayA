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
            "protocol": "vless",
            "settings": {
                "clients": [
                    {
                        "id": ""
                    }
                ],
                "decryption": "none"
            },
            "streamSettings": {
                "network": "grpc",
                "security": "tls",
                "tlsSettings": {
                    "serverName": "",
                    "alpn": [
                        "h2"
                    ],
                    "certificates": [
                        {
                            "certificateFile": "/etc/v2raya/vlessGrpc.crt",
                            "keyFile": "/etc/v2raya/vlessGrpc.key"
                        }
                    ]
                },
                "grpcSettings": {
                    "serviceName": "v2rayA_VLESS_GRPC"
                }
            },
            "tag": "vlessGrpc"
        }
    ],
    "outbounds": [],
    "routing": {
        "domainStrategy": "IPOnDemand",
        "rules": []
    }
}`
