package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/kernel/touch"
	"github.com/v2rayA/v2rayA/kernel/v2ray"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/server/service"
)

func GetTouch(ctx *gin.Context) {
	updatingMu.Lock()
	if updating {
		common.ResponseError(ctx, processingErr)
		updatingMu.Unlock()
		return
	}
	updatingMu.Unlock()
	defer func() {
		updatingMu.Lock()
		updatingMu.Unlock()
	}()
	getTouch(ctx)

}
func getTouch(ctx *gin.Context) {
	running := v2ray.ProcessManager.ServiceRunning()
	networkPaused := v2ray.ProcessManager.NetworkPaused()
	t := touch.GenerateTouch()
	common.ResponseSuccess(ctx, gin.H{
		"running":       running,
		"networkPaused": networkPaused,
		"touch":         t,
	})
}

func DeleteTouch(ctx *gin.Context) {
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

	var ws configure.Whiches
	err := ctx.ShouldBindJSON(&ws)
	if err != nil {
		common.ResponseError(ctx, logError("bad request"))
		return
	}
	err = service.DeleteWhich(ws.Get())
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	getTouch(ctx)
}
