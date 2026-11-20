package service

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/v2rayA/v2rayA/db/configure"
)

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
