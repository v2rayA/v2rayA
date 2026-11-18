package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
)

func TestPutOutboundStrategySettings(t *testing.T) {
	previous := configure.GetOutboundSetting("proxy")
	t.Cleanup(func() { _ = configure.SetOutboundSetting("proxy", previous) })
	for _, tc := range []struct {
		name, strategy, url, interval string
		valid                         bool
	}{
		{"leastping", "leastping", "https://example.com", "60s", true},
		{"leastload", "leastload", "https://example.com", "60s", true},
		{"roundrobin", "roundrobin", "https://example.com", "60s", true},
		{"random", "random", "https://example.com", "60s", true},
		{"unknown strategy", "unknown", "https://example.com", "60s", false},
		{"empty strategy", "", "https://example.com", "60s", false},
		{"unsupported scheme", "random", "file:///probe", "60s", false},
		{"missing host", "random", "https:///probe", "60s", false},
		{"invalid URL", "random", "http://%", "60s", false},
		{"invalid interval", "random", "https://example.com", "invalid", false},
		{"short interval", "random", "https://example.com", "0s", false},
		{"long interval", "random", "https://example.com", "9000h", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			before := configure.GetOutboundSetting("proxy")
			setting := configure.OutboundSetting{
				Type: configure.ObservatoryType(tc.strategy), ProbeURL: tc.url, ProbeInterval: tc.interval,
			}
			body, err := json.Marshal(gin.H{"outbound": "proxy", "setting": setting})
			if err != nil {
				t.Fatal(err)
			}
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(http.MethodPut, "/outbound", bytes.NewReader(body))
			ctx.Request.Header.Set("Content-Type", "application/json")
			PutOutbound(ctx)
			var response struct{ Code common.Code }
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			wantCode, wantSetting := common.Code(common.FAIL), before
			if tc.valid {
				wantCode, wantSetting = common.SUCCESS, setting
			}
			if response.Code != wantCode {
				t.Fatalf("unexpected response: %s", recorder.Body.String())
			}
			if got := configure.GetOutboundSetting("proxy"); got != wantSetting {
				t.Fatalf("stored setting = %+v, want %+v", got, wantSetting)
			}
		})
	}
}
