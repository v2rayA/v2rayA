package conf

import "testing"

func TestParametersFollowTheStructTags(t *testing.T) {
	params := Parameters()
	byFlag := map[string]Param{}
	for _, p := range params {
		byFlag[p.Flag] = p
	}
	address, ok := byFlag["--address"]
	if !ok {
		t.Fatal("--address missing")
	}
	if address.Short != "a" || address.Env != "V2RAYA_ADDRESS" || address.Default != "0.0.0.0:2017" {
		t.Errorf("--address = %+v", address)
	}
	if p, ok := byFlag["--core-startup-timeout"]; !ok || p.Env != "V2RAYA_CORE_STARTUP_TIMEOUT" {
		t.Errorf("--core-startup-timeout = %+v, present %v", p, ok)
	}
	// hidden switches without a description are not documented
	if _, ok := byFlag["--reset-password"]; ok {
		t.Error("--reset-password has no desc and must not be listed")
	}
	for _, p := range params {
		if p.Desc == "" {
			t.Errorf("%s has no description", p.Flag)
		}
	}
}
