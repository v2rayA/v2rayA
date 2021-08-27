package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/server/service"
)

func GetSetting(ctx *gin.Context) {
	s := service.GetSetting()
	var localGFWListVersion string
	t, err := asset.GetGFWListModTime()
	if err == nil {
		localGFWListVersion = t.Local().Format("2006-01-02")
	}
	common.ResponseSuccess(ctx, gin.H{
		"setting":             s,
		"localGFWListVersion": localGFWListVersion,
	})
}

func PutSetting(ctx *gin.Context) {
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

	var data configure.Setting
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		common.ResponseError(ctx, logError("bad request"))
		return
	}
	if data.MuxOn == configure.Yes && (data.Mux < 1 || data.Mux > 1024) {
		common.ResponseError(ctx, logError("mux should be between 1 and 1024"))
		return
	}
	err = service.UpdateSetting(&data)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		_ = service.StopV2ray()
		return
	}
	common.ResponseSuccess(ctx, nil)
}
