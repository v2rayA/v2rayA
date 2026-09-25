package configure

import "testing"

func TestOpenWrtBridgeDefault(t *testing.T) {
	if got := defaultTproxyExcludedInterfaces(true); got != "docker*,veth*,wg*,ppp*" {
		t.Fatal(got)
	}
	if got := defaultTproxyExcludedInterfaces(false); got != "docker*,veth*,wg*,ppp*,br-*" {
		t.Fatal(got)
	}
}
