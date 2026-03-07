package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPostRefreshSubscriptionAndReselectBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	updating = false

	req := httptest.NewRequest(http.MethodPost, "/api/workflow/refresh-subscription-and-reselect", bytes.NewBufferString(`{"subscriptionId":0}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = req

	PostRefreshSubscriptionAndReselect(ctx)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: %d", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body["code"] != "FAIL" {
		t.Fatalf("unexpected response code: %#v", body["code"])
	}
}

func TestPostRefreshSubscriptionAndReselectSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	updating = false

	original := refreshSubscriptionAndReselect
	t.Cleanup(func() {
		refreshSubscriptionAndReselect = original
		updating = false
	})

	calledWith := -1
	refreshSubscriptionAndReselect = func(index int) error {
		calledWith = index
		return nil
	}

	req := httptest.NewRequest(http.MethodPost, "/api/workflow/refresh-subscription-and-reselect", bytes.NewBufferString(`{"subscriptionId":3}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = req

	PostRefreshSubscriptionAndReselect(ctx)

	if calledWith != 2 {
		t.Fatalf("expected service call with index 2, got %d", calledWith)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status code: %d", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if body["code"] != "SUCCESS" {
		t.Fatalf("unexpected response code: %#v", body["code"])
	}
}
