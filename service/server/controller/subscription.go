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

/*批量更新订阅*/
func PutSubscriptions(ctx *gin.Context) {
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

	var data struct {
		IDs []int `json:"ids"`
	}
	err := ctx.ShouldBindJSON(&data)
	if err != nil || len(data.IDs) == 0 {
		common.ResponseError(ctx, logError("bad request"))
		return
	}
	indexes := make([]int, 0, len(data.IDs))
	limit := configure.GetLenSubscriptions()
	for _, id := range data.IDs {
		index := id - 1
		if id <= 0 || index < 0 || index >= limit {
			common.ResponseError(ctx, logError("bad request: ID exceed range"))
			return
		}
		indexes = append(indexes, index)
	}
	err = service.UpdateSubscriptionsInOrder(indexes)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	getTouch(ctx)
}
