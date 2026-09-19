package configure

import (
	"github.com/v2rayA/v2rayA/kernel/serverObj"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestSortSameTypeReverse(t *testing.T) {
	ws := NewWhiches([]*Which{
		{TYPE: SubscriptionType, ID: 1}, {TYPE: ServerType, ID: 1},
		{TYPE: SubscriptionServerType, Sub: 0, ID: 1}, {TYPE: SubscriptionType, ID: 2},
		{TYPE: SubscriptionServerType, Sub: 1, ID: 1}, {TYPE: ServerType, ID: 2},
	})
	ws.SortSameTypeReverse()
	want := []*Which{
		{TYPE: ServerType, ID: 2}, {TYPE: ServerType, ID: 1},
		{TYPE: SubscriptionType, ID: 2}, {TYPE: SubscriptionType, ID: 1},
		{TYPE: SubscriptionServerType, Sub: 1, ID: 1}, {TYPE: SubscriptionServerType, Sub: 0, ID: 1},
	}
	if !reflect.DeepEqual(ws.Get(), want) {
		t.Fatalf("got %+v, want %+v", ws.Get(), want)
	}
}

func TestPingRefusedPortIsNotLatency(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	index := GetLenServers()
	if err := AppendServers([]*ServerRaw{{ServerObj: &serverObj.SOCKS{Server: "127.0.0.1", Port: port, Protocol: "socks5"}}}); err != nil {
		t.Fatal(err)
	}
	defer RemoveServers([]int{index})
	w := &Which{TYPE: ServerType, ID: index + 1}
	if err := w.Ping(NewLocator(), time.Second); err != nil {
		t.Fatal(err)
	}
	if strings.HasSuffix(w.Latency, "ms") || w.Latency == "" {
		t.Fatalf("refused port reported as %q", w.Latency)
	}
}

func TestLocatorMatchesLocateServerRaw(t *testing.T) {
	serverIndex := GetLenServers()
	if err := AppendServers([]*ServerRaw{{ServerObj: &serverObj.SOCKS{Server: "127.0.0.1", Port: 1080, Protocol: "socks5", Name: "plain"}}}); err != nil {
		t.Fatal(err)
	}
	defer RemoveServers([]int{serverIndex})
	subIndex := GetLenSubscriptions()
	if err := AppendSubscriptions([]*SubscriptionRaw{{Address: "https://example.invalid/sub", Servers: []ServerRaw{
		{ServerObj: &serverObj.SOCKS{Server: "127.0.0.1", Port: 1081, Protocol: "socks5", Name: "first"}},
		{ServerObj: &serverObj.SOCKS{Server: "127.0.0.1", Port: 1082, Protocol: "socks5", Name: "second"}},
	}}}); err != nil {
		t.Fatal(err)
	}
	defer RemoveSubscriptions([]int{subIndex})

	loc := NewLocator()
	for _, w := range []*Which{
		{TYPE: ServerType, ID: serverIndex + 1},
		{TYPE: SubscriptionServerType, ID: 2, Sub: subIndex},
	} {
		want, err := w.LocateServerRaw()
		if err != nil {
			t.Fatal(err)
		}
		got, err := loc.Locate(w)
		if err != nil {
			t.Fatal(err)
		}
		if got.ServerObj.GetName() != want.ServerObj.GetName() {
			t.Fatalf("%+v: located %q, want %q", w, got.ServerObj.GetName(), want.ServerObj.GetName())
		}
	}
	for _, w := range []*Which{
		{TYPE: ServerType, ID: serverIndex + 2},
		{TYPE: SubscriptionServerType, ID: 3, Sub: subIndex},
		{TYPE: SubscriptionType, ID: 1},
	} {
		_, want := w.LocateServerRaw()
		_, got := loc.Locate(w)
		if want == nil || got == nil || got.Error() != want.Error() {
			t.Fatalf("%+v: locator error %v, want %v", w, got, want)
		}
	}
}
