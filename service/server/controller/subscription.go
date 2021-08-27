package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/core/touch"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/server/service"
)

/*修改Remarks*/
func PatchSubscription(ctx *gin.Context) {
	var data struct {
		Subscription touch.Subscription `json:"subscription"`
	}
	err := ctx.ShouldBindJSON(&data)
	s := data.Subscription
	index := s.ID - 1
	if err != nil || s.TYPE != configure.SubscriptionType || index < 0 || index >= configure.GetLenSubscriptions() {
		common.ResponseError(ctx, logError("bad request"))
		return
	}
	err = service.ModifySubscriptionRemark(s)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	getTouch(ctx)
}

/*更新订阅*/
func PutSubscription(ctx *gin.Context) {
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

	var data configure.Which
	err := ctx.ShouldBindJSON(&data)
	index := data.ID - 1
	if err != nil || data.TYPE != configure.SubscriptionType || index < 0 || index >= configure.GetLenSubscriptions() {
		common.ResponseError(ctx, logError("bad request: ID exceed range"))
		return
	}
	err = service.UpdateSubscription(index, false)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	getTouch(ctx)
}
