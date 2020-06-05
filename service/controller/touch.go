package controller

import (
	"v2rayA/common"
	"v2rayA/core/touch"
	"v2rayA/core/v2ray"
	"v2rayA/db/configure"
	"v2rayA/service"
	"github.com/gin-gonic/gin"
)

func GetTouch(ctx *gin.Context) {
	running := v2ray.IsV2RayRunning()
	t := touch.GenerateTouch()
	if !running { //如果没有运行，把connectedServer删掉，防止前端错误渲染
		t.ConnectedServer = nil
	}
	common.ResponseSuccess(ctx, gin.H{
		"running": running,
		"touch":   t,
	})
}

func DeleteTouch(ctx *gin.Context) {
	var ws configure.Whiches
	err := ctx.ShouldBindJSON(&ws)
	if err != nil {
		common.ResponseError(ctx, logError(nil, "bad request"))
		return
	}
	err = service.DeleteWhich(ws.Get())
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	GetTouch(ctx)
}
