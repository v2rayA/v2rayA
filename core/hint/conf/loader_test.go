// SPDX-License-Identifier: MPL-2.0
package conf

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	hint_tunmips "github.com/v2rayA/v2raya-core/hint/proxy/tunmips"
	xray_proxyman "github.com/xtls/xray-core/app/proxyman"
	xray_core "github.com/xtls/xray-core/core"
	_ "github.com/xtls/xray-core/main/confloader/external"
)

const tunMipsDoc = `{
  "inbounds": [
    {"port": 20170, "protocol": "socks", "listen": "127.0.0.1", "settings": {"udp": true}, "tag": "socks"},
    {
      "protocol": "tun-mips",
      "tag": "transparent",
      "sniffing": {"enabled": true, "destOverride": ["http", "tls", "quic"], "routeOnly": false},
      "settings": {
        "name": "tun0",
        "mtu": 1500,
        "address4": "10.0.85.2/30",
        "address6": "%s",
        "dnsTarget": "127.0.0.1:52353",
        "excludeProcesses": ["curl", "v2raya"],
        "selfPids": [4242]
      }
    }
  ],
  "outbounds": [
    {"protocol": "freedom", "tag": "direct"}
  ]
}`

func tunMipsJSON(address6 string) []byte {
	return []byte(strings.Replace(tunMipsDoc, "%s", address6, 1))
}

// loadBothWays runs the document through the file builder and the reader
// loader, the two production entry points, and checks they agree.
func loadBothWays(t *testing.T, raw []byte) *xray_core.Config {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	fromFile, err := buildConfigFromFiles([]*xray_core.ConfigSource{{Name: path, Format: "json"}})
	if err != nil {
		t.Fatalf("buildConfigFromFiles: %v", err)
	}
	fromReader, err := xray_core.LoadConfig("json", bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("LoadConfig(reader): %v", err)
	}
	if len(fromFile.Inbound) != len(fromReader.Inbound) || len(fromFile.Outbound) != len(fromReader.Outbound) {
		t.Fatalf("file and reader paths disagree: %d/%d inbounds, %d/%d outbounds",
			len(fromFile.Inbound), len(fromReader.Inbound), len(fromFile.Outbound), len(fromReader.Outbound))
	}
	return fromFile
}

func findInbound(t *testing.T, cfg *xray_core.Config, tag string) *xray_core.InboundHandlerConfig {
	t.Helper()
	for _, ib := range cfg.Inbound {
		if ib.Tag == tag {
			return ib
		}
	}
	t.Fatalf("no inbound tagged %q among %d", tag, len(cfg.Inbound))
	return nil
}

func TestTunMipsInboundIsBuilt(t *testing.T) {
	for _, address6 := range []string{"fdfe:dcba:9876::2/126", ""} {
		t.Run("address6="+address6, func(t *testing.T) {
			cfg := loadBothWays(t, tunMipsJSON(address6))
			if len(cfg.Inbound) != 2 {
				t.Fatalf("%d inbounds, want the socks one and tun-mips", len(cfg.Inbound))
			}
			// The native inbound went through xray untouched.
			findInbound(t, cfg, "socks")

			ib := findInbound(t, cfg, "transparent")
			proxy, err := ib.ProxySettings.GetInstance()
			if err != nil {
				t.Fatal(err)
			}
			tc, ok := proxy.(*hint_tunmips.Config)
			if !ok {
				t.Fatalf("proxy settings are %T", proxy)
			}
			if tc.Name != "tun0" || tc.Mtu != 1500 || tc.Address4 != "10.0.85.2/30" || tc.Address6 != address6 ||
				tc.DnsTarget != "127.0.0.1:52353" || len(tc.ExcludeProcesses) != 2 || len(tc.SelfPids) != 1 || tc.SelfPids[0] != 4242 {
				t.Fatalf("config %+v", tc)
			}

			receiver, err := ib.ReceiverSettings.GetInstance()
			if err != nil {
				t.Fatal(err)
			}
			rc := receiver.(*xray_proxyman.ReceiverConfig)
			if rc.SniffingSettings == nil || !rc.SniffingSettings.Enabled || len(rc.SniffingSettings.DestinationOverride) != 3 {
				t.Fatalf("sniffing settings %+v", rc.SniffingSettings)
			}
			if rc.PortList != nil {
				t.Fatalf("tun-mips inbound must not listen, got ports %+v", rc.PortList)
			}
		})
	}
}

func TestWrongDirectionIsRejectedNotDropped(t *testing.T) {
	for _, tc := range []struct{ name, doc, want string }{
		{"tun-mips as outbound",
			`{"outbounds": [{"protocol": "freedom", "tag": "direct"}, {"protocol": "tun-mips", "tag": "oops", "settings": {}}]}`,
			"not valid as outbound"},
		{"anytls as inbound",
			`{"inbounds": [{"protocol": "anytls", "tag": "oops", "settings": {}}]}`,
			"not valid as inbound"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := stripCustom([]byte(tc.doc))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want %q", err, tc.want)
			}
			if _, err := xray_core.LoadConfig("json", strings.NewReader(tc.doc)); err == nil {
				t.Fatal("LoadConfig accepted a protocol in the wrong direction")
			}
		})
	}
}

func TestTunMipsMissingRequiredFieldFails(t *testing.T) {
	doc := strings.Replace(string(tunMipsJSON("")), `"dnsTarget": "127.0.0.1:52353",`, "", 1)
	_, err := xray_core.LoadConfig("json", strings.NewReader(doc))
	if err == nil || !strings.Contains(err.Error(), "dnsTarget") {
		t.Fatalf("err = %v, want a dnsTarget complaint", err)
	}
}

// TestDocumentWithoutCustomProtocolsIsUntouched is the behaviour-neutrality
// check for configurations the service generates today.
func TestDocumentWithoutCustomProtocolsIsUntouched(t *testing.T) {
	raw := []byte(`{"inbounds": [{"port": 20170, "protocol": "socks", "tag": "socks"}], "outbounds": [{"protocol": "freedom", "tag": "direct"}]}`)
	modified, custom, err := stripCustom(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(modified, raw) || !custom.empty() {
		t.Fatalf("document was rewritten: %s", modified)
	}
}
