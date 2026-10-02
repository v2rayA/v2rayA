package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func stubAutomaticRefresh(t *testing.T, refresh func(context.Context, string) error) {
	t.Helper()
	previousRefresh := refreshAutomaticGroups
	previousStart := startV2ray
	previousNeedsInitial := needsInitialGroupRefresh
	refreshAutomaticGroups = refresh
	t.Cleanup(func() {
		refreshAutomaticGroups = previousRefresh
		startV2ray = previousStart
		needsInitialGroupRefresh = previousNeedsInitial
	})
}

func TestManualStartRefreshesEmptyAutomaticGroupBeforeCore(t *testing.T) {
	var order []string
	stubAutomaticRefresh(t, func(context.Context, string) error {
		order = append(order, "refresh")
		return nil
	})
	needsInitialGroupRefresh = func() bool { return true }
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

func TestManualStartDoesNotWaitForExistingGroupRefresh(t *testing.T) {
	refreshStarted := make(chan struct{})
	releaseRefresh := make(chan struct{})
	refreshDone := make(chan struct{})
	stubAutomaticRefresh(t, func(context.Context, string) error {
		close(refreshStarted)
		<-releaseRefresh
		close(refreshDone)
		return nil
	})
	needsInitialGroupRefresh = func() bool { return false }
	started := false
	startV2ray = func() error {
		started = true
		return nil
	}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v2ray", nil)
	returned := make(chan struct{})
	go func() {
		PostV2ray(ctx)
		close(returned)
	}()
	select {
	case <-returned:
	case <-time.After(3 * time.Second):
		close(releaseRefresh)
		t.Fatal("manual start waited for automatic group refresh")
	}
	if !started {
		t.Fatal("manual start did not start the core")
	}
	select {
	case <-refreshStarted:
	case <-time.After(3 * time.Second):
		close(releaseRefresh)
		t.Fatal("manual start did not trigger automatic group refresh")
	}
	close(releaseRefresh)
	<-refreshDone
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
