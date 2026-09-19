package configure

import (
	"bytes"
	"encoding/json"
	"net"
	"reflect"
	"strings"
	"testing"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
)

func TestWhichMarshalMatchesLegacyWireFormat(t *testing.T) {
	type legacyWhich struct {
		TYPE     TouchType `json:"_type"`
		ID       int       `json:"id"`
		Sub      int       `json:"sub"`
		Latency  string    `json:"pingLatency,omitempty"`
		Link     string
		Outbound string `json:"outbound"`
		Selected bool   `json:"selected,omitempty"`
	}
	tests := []Which{
		{
			NodeRef:  NodeRef{TYPE: SubscriptionServerType, ID: 3, Sub: 2, Outbound: "proxy"},
			Latency:  "27ms",
			Link:     "socks5://example.test:1080",
			Selected: true,
		},
		{NodeRef: NodeRef{TYPE: ServerType, ID: 1}},
	}
	for _, fixture := range tests {
		legacy := legacyWhich{
			TYPE: fixture.TYPE, ID: fixture.ID, Sub: fixture.Sub,
			Latency: fixture.Latency, Link: fixture.Link,
			Outbound: fixture.Outbound, Selected: fixture.Selected,
		}
		want, err := json.Marshal(legacy)
		if err != nil {
			t.Fatal(err)
		}
		got, err := json.Marshal(fixture)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("encoding/json bytes changed:\n got %s\nwant %s", got, want)
		}
		got, err = jsoniter.Marshal(fixture)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("jsoniter bytes changed:\n got %s\nwant %s", got, want)
		}
		var decoded Which
		if err := jsoniter.Unmarshal(want, &decoded); err != nil {
			t.Fatal(err)
		}
		if decoded != fixture {
			t.Fatalf("decoded Which %+v, want %+v", decoded, fixture)
		}
	}
}

func TestNodeRefsDecodeLegacyStoredTouches(t *testing.T) {
	legacy := []byte(`{"touches":[{"_type":"subscriptionServer","id":2,"sub":1,"pingLatency":"19ms","Link":"socks5://example.test:1080","outbound":"proxy","selected":true}]}`)
	var refs NodeRefs
	if err := jsoniter.Unmarshal(legacy, &refs); err != nil {
		t.Fatal(err)
	}
	want := NodeRef{TYPE: SubscriptionServerType, ID: 2, Sub: 1, Outbound: "proxy"}
	if refs.Len() != 1 || *refs.Get()[0] != want {
		t.Fatalf("decoded refs %+v, want %+v", refs.Get(), want)
	}
	got, err := jsoniter.Marshal(&refs)
	if err != nil {
		t.Fatal(err)
	}
	wantStored := []byte(`{"touches":[{"_type":"subscriptionServer","id":2,"sub":1,"outbound":"proxy"}]}`)
	if !bytes.Equal(got, wantStored) {
		t.Fatalf("stored refs changed:\n got %s\nwant %s", got, wantStored)
	}
}

func TestSortSameTypeReverse(t *testing.T) {
	ws := NewWhiches([]*Which{
		{NodeRef: NodeRef{TYPE: SubscriptionType, ID: 1}}, {NodeRef: NodeRef{TYPE: ServerType, ID: 1}},
		{NodeRef: NodeRef{TYPE: SubscriptionServerType, Sub: 0, ID: 1}}, {NodeRef: NodeRef{TYPE: SubscriptionType, ID: 2}},
		{NodeRef: NodeRef{TYPE: SubscriptionServerType, Sub: 1, ID: 1}}, {NodeRef: NodeRef{TYPE: ServerType, ID: 2}},
	})
	ws.SortSameTypeReverse()
	want := []*Which{
		{NodeRef: NodeRef{TYPE: ServerType, ID: 2}}, {NodeRef: NodeRef{TYPE: ServerType, ID: 1}},
		{NodeRef: NodeRef{TYPE: SubscriptionType, ID: 2}}, {NodeRef: NodeRef{TYPE: SubscriptionType, ID: 1}},
		{NodeRef: NodeRef{TYPE: SubscriptionServerType, Sub: 1, ID: 1}}, {NodeRef: NodeRef{TYPE: SubscriptionServerType, Sub: 0, ID: 1}},
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
	w := &Which{NodeRef: NodeRef{TYPE: ServerType, ID: index + 1}}
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
	for _, w := range []*NodeRef{
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
	for _, w := range []*NodeRef{
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
