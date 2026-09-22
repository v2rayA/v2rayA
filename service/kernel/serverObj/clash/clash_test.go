package clash

import (
	"encoding/base64"
	"strings"
	"testing"
)

const clashSample = `
port: 7890
proxies:
  - {name: any1, server: a.example.com, port: 6792, type: anytls, client-fingerprint: chrome, alpn: [h2], password: pw1, sni: sni.example.com, skip-cert-verify: true, udp: true}
  - name: tr1
    type: trojan
    server: t.example.com
    port: 443
    password: pw2
    sni: t.example.com
    network: ws
    ws-opts:
      path: /ws
      headers:
        Host: h.example.com
  - {name: vl1, type: vless, server: v.example.com, port: 443, uuid: 8f6d1a3b-0000-4000-8000-000000000001, network: tcp, tls: true, servername: v.example.com, flow: xtls-rprx-vision, client-fingerprint: chrome, reality-opts: {public-key: pbk, short-id: abcd, spider-x: /spider}}
  - {name: vm1, type: vmess, server: m.example.com, port: 80, uuid: 8f6d1a3b-0000-4000-8000-000000000002, alterId: 0, cipher: auto, client-fingerprint: firefox, network: ws, ws-opts: {path: /vm, headers: {Host: m.example.com}}}
  - {name: 香港 ss, type: ss, server: s.example.com, port: 8388, cipher: aes-128-gcm, password: 密码, plugin: obfs, plugin-opts: {mode: tls, host: cdn.example.com}}
  - {name: hy1, type: hysteria2, server: h.example.com, port: 443, password: pw3, sni: h.example.com, obfs: salamander, obfs-password: ob}
  - {name: tu1, type: tuic, server: u.example.com, port: 443, uuid: 8f6d1a3b-0000-4000-8000-000000000003, password: pw4, congestion-controller: bbr, udp-relay-mode: native, alpn: [h3], sni: u.example.com}
  - {name: sk1, type: socks5, server: k.example.com, port: 1080, username: u, password: p}
  - {name: ht1, type: http, server: p.example.com, port: 8080}
  - {name: unknown1, type: wireguard, server: w.example.com, port: 51820}
proxy-groups:
  - name: auto
    type: url-test
    url: http://www.gstatic.com/generate_204
    proxies: [any1, tr1]
`

func TestResolveClash(t *testing.T) {
	infos, ok, err := Resolve(clashSample)
	if !ok || err != nil {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
	want := []struct {
		name, protocol, host string
		port                 int
	}{
		{"any1", "anytls", "a.example.com", 6792},
		{"tr1", "trojan", "t.example.com", 443},
		{"vl1", "vless", "v.example.com", 443},
		{"vm1", "vmess", "m.example.com", 80},
		{"香港 ss", "shadowsocks", "s.example.com", 8388},
		{"hy1", "hysteria2", "h.example.com", 443},
		{"tu1", "tuic", "u.example.com", 443},
		{"sk1", "socks5", "k.example.com", 1080},
		{"ht1", "http", "p.example.com", 8080},
	}
	if len(infos) != len(want) {
		for _, i := range infos {
			t.Logf("got %s %s %s:%d", i.GetName(), i.GetProtocol(), i.GetHostname(), i.GetPort())
		}
		t.Fatalf("got %d nodes, want %d", len(infos), len(want))
	}
	for i, w := range want {
		g := infos[i]
		if g.GetName() != w.name || g.GetProtocol() != w.protocol || g.GetHostname() != w.host || g.GetPort() != w.port {
			t.Errorf("node %d: got %s %s %s:%d, want %s %s %s:%d", i, g.GetName(), g.GetProtocol(), g.GetHostname(), g.GetPort(), w.name, w.protocol, w.host, w.port)
		}
	}
	// reality and uTLS parameters must survive the mapping: a node that drops
	// them fails the handshake the provider configured.
	vless := infos[2].ExportToURL()
	for _, s := range []string{"pbk=pbk", "sid=abcd", "spx=%2Fspider", "fp=chrome"} {
		if !contains(vless, s) {
			t.Errorf("vless link %q lacks %q", vless, s)
		}
	}
	vmess := infos[3].ExportToURL()
	raw, err := base64.StdEncoding.WithPadding(base64.NoPadding).DecodeString(strings.TrimPrefix(vmess, "vmess://"))
	if err != nil {
		t.Fatalf("vmess link %q is not base64: %v", vmess, err)
	}
	if !contains(string(raw), `"fingerprint":"firefox"`) {
		t.Errorf("vmess body %s lacks the uTLS fingerprint", raw)
	}

	// the anytls link must carry what the outbound builder reads
	link := infos[0].ExportToURL()
	for _, s := range []string{"sni=sni.example.com", "anytls://pw1@"} {
		if !contains(link, s) {
			t.Errorf("anytls link %q lacks %q", link, s)
		}
	}
}

func TestResolveClashRejectsNonClash(t *testing.T) {
	if _, ok, _ := Resolve("trojan://pw@1.2.3.4:443#a\n"); ok {
		t.Error("a link list must not be taken for a Clash config")
	}
	if _, ok, err := Resolve("proxies:\n  - {name: x, type: wireguard, server: w, port: 1}\n"); !ok || err == nil {
		t.Error("a Clash config with only unsupported types must report an error")
	}
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
