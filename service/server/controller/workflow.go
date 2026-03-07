package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/server/service"
)

var refreshSubscriptionAndReselect = service.RefreshSubscriptionAndReselect
var respondWorkflowTouch = getTouch

type refreshSubscriptionAndReselectRequest struct {
	SubscriptionID int `json:"subscriptionId"`
}

func PostRefreshSubscriptionAndReselect(ctx *gin.Context) {
	updatingMu.Lock()
	if updating {
		common.ResponseError(ctx, processingErr)
		updatingMu.Unlock()
		return
	}
	updating = true
	updatingMu.Unlock()
	defer func() {
		updatingMu.Lock()
		updating = false
		updatingMu.Unlock()
	}()

	var data refreshSubscriptionAndReselectRequest
	if err := ctx.ShouldBindJSON(&data); err != nil || data.SubscriptionID <= 0 {
		common.ResponseError(ctx, logError("bad request: invalid subscriptionId"))
		return
	}
	if err := refreshSubscriptionAndReselect(data.SubscriptionID - 1); err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	respondWorkflowTouch(ctx)
}
