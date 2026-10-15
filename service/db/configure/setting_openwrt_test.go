package configure

import "testing"

func TestOpenWrtBridgeDefault(t *testing.T) {
	if got := defaultTproxyExcludedInterfaces(true); got != "docker*,veth*,wg*,ppp*" {
		t.Fatalf("OpenWrt default = %q", got)
	}
	if got := defaultTproxyExcludedInterfaces(false); got != legacyTproxyExcludedInterfaces {
		t.Fatalf("other-platform default = %q", got)
	}
}

func TestMigrateOpenWrtBridgeExclusion(t *testing.T) {
	tests := []struct {
		name     string
		openwrt  bool
		value    string
		changed  bool
		expected string
	}{
		{"legacy OpenWrt default", true, legacyTproxyExcludedInterfaces, true, "docker*,veth*,wg*,ppp*"},
		{"custom OpenWrt list", true, "br-guest,wg*", false, "br-guest,wg*"},
		{"other platform", false, legacyTproxyExcludedInterfaces, false, legacyTproxyExcludedInterfaces},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setting := &Setting{TproxyExcludedInterfaces: tt.value}
			if got := MigrateOpenWrtBridgeExclusion(setting, tt.openwrt); got != tt.changed {
				t.Fatalf("changed = %v, want %v", got, tt.changed)
			}
			if setting.TproxyExcludedInterfaces != tt.expected {
				t.Fatalf("value = %q, want %q", setting.TproxyExcludedInterfaces, tt.expected)
			}
		})
	}
}
