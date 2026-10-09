package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
)

func putSettingBody(t *testing.T, body string) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPut, "/setting", strings.NewReader(body))
	ctx.Request.Header.Set("Content-Type", "application/json")
	PutSetting(ctx)
	return recorder
}

func storedDnsMode(t *testing.T) (configure.DnsMode, configure.DefaultYesNo) {
	t.Helper()
	setting := configure.GetSettingNotNil()
	return setting.DnsMode, setting.DnsHijack
}

func resetSetting(t *testing.T) {
	t.Helper()
	t.Cleanup(func() { _ = configure.SetSetting(configure.NewSetting()) })
	// A saved setting restarts the auto-update tickers, which only the
	// service's own start creates.
	if conf.TickerUpdateGFWList == nil {
		conf.TickerUpdateGFWList = time.NewTicker(time.Hour)
		t.Cleanup(conf.TickerUpdateGFWList.Stop)
	}
	if conf.TickerUpdateSubscription == nil {
		conf.TickerUpdateSubscription = time.NewTicker(time.Hour)
		t.Cleanup(conf.TickerUpdateSubscription.Stop)
	}
}

// putSetting expects body to be accepted and reports what the store then
// holds; a rejected body fails the test with the answer it produced.
func putSetting(t *testing.T, body string) (configure.DnsMode, configure.DefaultYesNo) {
	t.Helper()
	recorder := putSettingBody(t, body)
	if code, _ := codeOf(t, recorder); code != common.SUCCESS {
		t.Fatalf("request %s answered %s", body, recorder.Body.String())
	}
	setting := configure.GetSettingNotNil()
	return setting.DnsMode, setting.DnsHijack
}

// hijackFor is the opt-out the migration writes for a mode, so a test can
// seed a store the way a previous save would have left it.
func hijackFor(mode configure.DnsMode) configure.DefaultYesNo {
	if mode == configure.DnsModeHijack {
		return configure.Yes
	}
	return configure.No
}

// The mode is written the way it is chosen, and the opt-out is rewritten to
// match so a client that only knows dnsHijack reads the same decision.
func TestPutSettingStoresTheDnsMode(t *testing.T) {
	resetSetting(t)
	for _, mode := range []configure.DnsMode{
		configure.DnsModeOff,
		configure.DnsModeService,
		configure.DnsModeHijack,
	} {
		body, err := json.Marshal(map[string]configure.DnsMode{"dnsMode": mode})
		if err != nil {
			t.Fatal(err)
		}
		stored, hijack := putSetting(t, string(body))
		if stored != mode {
			t.Fatalf("stored dnsMode = %q, want %q", stored, mode)
		}
		if hijack != hijackFor(mode) {
			t.Fatalf("dnsMode %s left dnsHijack = %q, want %q", mode, hijack, hijackFor(mode))
		}
	}
}

// A mode this build does not know is refused rather than resolved: a typo
// must not leave the store in a state where the only safe reading turns
// interception off behind the user's back.
func TestPutSettingRejectsAnUnknownDnsMode(t *testing.T) {
	resetSetting(t)
	for _, mode := range []configure.DnsMode{
		configure.DnsModeOff,
		configure.DnsModeService,
		configure.DnsModeHijack,
	} {
		if err := configure.SetSetting(&configure.Setting{DnsMode: mode, DnsHijack: hijackFor(mode)}); err != nil {
			t.Fatal(err)
		}
		recorder := putSettingBody(t, `{"dnsMode":"hijackk"}`)
		if code, errorCode := codeOf(t, recorder); code != common.FAIL || errorCode != "BAD_REQUEST" {
			t.Fatalf("unknown mode answered %s", recorder.Body.String())
		}
		if stored, hijack := storedDnsMode(t); stored != mode || hijack != hijackFor(mode) {
			t.Fatalf("a rejected request changed the store to %q/%q, want %q/%q", stored, hijack, mode, hijackFor(mode))
		}
	}
}

// The opt-out mirrors the mode, so a stored mode of service carries a "no"
// that reads as "off". A request that never spoke about DNS must therefore
// not resolve the stored mode through that mirror: saving an unrelated page
// would quietly turn the DNS module off.
func TestPutSettingWithoutEitherDnsFieldKeepsTheStoredMode(t *testing.T) {
	resetSetting(t)
	for _, mode := range []configure.DnsMode{
		configure.DnsModeOff,
		configure.DnsModeService,
		configure.DnsModeHijack,
	} {
		if err := configure.SetSetting(&configure.Setting{DnsMode: mode, DnsHijack: hijackFor(mode)}); err != nil {
			t.Fatal(err)
		}
		// A settings page that has nothing to do with DNS.
		stored, hijack := putSetting(t, `{"logLevel":"debug"}`)
		if stored != mode || hijack != hijackFor(mode) {
			t.Fatalf("an unrelated save stored %q/%q, want the stored %q/%q", stored, hijack, mode, hijackFor(mode))
		}
		if got := configure.GetSettingNotNil().LogLevel; got != "debug" {
			t.Fatalf("the unrelated field was not applied: logLevel = %q", got)
		}
	}
}

// A client that predates the mode sends only the opt-out, and the opt-out it
// moved has to take effect even though a mode is already in the database.
func TestPutSettingFollowsAnExplicitLegacyOptOut(t *testing.T) {
	resetSetting(t)
	for _, tc := range []struct {
		name    string
		stored  configure.DnsMode
		body    string
		want    configure.DnsMode
		wantHij configure.DefaultYesNo
	}{
		{"opt-out from hijack", configure.DnsModeHijack, `{"dnsHijack":"no"}`, configure.DnsModeOff, configure.No},
		{"opt-in from off", configure.DnsModeOff, `{"dnsHijack":"yes"}`, configure.DnsModeHijack, configure.Yes},
		{"opt-in from service", configure.DnsModeService, `{"dnsHijack":"yes"}`, configure.DnsModeHijack, configure.Yes},
		{"opt-out from service", configure.DnsModeService, `{"dnsHijack":"no"}`, configure.DnsModeOff, configure.No},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := configure.SetSetting(&configure.Setting{DnsMode: tc.stored, DnsHijack: hijackFor(tc.stored)}); err != nil {
				t.Fatal(err)
			}
			if stored, hijack := putSetting(t, tc.body); stored != tc.want || hijack != tc.wantHij {
				t.Fatalf("%s stored %q/%q, want %q/%q", tc.body, stored, hijack, tc.want, tc.wantHij)
			}
		})
	}
}

// The mode is what a client that knows it is asking for, so it stands over
// both the mode already stored and the opt-out in the same body.
func TestPutSettingTreatsAnExplicitModeAsAuthoritative(t *testing.T) {
	resetSetting(t)
	for _, tc := range []struct {
		name    string
		stored  configure.DnsMode
		body    string
		want    configure.DnsMode
		wantHij configure.DefaultYesNo
	}{
		{"service over a stored off", configure.DnsModeOff, `{"dnsMode":"service"}`, configure.DnsModeService, configure.No},
		{"off over a stored service", configure.DnsModeService, `{"dnsMode":"off"}`, configure.DnsModeOff, configure.No},
		{"mode over a contradicting opt-out", configure.DnsModeHijack, `{"dnsMode":"service","dnsHijack":"yes"}`, configure.DnsModeService, configure.No},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := configure.SetSetting(&configure.Setting{DnsMode: tc.stored, DnsHijack: hijackFor(tc.stored)}); err != nil {
				t.Fatal(err)
			}
			if stored, hijack := putSetting(t, tc.body); stored != tc.want || hijack != tc.wantHij {
				t.Fatalf("%s stored %q/%q, want %q/%q", tc.body, stored, hijack, tc.want, tc.wantHij)
			}
		})
	}
}
