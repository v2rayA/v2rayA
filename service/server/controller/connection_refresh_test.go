package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func stubAutomaticRefresh(t *testing.T, refresh func(context.Context, string) error) {
	t.Helper()
	previousRefresh := refreshAutomaticGroups
	previousStart := startV2ray
	refreshAutomaticGroups = refresh
	t.Cleanup(func() {
		refreshAutomaticGroups = previousRefresh
		startV2ray = previousStart
	})
}

func TestManualStartRefreshesAutomaticGroupsBeforeCore(t *testing.T) {
	var order []string
	stubAutomaticRefresh(t, func(context.Context, string) error {
		order = append(order, "refresh")
		return nil
	})
	startV2ray = func() error {
		order = append(order, "start")
		return nil
	}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v2ray", nil)
	PostV2ray(ctx)
	if len(order) != 2 || order[0] != "refresh" || order[1] != "start" {
		t.Fatalf("manual start order = %v; want refresh then start", order)
	}
}

func TestPostOutboundRefreshRequestsNamedGroup(t *testing.T) {
	var refreshed string
	stubAutomaticRefresh(t, func(_ context.Context, name string) error {
		refreshed = name
		return nil
	})
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/outboundRefresh", strings.NewReader(`{"outbound":"proxy"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	PostOutboundRefresh(ctx)
	if refreshed != "proxy" {
		t.Fatalf("refreshed group = %q; want proxy", refreshed)
	}
}
