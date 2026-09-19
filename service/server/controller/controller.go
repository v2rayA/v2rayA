package controller

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
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
	timer := time.NewTimer(mutationWait)
	defer timer.Stop()
	select {
	case mutation <- struct{}{}:
		return func() { <-mutation }, true
	case <-timer.C:
		common.ResponseError(ctx, processingErr)
		return nil, false
	case <-ctx.Request.Context().Done():
		return nil, false
	}
}
