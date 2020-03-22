package controller

import (
	"V2RayA/core/touch"
	"V2RayA/core/v2ray"
	"V2RayA/persistence/configure"
	"V2RayA/service"
	"V2RayA/common"
	"errors"
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
		common.ResponseError(ctx, errors.New("bad request"))
		return
	}
	err = service.DeleteWhich(ws.Get())
	if err != nil {
		common.ResponseError(ctx, err)
		return
	}
	GetTouch(ctx)
}
