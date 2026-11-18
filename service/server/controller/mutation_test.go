package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/synctest"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
)

func codeOf(t *testing.T, recorder *httptest.ResponseRecorder) (common.Code, string) {
	t.Helper()
	var response struct {
		Code      common.Code `json:"code"`
		ErrorCode string      `json:"errorCode"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("%v: %s", err, recorder.Body.String())
	}
	return response.Code, response.ErrorCode
}

func TestTouchAnswersDuringAMutation(t *testing.T) {
	release, ok := beginMutation(testContext(t))
	if !ok {
		t.Fatal("free turn refused")
	}
	defer release()
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/touch", nil)
	GetTouch(ctx)
	if code, _ := codeOf(t, recorder); code != common.SUCCESS {
		t.Fatalf("touch during a mutation answered %s", recorder.Body.String())
	}
}

func TestMutationWaitsForTheTurn(t *testing.T) {
	previous := mutationWait
	mutationWait = 200 * time.Millisecond
	t.Cleanup(func() { mutationWait = previous })

	release, ok := beginMutation(testContext(t))
	if !ok {
		t.Fatal("free turn refused")
	}
	go func() {
		time.Sleep(50 * time.Millisecond)
		release()
	}()
	second, ok := beginMutation(testContext(t))
	if !ok {
		t.Fatal("a turn released within the wait was refused")
	}
	// Holding the turn past the wait answers REQUEST_IN_PROGRESS.
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/connection", nil)
	if _, ok := beginMutation(ctx); ok {
		t.Fatal("a held turn was granted twice")
	}
	if _, errorCode := codeOf(t, recorder); errorCode != "REQUEST_IN_PROGRESS" {
		t.Fatalf("held turn answered %s", recorder.Body.String())
	}
	second()
}

func TestCustomInboundMutationsWaitForTheTurn(t *testing.T) {
	previous := mutationWait
	mutationWait = 10 * time.Millisecond
	t.Cleanup(func() { mutationWait = previous })

	release, ok := beginMutation(testContext(t))
	if !ok {
		t.Fatal("free turn refused")
	}
	defer release()

	tests := []struct {
		name    string
		method  string
		handler func(*gin.Context)
	}{
		{name: "post", method: http.MethodPost, handler: PostCustomInbound},
		{name: "delete", method: http.MethodDelete, handler: DeleteCustomInbound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest(tt.method, "/customInbound", nil)
			tt.handler(ctx)
			if _, errorCode := codeOf(t, recorder); errorCode != "REQUEST_IN_PROGRESS" {
				t.Fatalf("held turn answered %s", recorder.Body.String())
			}
		})
	}
}

func TestPatchSubscriptionWaitsForTheMutationTurn(t *testing.T) {
	previous := mutationWait
	mutationWait = 10 * time.Millisecond
	t.Cleanup(func() { mutationWait = previous })
	release, ok := beginMutation(testContext(t))
	if !ok {
		t.Fatal("free turn refused")
	}
	defer release()
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPatch, "/subscription", nil)
	PatchSubscription(ctx)
	if _, errorCode := codeOf(t, recorder); errorCode != "REQUEST_IN_PROGRESS" {
		t.Fatalf("held turn answered %s", recorder.Body.String())
	}
}

func TestCoreControlCancelsQueuedHTTPTest(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		previous := mutation
		mutation = make(chan struct{}, 1)
		defer func() { mutation = previous }()
		release, err := waitForMutation(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		defer release()

		for _, cancelWithStop := range []bool{false, true} {
			ctx, cancel := context.WithCancel(t.Context())
			done := make(chan error, 1)
			go func() {
				// Parsing this URL would fail if the queued test acquires the turn.
				_, err := runHTTPTest(ctx, nil, "%")
				done <- err
			}()
			synctest.Wait()
			select {
			case <-done:
				t.Error("probe returned before cancellation")
				cancel()
				continue
			default:
			}
			if cancelWithStop {
				finish, ok := interruptHTTPTests(t.Context())
				if !ok {
					t.Fatal("core control was cancelled")
				}
				defer finish()
				if err := <-done; err != nil {
					t.Errorf("queued probe cancelled by core control: %v", err)
				}
			} else {
				cancel()
				if err := <-done; err != context.Canceled {
					t.Errorf("queued probe cancelled by client: %v", err)
				}
			}
			cancel()
		}
	})
}

func testContext(t *testing.T) *gin.Context {
	t.Helper()
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodPost, "/connection", nil)
	return ctx
}
