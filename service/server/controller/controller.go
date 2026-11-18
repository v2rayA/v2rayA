package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/server/service"
)

var (
	// mutation serialises the requests that rewrite the store or restart
	// the core. Reads do not take it: a touch during a latency test is a
	// database read and needs no turn.
	mutation      = make(chan struct{}, 1)
	mutationWait  = 5 * time.Second
	processingErr = common.Coded("REQUEST_IN_PROGRESS", fmt.Errorf("the last request is being processed"), nil)
)

// beginMutation waits for the turn to change state. When it returns false
// the request has been answered with REQUEST_IN_PROGRESS, or the client
// went away.
func beginMutation(ctx *gin.Context) (release func(), ok bool) {
	release, err := waitForMutation(ctx.Request.Context())
	if err == processingErr {
		common.ResponseError(ctx, err)
	}
	return release, err == nil
}

func waitForMutation(ctx context.Context) (func(), error) {
	timer := time.NewTimer(mutationWait)
	defer timer.Stop()
	select {
	case mutation <- struct{}{}:
		if err := ctx.Err(); err != nil {
			<-mutation
			return nil, err
		}
		service.CancelAutomation()
		for !service.ConfigurationMu.TryLock() {
			select {
			case <-timer.C:
				<-mutation
				return nil, processingErr
			case <-ctx.Done():
				<-mutation
				return nil, ctx.Err()
			case <-time.After(10 * time.Millisecond):
			}
		}
		return func() {
			service.ConfigurationMu.Unlock()
			service.NotifyAutomation()
			<-mutation
		}, nil
	case <-timer.C:
		return nil, processingErr
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
