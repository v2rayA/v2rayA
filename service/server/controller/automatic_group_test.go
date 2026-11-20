package controller

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/server/service"
)

func TestAutomaticGroupSettingRoundTripAndLegacyClient(t *testing.T) {
	const group = "catalog-api-test"
	if err := configure.AddOutbound(group); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = configure.RemoveOutbound(group) })
	index := configure.GetLenServers()
	if err := service.ImportServer("socks5://192.0.2.1:1#offline", nil); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = configure.RemoveServers([]int{index}) })
	for _, flag := range []string{`,"autoAdd":true`, ``, `,"autoAdd":false`} {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(http.MethodPut, "/outbound", strings.NewReader(`{"outbound":"`+group+`","setting":{"probeURL":"https://example.com/","probeInterval":"60s","type":"leastping"`+flag+`}}`))
		ctx.Request.Header.Set("Content-Type", "application/json")
		PutOutbound(ctx)
		var response struct{ Code string }
		if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		if response.Code != "SUCCESS" {
			t.Fatal(recorder.Body.String())
		}
		if got := configure.GetOutboundSetting(group).AutoAdd; got != (flag != `,"autoAdd":false`) {
			t.Fatalf("autoAdd lost: %s", flag)
		}
	}
	if configure.GetConnectedServersByOutbound(group).Len() == 0 {
		t.Fatal("disabling automation discarded membership")
	}
}
