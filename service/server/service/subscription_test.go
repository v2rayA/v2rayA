package service

import (
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/touch"
)

func TestModifyMissingSubscriptionIncludesID(t *testing.T) {
	resetSubscription(t)
	err := ModifySubscriptionRemark(touch.Subscription{ID: 42})
	var coded *common.CodedError
	if !errors.As(err, &coded) || coded.Code != "SUBSCRIPTION_NOT_FOUND" || coded.Params["id"] != 42 {
		t.Fatalf("error = %#v, want SUBSCRIPTION_NOT_FOUND with id=42", err)
	}
}

func TestSubscriptionBodyOverLimitIsRejected(t *testing.T) {
	client := &http.Client{Transport: updateTransport(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(strings.NewReader(strings.Repeat("x", int(maxSubscriptionDocumentSize+1)))),
			Header:     make(http.Header),
		}, nil
	})}
	if _, _, err := ResolveSubscriptionWithClient("https://example.test/subscription", client); err == nil || !strings.Contains(err.Error(), "32 MiB") {
		t.Fatalf("error = %v, want size limit", err)
	}
}

func TestSubscriptionCoreApplyErrorWrapsCause(t *testing.T) {
	cause := errors.New("reload failed")
	err := subscriptionCoreApplyError(cause)
	if !errors.Is(err, cause) || !strings.HasPrefix(err.Error(), "subscription stored, but the core could not apply it:") {
		t.Fatalf("error = %v", err)
	}
}

func TestParseProfileTitle(t *testing.T) {
	for _, tc := range []struct{ header, want string }{
		{"", ""},
		{"  Example Network  ", "Example Network"},
		{"base64:" + base64.StdEncoding.EncodeToString([]byte("Сеть 🌐")), "Сеть 🌐"},
		{"base64:" + base64.RawStdEncoding.EncodeToString([]byte("日本語!")), "日本語!"},
		{"base64:???", ""},
		{"base64:/w==", ""},
		{"base64:YQpi", ""},
	} {
		if got := parseProfileTitle(tc.header); got != tc.want {
			t.Errorf("parseProfileTitle(%q) = %q, want %q", tc.header, got, tc.want)
		}
	}
}

func TestSubscriptionProfileTitleLifecycle(t *testing.T) {
	const address = "https://example.test/sub"
	title, body := "Example", "http-proxy://127.0.0.1:1234#node"
	client := &http.Client{Transport: updateTransport(func(*http.Request) (*http.Response, error) {
		header := make(http.Header)
		header.Set("Profile-Title", title)
		return &http.Response{
			StatusCode: http.StatusOK, Header: header,
			Body: io.NopCloser(strings.NewReader(body)),
		}, nil
	})}
	oldClient := directSubscriptionClient
	directSubscriptionClient = func() *http.Client { return client }
	t.Cleanup(func() { directSubscriptionClient = oldClient })
	options := SubscriptionFetchOptions{BypassProxy: true}
	index := configure.GetLenSubscriptions()
	if err := ImportSubscriptionWithOptions(address, options); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = configure.RemoveSubscriptions([]int{index}) })
	check := func(wantTitle, wantRemark string) {
		t.Helper()
		sub := touch.GenerateTouch().Subscriptions[index]
		if sub.ProfileTitle != wantTitle || sub.Remarks != wantRemark {
			t.Fatalf("title %q, remark %q; want %q, %q", sub.ProfileTitle, sub.Remarks, wantTitle, wantRemark)
		}
	}
	check("Example", "")
	item := touch.Subscription{ID: index + 1, Address: address, Remarks: "Local"}
	if err := ModifySubscriptionRemark(item); err != nil {
		t.Fatal(err)
	}
	title = "base64:" + base64.StdEncoding.EncodeToString([]byte("Новая сеть"))
	if err := UpdateSubscriptionWithOptions(index, false, options); err != nil {
		t.Fatal(err)
	}
	check("Новая сеть", "Local")
	title, body = "", ""
	if err := UpdateSubscriptionWithOptions(index, false, options); err == nil {
		t.Fatal("empty subscription accepted")
	}
	check("Новая сеть", "Local")
	body = "http-proxy://127.0.0.1:1234#node"
	if err := UpdateSubscriptionWithOptions(index, false, options); err != nil {
		t.Fatal(err)
	}
	check("", "Local")
}

func TestSubscriptionUserInfoString(t *testing.T) {
	const gib = 1024 * 1024 * 1024
	cases := []struct {
		name string
		sui  SubscriptionUserInfo
		want string
	}{
		{"usage and expiry", SubscriptionUserInfo{Upload: 0, Download: 1041529159, Total: 100 * gib, Expire: time.Date(2026, 10, 4, 18, 59, 0, 0, time.UTC)},
			"Used 0.97 GiB / 100.00 GiB · Expires " + time.Date(2026, 10, 4, 18, 59, 0, 0, time.UTC).Local().Format("2006-01-02")},
		{"no total", SubscriptionUserInfo{Upload: gib, Download: gib, Total: -1}, "Used 2.00 GiB"},
		{"total only", SubscriptionUserInfo{Upload: -1, Download: -1, Total: 10 * gib}, "Total 10.00 GiB"},
		{"nothing", SubscriptionUserInfo{Upload: -1, Download: -1, Total: -1}, ""},
	}
	for _, c := range cases {
		if got := c.sui.String(); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
		if c.sui.Known() != (c.want != "") {
			t.Errorf("%s: Known() disagrees with the rendered text", c.name)
		}
	}
}

func TestParseSubscriptionUserInfoHeader(t *testing.T) {
	sui := parseSubscriptionUserInfo("upload=0; download=1041529159; total=107374182400; expire=1791140340")
	if sui.Upload != 0 || sui.Download != 1041529159 || sui.Total != 107374182400 {
		t.Fatalf("fields: %+v", sui)
	}
	if sui.Expire.UTC().Format("2006-01-02 15:04") != "2026-10-04 18:59" {
		t.Fatalf("expire: %v", sui.Expire)
	}
	if !sui.Known() {
		t.Fatal("a header with fields must count as known")
	}
	empty := parseSubscriptionUserInfo("")
	if empty.Known() {
		t.Fatal("an empty header must not count as known")
	}
}

func TestModifySubscriptionNodeDNS(t *testing.T) {
	index := configure.GetLenSubscriptions()
	if err := configure.AppendSubscriptions([]*configure.SubscriptionRaw{{
		Address: "https://example.test/sub",
		NodeDns: "tcp://1.1.1.1:53",
	}}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = configure.RemoveSubscriptions([]int{index}) })
	item := touch.Subscription{ID: index + 1, Address: "https://example.test/sub", Remarks: "old client"}
	if err := ModifySubscriptionRemark(item); err != nil {
		t.Fatal(err)
	}
	if got := configure.GetSubscription(index).NodeDns; got != "tcp://1.1.1.1:53" {
		t.Fatalf("old client cleared node DNS: %q", got)
	}
	item.NodeDns = "tls://dns.example"
	if err := ModifySubscriptionRemark(item); err == nil {
		t.Fatal("accepted a domain node DNS")
	}
	if got := configure.GetSubscription(index).NodeDns; got != "tcp://1.1.1.1:53" {
		t.Fatalf("rejected node DNS was persisted: %q", got)
	}
	item.NodeDns = "[2001:db8::1]:5353"
	if err := ModifySubscriptionRemark(item); err != nil {
		t.Fatal(err)
	}
	if got := configure.GetSubscription(index).NodeDns; got != "udp://[2001:db8::1]:5353" {
		t.Fatalf("normalized node DNS = %q", got)
	}
	item.NodeDns = "auto"
	if err := ModifySubscriptionRemark(item); err != nil {
		t.Fatal(err)
	}
	if got := touch.GenerateTouch().Subscriptions[index].NodeDns; got != "auto" {
		t.Fatalf("touch node DNS = %q", got)
	}
}

func TestModifySubscriptionPreservesUnchangedInvalidNodeDNS(t *testing.T) {
	for _, endpoint := range []string{"invalid", "udp://127.0.0.1:52353"} {
		t.Run(endpoint, func(t *testing.T) {
			index := configure.GetLenSubscriptions()
			if err := configure.AppendSubscriptions([]*configure.SubscriptionRaw{{
				Address: "https://example.test/sub", NodeDns: endpoint,
				UpdateMode:            configure.SubscriptionUpdateIntervalFailsafe,
				UpdateIntervalMinutes: 15, FailureIntervalMinutes: 2, AllowDirectRecovery: true,
			}}); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = configure.RemoveSubscriptions([]int{index}) })
			before := configure.GetSubscription(index)
			item := touch.Subscription{ID: index + 1, Address: before.Address, Remarks: "renamed", NodeDns: endpoint}
			if err := ModifySubscriptionRemark(item); err != nil {
				t.Fatal(err)
			}
			got := configure.GetSubscription(index)
			if got.Remarks != "renamed" || got.NodeDns != endpoint || got.DatabaseID != before.DatabaseID || got.UpdateMode != before.UpdateMode || got.UpdateIntervalMinutes != 15 || got.FailureIntervalMinutes != 2 || !got.AllowDirectRecovery {
				t.Fatalf("settings not preserved: %+v", got)
			}
			item.NodeDns = "different-invalid"
			if err := ModifySubscriptionRemark(item); err == nil {
				t.Fatal("accepted a changed invalid endpoint")
			}
			if got := configure.GetSubscription(index); got.NodeDns != endpoint {
				t.Fatalf("changed invalid endpoint persisted: %+v", got)
			}
		})
	}
}
