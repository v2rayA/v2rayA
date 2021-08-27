package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/server/service"
)

func GetOutbounds(ctx *gin.Context) {
	outbounds := configure.GetOutbounds()
	common.ResponseSuccess(ctx, gin.H{
		"outbounds": outbounds,
	})
}

func PostOutbound(ctx *gin.Context) {
	var data struct {
		Outbound string `json:"outbound"`
	}
	if err := ctx.ShouldBindJSON(&data); err != nil || data.Outbound == "" {
		common.ResponseError(ctx, logError("bad request"))
		return
	}
	if err := configure.AddOutbound(data.Outbound); err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	GetOutbounds(ctx)
}

func DeleteOutbound(ctx *gin.Context) {
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
		Outbound string `json:"outbound"`
	}
	if err := ctx.ShouldBindJSON(&data); err != nil || data.Outbound == "" {
		common.ResponseError(ctx, logError("bad request"))
		return
	}
	if w := configure.GetConnectedServersByOutbound(data.Outbound); w != nil {
		if err := service.Disconnect(configure.Which{Outbound: data.Outbound}, true); err != nil {
			common.ResponseError(ctx, logError(err))
			return
		}
	}
	if err := configure.RemoveOutbound(data.Outbound); err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	GetOutbounds(ctx)
}
