package v2ray

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/v2rayA/v2rayA/db/configure"
)

// ext:"file.dat:tag" in a DNS rule must reach the core's DNS module as
// ext:file.dat:tag so the module expands the tag from that file; before
// this the line was stored as a literal domain and matched nothing.
func TestDnsRuleExtGeositeReachesTheModule(t *testing.T) {
	assetDir := t.TempDir()
	t.Setenv("XRAY_LOCATION_ASSET", assetDir)
	if err := os.WriteFile(filepath.Join(assetDir, "rf-geosite.dat"), []byte{}, 0o644); err != nil {
		t.Fatal(err)
	}

	previous := configure.GetDnsRulesNotNil()
	t.Cleanup(func() {
		if err := configure.SetDnsRules(previous); err != nil {
			t.Error(err)
		}
	})
	if err := configure.SetDnsRules([]configure.DnsRule{
		{Upstream: "1.1.1.1", Domain: `ext:"rf-geosite.dat:RU-BLOCKED"`, Outbound: "proxy"},
	}); err != nil {
		t.Fatal(err)
	}

	tmpl := baseTemplate(t, configure.NewSetting())
	if err := tmpl.setDNS(nil); err != nil {
		t.Fatal(err)
	}
	var config struct {
		Rules []struct {
			Domain []string `json:"domain"`
		} `json:"rules"`
	}
	if err := json.Unmarshal(tmpl.DnsModuleConfig, &config); err != nil {
		t.Fatal(err)
	}
	if len(config.Rules) != 1 || len(config.Rules[0].Domain) != 1 ||
		config.Rules[0].Domain[0] != "ext:rf-geosite.dat:RU-BLOCKED" {
		t.Fatalf("DNS rules = %+v", config.Rules)
	}
}

// The module expands the tag from a file in the asset directory, so a rule
// naming a file that is not there is refused instead of silently matching
// nothing; a value without a tag is not a reference at all.
func TestDnsRuleExtGeositeNeedsAnExistingFileAndATag(t *testing.T) {
	assetDir := t.TempDir()
	t.Setenv("XRAY_LOCATION_ASSET", assetDir)

	previous := configure.GetDnsRulesNotNil()
	t.Cleanup(func() {
		if err := configure.SetDnsRules(previous); err != nil {
			t.Error(err)
		}
	})

	for _, domain := range []string{`ext:"missing.dat:TAG"`, `ext:"rf-geosite.dat"`} {
		if err := configure.SetDnsRules([]configure.DnsRule{
			{Upstream: "1.1.1.1", Domain: domain, Outbound: "direct"},
		}); err != nil {
			t.Fatal(err)
		}
		tmpl := baseTemplate(t, configure.NewSetting())
		if err := tmpl.setDNS(nil); err == nil {
			t.Fatalf("%s accepted without the file or the tag", domain)
		}
	}
}
