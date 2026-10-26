package serverObj

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/v2rayA/v2rayA/kernel/coreObj"
)

func TestVlessEncryptionSurvivesImportStorageAndExport(t *testing.T) {
	for _, encryption := range []string{"", "none", "mlkem768x25519plus.native.0rtt.example-key"} {
		t.Run(encryption, func(t *testing.T) {
			q := url.Values{"type": {"xhttp"}, "security": {"tls"}, "path": {"/test"}}
			if encryption != "" {
				q.Set("encryption", encryption)
			}
			link := "vless://b831381d-6324-4d53-ad4f-8cda48b30811@example.com:443?" + q.Encode()
			obj, err := ParseVlessURL(link)
			if err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(obj)
			if err != nil {
				t.Fatal(err)
			}
			var stored V2Ray
			if err := json.Unmarshal(raw, &stored); err != nil {
				t.Fatal(err)
			}
			cfg, err := stored.Configuration(PriorInfo{Tag: "proxy"})
			if err != nil {
				t.Fatal(err)
			}
			want := encryption
			if want == "" {
				want = "none"
			}
			got := cfg.CoreOutbound.Settings.Vnext.([]coreObj.Vnext)[0].Users[0].Encryption
			if got != want {
				t.Errorf("generated encryption = %q, want %q", got, want)
			}
			exported, err := url.Parse(stored.ExportToURL())
			if err != nil {
				t.Fatal(err)
			}
			if got := exported.Query().Get("encryption"); got != encryption {
				t.Errorf("exported encryption = %q, want %q", got, encryption)
			}
		})
	}
}

func TestVlessECHSurvivesImportStorageAndExport(t *testing.T) {
	const ech = "AAH+/w=="
	q := url.Values{"type": {"xhttp"}, "security": {"tls"}, "sni": {"example.com"}, "ech": {ech}}
	obj, err := ParseVlessURL("vless://b831381d-6324-4d53-ad4f-8cda48b30811@example.com:443?" + q.Encode())
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(obj)
	if err != nil {
		t.Fatal(err)
	}
	var stored V2Ray
	if err := json.Unmarshal(raw, &stored); err != nil {
		t.Fatal(err)
	}
	cfg, err := stored.Configuration(PriorInfo{Tag: "proxy"})
	if err != nil {
		t.Fatal(err)
	}
	raw, err = json.Marshal(cfg.CoreOutbound.StreamSettings.TLSSettings)
	if err != nil {
		t.Fatal(err)
	}
	var tls map[string]interface{}
	if err := json.Unmarshal(raw, &tls); err != nil {
		t.Fatal(err)
	}
	if tls["echConfigList"] != ech {
		t.Errorf("generated echConfigList = %v, want %q", tls["echConfigList"], ech)
	}
	exported, err := url.Parse(stored.ExportToURL())
	if err != nil {
		t.Fatal(err)
	}
	if got := exported.Query().Get("ech"); got != ech {
		t.Errorf("exported ech = %q, want %q", got, ech)
	}
}
