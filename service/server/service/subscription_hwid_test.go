package service

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/touch"
)

func TestSubscriptionHWIDExceptions(t *testing.T) {
	resetSubscription(t)
	base := configure.GetLenSubscriptions()
	const address = "https://a.test/sub"
	if err := configure.AppendSubscriptions([]*configure.SubscriptionRaw{{Address: address}, {Address: address}}); err != nil {
		t.Fatal(err)
	}
	first, second := configure.GetSubscription(base), configure.GetSubscription(base+1)
	if first.HWID == "" || second.HWID == "" || first.HWID == second.HWID {
		t.Fatal("two records with the same URL must get distinct random IDs")
	}
	enabled := true
	edit := touch.Subscription{ID: base + 1, DatabaseID: first.DatabaseID, Address: address, SendHWID: &enabled}
	if err := ModifySubscriptionRemark(edit); err != nil {
		t.Fatal(err)
	}
	if !configure.GetSubscription(base).SendHWID || configure.GetSubscription(base+1).SendHWID {
		t.Fatal("consent crossed subscription records")
	}
	edit.Address = "https://b.test/sub"
	if err := ModifySubscriptionRemark(edit); err != nil {
		t.Fatal(err)
	}
	if configure.GetSubscription(base).SendHWID {
		t.Fatal("address change kept consent")
	}
	setting := configure.GetSettingNotNil()
	if err := configure.SetSettingWithHWID(setting, map[int64]string{first.DatabaseID: address, second.DatabaseID: address}); err != nil {
		t.Fatal(err)
	}
	if configure.GetSubscription(base).SendHWID || !configure.GetSubscription(base+1).SendHWID {
		t.Fatal("retained a changed address or lost a matching subscription")
	}
	if err := configure.SetSettingWithHWID(setting, map[int64]string{}); err != nil {
		t.Fatal(err)
	}
	if after := configure.GetSubscription(base + 1); after.SendHWID || after.HWID != second.HWID {
		t.Fatal("revocation must stop sending without changing the identifier")
	}
	if err := configure.RemoveSubscriptions([]int{base + 1}); err != nil {
		t.Fatal(err)
	}
	if err := configure.AppendSubscriptions([]*configure.SubscriptionRaw{{Address: address}}); err != nil {
		t.Fatal(err)
	}
	if after := configure.GetSubscription(base + 1); after.SendHWID || after.HWID == second.HWID {
		t.Fatal("re-import reused the old permission or identifier")
	}
}

func TestSubscriptionHWIDConsentAndRedirect(t *testing.T) {
	original := configure.GetSettingNotNil()
	t.Cleanup(func() { _ = configure.SetSetting(original) })
	setting := *original
	seen := map[string]string{}
	client := &http.Client{Transport: updateTransport(func(req *http.Request) (*http.Response, error) {
		seen[req.URL.String()] = req.Header.Get("X-HWID")
		res := &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(""))}
		if req.URL.Path == "/redirect" {
			res.StatusCode = 302
			res.Header.Set("Location", "https://b.test/received")
		}
		return res, nil
	})}
	id := newSubscriptionHWID()
	fetch := func(url string) string {
		t.Helper()
		delete(seen, url)
		// An empty document can fail parsing after its headers have been observed.
		_, _, _ = resolveSubscriptionWithContext(context.Background(), &configure.SubscriptionRaw{Address: url, HWID: id}, client)
		value, ok := seen[url]
		if !ok {
			t.Fatalf("missing request to %s", url)
		}
		return value
	}
	setting.SendHWID = false
	if err := configure.SetSetting(&setting); err != nil {
		t.Fatal(err)
	}
	if fetch("https://a.test/sub?xhwid=true") != "" {
		t.Fatal("URL enabled HWID")
	}
	setting.SendHWID = true
	if err := configure.SetSetting(&setting); err != nil {
		t.Fatal(err)
	}
	if id == "" || fetch("https://a.test/sub") != id || fetch("https://a.test:443/second") != id {
		t.Fatal("expected the identifier of this subscription")
	}
	if fetch("http://a.test/sub") != "" {
		t.Fatal("sent HWID over HTTP")
	}
	fetch("https://a.test/redirect")
	if value, ok := seen["https://b.test/received"]; !ok || value != "" {
		t.Fatal("redirect leaked HWID or did not complete")
	}
}
