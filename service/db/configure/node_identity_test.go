package configure

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/v2rayA/v2rayA/kernel/serverObj"
)

func TestNodeIdentityIgnoresPresentationButKeepsCredentialsAndTransport(t *testing.T) {
	base := "vless://account@node.example:443?security=reality&sni=front.example&pbk=key&sid=01&fp=chrome#old"
	renamed := "vless://account@node.example:443?fp=chrome&sid=01&pbk=key&sni=front.example&security=reality#new"
	if NodeFingerprint(base) != NodeFingerprint(renamed) {
		t.Fatal("name or query order changed connection identity")
	}
	for _, changed := range []string{
		"vless://other@node.example:443?security=reality&sni=front.example&pbk=key&sid=01&fp=chrome#old",
		"vless://account@node.example:443?security=reality&sni=other.example&pbk=key&sid=01&fp=chrome#old",
		"vless://account@node.example:443?security=reality&sni=front.example&pbk=key&sid=02&fp=chrome#old",
	} {
		if NodeFingerprint(base) == NodeFingerprint(changed) {
			t.Fatal("connection change was hidden by identity normalization")
		}
	}
}

func TestMigrateNodeFingerprintsPreservesPolicyPinAndUnknownMembers(t *testing.T) {
	const name = "identity-migration"
	if err := AddOutbound(name); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = RemoveOutbound(name) })
	node := &serverObj.SOCKS{Server: "127.0.0.1", Port: 14111, Protocol: "socks5", Name: "old"}
	index := GetLenServers()
	if err := AppendServers([]*ServerRaw{{ServerObj: node}}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = RemoveServers([]int{index}) })
	if err := AddConnect(NodeRef{TYPE: ServerType, ID: index + 1, Outbound: name}); err != nil {
		t.Fatal(err)
	}
	legacy := fmt.Sprintf("%x", sha256.Sum256([]byte(node.ExportToURL())))
	setting := DefaultOutboundSetting()
	setting.Type, setting.AutoAdd, setting.Selected = KeepCurrent, true, node.ExportToURL()
	setting.StickyCurrent, setting.EligibleMembers = legacy, legacy+" missing"
	if err := SetOutboundSetting(name, setting); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := MigrateNodeFingerprints(); err != nil {
			t.Fatal(err)
		}
		want := setting
		want.StickyCurrent = NodeFingerprint(node.ExportToURL())
		want.EligibleMembers = want.StickyCurrent + " missing"
		if got := GetOutboundSetting(name); got != want {
			t.Fatalf("migration changed selection semantics: %+v", got)
		}
	}
}
