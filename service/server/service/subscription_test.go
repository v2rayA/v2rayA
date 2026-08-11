package service

import (
	"testing"

	"github.com/v2rayA/v2rayA/db/configure"
)

func TestConnectedSubscriptionServersOnlyTargetsOneSource(t *testing.T) {
	connected := &configure.Whiches{Touches: []*configure.Which{
		{TYPE: configure.SubscriptionServerType, ID: 1, Sub: 0, Outbound: "proxy"},
		{TYPE: configure.SubscriptionServerType, ID: 2, Sub: 0, Outbound: "proxy"},
		{TYPE: configure.SubscriptionServerType, ID: 1, Sub: 1, Outbound: "proxy"},
		{TYPE: configure.ServerType, ID: 1, Outbound: "proxy"},
	}}

	got := connectedSubscriptionServers(0, connected)
	if len(got) != 2 {
		t.Fatalf("matched connections = %d, want 2", len(got))
	}
	for _, which := range got {
		if which.Sub != 0 || which.TYPE != configure.SubscriptionServerType {
			t.Fatalf("matched unexpected connection: %+v", which)
		}
	}
}

func TestConnectedSubscriptionServersHandlesNil(t *testing.T) {
	if got := connectedSubscriptionServers(0, nil); got != nil {
		t.Fatalf("matched connections = %+v, want nil", got)
	}
}
