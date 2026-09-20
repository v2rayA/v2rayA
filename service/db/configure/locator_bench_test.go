package configure

import (
	"fmt"
	"testing"

	"github.com/v2rayA/v2rayA/kernel/serverObj"
)

func seedSubscription(b *testing.B, n int) (sub int) {
	b.Helper()
	servers := make([]ServerRaw, n)
	for i := range servers {
		servers[i] = ServerRaw{ServerObj: &serverObj.SOCKS{Server: fmt.Sprintf("10.0.%d.%d", i/256, i%256), Port: 1080, Protocol: "socks5", Name: fmt.Sprintf("node %d", i)}}
	}
	sub = GetLenSubscriptions()
	if err := AppendSubscriptions([]*SubscriptionRaw{{Address: "https://example.invalid/sub", Servers: servers}}); err != nil {
		b.Fatal(err)
	}
	return sub
}

func BenchmarkLocateAll(b *testing.B) {
	const n = 1000
	sub := seedSubscription(b, n)
	defer RemoveSubscriptions([]int{sub})
	whiches := make([]*NodeRef, n)
	for i := range whiches {
		whiches[i] = &NodeRef{TYPE: SubscriptionServerType, ID: i + 1, Sub: sub}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, w := range whiches {
			if _, err := w.LocateServerRaw(); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkLocatorAll(b *testing.B) {
	const n = 1000
	sub := seedSubscription(b, n)
	defer RemoveSubscriptions([]int{sub})
	whiches := make([]*NodeRef, n)
	for i := range whiches {
		whiches[i] = &NodeRef{TYPE: SubscriptionServerType, ID: i + 1, Sub: sub}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loc := NewLocator()
		for _, w := range whiches {
			if _, err := loc.Locate(w); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkSetSubscription(b *testing.B) {
	sub := seedSubscription(b, 1000)
	defer RemoveSubscriptions([]int{sub})
	raw := GetSubscription(sub)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := SetSubscription(sub, raw); err != nil {
			b.Fatal(err)
		}
	}
}
