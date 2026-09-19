package reqCache

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
)

func TestRequestLocksReleased(t *testing.T) {
	r := gin.New()
	r.Use(ReqCache)
	r.GET("/", func(c *gin.Context) { common.ResponseSuccess(c, nil) })
	for i := range 1000 {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set(common.RequestIdHeader, fmt.Sprintf("release-%d", i))
		r.ServeHTTP(httptest.NewRecorder(), req)
	}
	if n := len(glen.reqIdMu); n > 256 {
		t.Fatalf("retained %d request locks", n)
	}
}

func TestConcurrentRequestReplay(t *testing.T) {
	var calls atomic.Int32
	r := gin.New()
	r.Use(ReqCache)
	r.GET("/", func(c *gin.Context) {
		calls.Add(1)
		time.Sleep(10 * time.Millisecond)
		common.ResponseSuccess(c, "once")
	})
	var wg sync.WaitGroup
	for range 32 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set(common.RequestIdHeader, "concurrent-replay")
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Errorf("status %d", rec.Code)
			}
		}()
	}
	wg.Wait()
	if n := calls.Load(); n != 1 {
		t.Fatalf("handler ran %d times", n)
	}
}
