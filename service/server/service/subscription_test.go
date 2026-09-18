package service

import (
	"testing"
	"time"
)

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
