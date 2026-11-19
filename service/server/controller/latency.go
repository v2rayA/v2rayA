package controller

import (
	"context"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/server/service"
)

var httpTests struct {
	sync.Mutex
	batch   context.Context
	cancel  context.CancelFunc
	active  chan struct{}
	control int
}

func interruptHTTPTests(ctx context.Context) (func(), bool) {
	httpTests.Lock()
	httpTests.control++
	if httpTests.cancel != nil {
		httpTests.cancel()
		httpTests.batch, httpTests.cancel = nil, nil
	}
	active := httpTests.active
	httpTests.Unlock()

	finish := func() {
		httpTests.Lock()
		httpTests.control--
		httpTests.Unlock()
	}
	if active != nil {
		// Only unrelated mutations should consume the normal wait budget.
		select {
		case <-active:
		case <-ctx.Done():
			finish()
			return nil, false
		}
	}
	return finish, true
}

func GetNodeLatency(ctx *gin.Context) {
	getNodeLatency(ctx, true)
}

func GetPingLatency(ctx *gin.Context) {
	getNodeLatency(ctx, false)
}

func getNodeLatency(ctx *gin.Context, configured bool) {
	var wt []*configure.Which
	err := jsoniter.Unmarshal([]byte(ctx.Query("whiches")), &wt)
	if err != nil {
		common.ResponseError(ctx, badRequest("whiches", "query parameter \"whiches\" must be a JSON array of server items"))
		return
	}
	setting := configure.GetSettingNotNil()
	if configured && setting.NodeProbeMode == "proxy" {
		wt, err = runHTTPTest(ctx.Request.Context(), wt, setting.NodeProbeURL)
	} else {
		release, ok := beginMutation(ctx)
		if !ok {
			return
		}
		defer release()
		wt, err = service.Ping(wt, time.Second)
	}
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	common.ResponseSuccess(ctx, gin.H{
		"whiches": wt,
	})
}

func GetHttpLatency(ctx *gin.Context) {
	var wt []*configure.Which
	err := jsoniter.Unmarshal([]byte(ctx.Query("whiches")), &wt)
	if err != nil {
		common.ResponseError(ctx, badRequest("whiches", "query parameter \"whiches\" must be a JSON array of server items"))
		return
	}

	request := ctx.Request.Context()
	wt, err = runHTTPTest(request, wt, ctx.Query("testUrl"))
	if request.Err() != nil {
		return
	}
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	common.ResponseSuccess(ctx, gin.H{
		"whiches": wt,
	})
}

func runHTTPTest(request context.Context, whiches []*configure.Which, testURL string) ([]*configure.Which, error) {
	httpTests.Lock()
	if httpTests.control != 0 {
		httpTests.Unlock()
		return []*configure.Which{}, nil
	}
	if httpTests.batch == nil {
		httpTests.batch, httpTests.cancel = context.WithCancel(context.Background())
	}
	batch := httpTests.batch
	httpTests.Unlock()

	probe, cancel := context.WithCancel(request)
	defer cancel()
	stop := context.AfterFunc(batch, cancel)
	defer stop()
	release, err := waitForMutation(probe)
	if err != nil {
		if batch.Err() != nil {
			return []*configure.Which{}, nil
		}
		return nil, err
	}
	defer release()

	httpTests.Lock()
	if batch.Err() != nil {
		httpTests.Unlock()
		return []*configure.Which{}, nil
	}
	httpTests.active = make(chan struct{})
	httpTests.Unlock()
	defer func() {
		httpTests.Lock()
		close(httpTests.active)
		httpTests.active = nil
		httpTests.Unlock()
	}()

	results, err := service.TestHttpLatency(probe, whiches, 8*time.Second, 32, false, testURL)
	if err != nil && err == probe.Err() {
		if results == nil {
			results = []*configure.Which{}
		}
		return results, nil
	}
	return results, err
}
