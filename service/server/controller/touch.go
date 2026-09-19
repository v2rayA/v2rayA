package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/touch"
	"github.com/v2rayA/v2rayA/kernel/v2ray"
	"github.com/v2rayA/v2rayA/server/service"
)

func GetTouch(ctx *gin.Context) {
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
	release, ok := beginMutation(ctx)
	if !ok {
		return
	}
	defer release()

	var ws configure.Whiches
	err := ctx.ShouldBindJSON(&ws)
	if err != nil {
		common.ResponseError(ctx, badRequest("touches", "request body must be {\"touches\": [...]} of server or subscription items"))
		return
	}
	err = service.DeleteWhich(ws.Get())
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	getTouch(ctx)
}
