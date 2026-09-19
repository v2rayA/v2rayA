package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestPostLoginBindsBodyBeforeTakingSemaphore(t *testing.T) {
	previous := loginSessions
	loginSessions = make(chan interface{}, 1)
	t.Cleanup(func() { loginSessions = previous })
	loginSessions <- nil

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/login", nil)
	ctx.Request.Header.Set("Content-Type", "application/json")
	ctx.Request.ContentLength = 10
	done := make(chan struct{})
	go func() {
		PostLogin(ctx)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		<-loginSessions
		select {
		case <-done:
		case <-time.After(time.Second):
		}
		t.Fatal("invalid body waited for the login semaphore")
	}
	if recorder.Code < 400 || recorder.Code >= 500 {
		t.Fatalf("status = %d, want 4xx", recorder.Code)
	}
	select {
	case <-loginSessions:
	default:
		t.Fatal("invalid body acquired the login semaphore")
	}
	select {
	case loginSessions <- nil:
		<-loginSessions
	default:
		t.Fatal("login semaphore is not free")
	}
}
