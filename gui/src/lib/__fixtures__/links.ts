// Hand-written share links, one per protocol the editor supports, plus a
// VLESS xhttp link with the mux fields. The expected forms and links in
// serverCodec.json were recorded from modalServer.vue's original
// resolveURL/generateURL before the move to lib/serverCodec.ts.
const b64 = (s: string) => Buffer.from(s, "utf8").toString("base64");
export const links = {
  vmess:
    "vmess://" +
    b64(
      JSON.stringify({
        v: "2",
        ps: "vmess node",
        add: "1.2.3.4",
        port: "443",
        id: "b831381d-6324-4d53-ad4f-8cda48b30811",
        aid: "0",
        scy: "auto",
        net: "ws",
        type: "none",
        host: "h.example.com",
        path: "/ws",
        tls: "tls",
        sni: "h.example.com",
        alpn: "h2,http/1.1",
        fp: "chrome",
      }),
    ),
  vless:
    "vless://b831381d-6324-4d53-ad4f-8cda48b30811@1.2.3.4:443?encryption=none&flow=xtls-rprx-vision&security=reality&sni=example.com&fp=chrome&pbk=PUBLICKEY&sid=abcd1234&spx=%2F&type=tcp#vless%20node",
  vlessXhttp:
    "vless://b831381d-6324-4d53-ad4f-8cda48b30811@1.2.3.4:443?encryption=none&security=tls&sni=example.com&type=xhttp&path=%2Fx&xhttpMode=packet-up&xmuxMaxConcurFrom=8&xmuxMaxConcurTo=16&xhttpHeaders=%7B%22X-A%22%3A%22b%22%7D#vless%20xhttp",
  ss:
    "ss://" +
    b64("chacha20-ietf-poly1305:passw0rd") +
    "@1.2.3.4:8388?plugin=obfs-local%3Bobfs%3Dhttp%3Bobfs-host%3Dexample.com#ss%20node",
  ssr:
    "ssr://" +
    b64(
      "1.2.3.4:8388:auth_aes128_md5:aes-256-cfb:tls1.2_ticket_auth:" +
        b64("passw0rd") +
        "/?obfsparam=" +
        b64("obfs.example.com") +
        "&protoparam=" +
        b64("1:abc") +
        "&remarks=" +
        b64("ssr node"),
    ),
  trojan:
    "trojan://passw0rd@1.2.3.4:443?security=tls&sni=example.com&type=ws&host=example.com&path=%2Ftr#trojan%20node",
  juicity:
    "juicity://b831381d-6324-4d53-ad4f-8cda48b30811:passw0rd@1.2.3.4:443?congestion_control=bbr&sni=example.com&allow_insecure=false#juicity%20node",
  tuic: "tuic://b831381d-6324-4d53-ad4f-8cda48b30811:passw0rd@1.2.3.4:443?congestion_control=bbr&alpn=h3&sni=example.com&udp_relay_mode=native&allow_insecure=0#tuic%20node",
  hysteria2:
    "hysteria2://passw0rd@1.2.3.4:443?sni=example.com&insecure=0&obfs=salamander&obfs-password=obf#hy2%20node",
  http: "http://user:passw0rd@1.2.3.4:8080#http%20node",
  socks5: "socks5://user:passw0rd@1.2.3.4:1080#socks%20node",
  anytls:
    "anytls://passw0rd@1.2.3.4:443?sni=example.com&insecure=0#anytls%20node",
  wireguard:
    "wireguard://PRIVATEKEY%3D@1.2.3.4:51820?publicKey=PUBKEY%3D&address=10.0.0.2%2F32&preSharedKey=&allowedIPs=0.0.0.0%2F0&keepAlive=25&mtu=1420#wg%20node",
};
