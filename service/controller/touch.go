package controller

import (
	"V2RayA/model/touch"
	"V2RayA/model/v2ray"
	"V2RayA/persistence/configure"
	"V2RayA/service"
	"V2RayA/tools"
	"errors"
	"github.com/gin-gonic/gin"
)

func GetTouch(ctx *gin.Context) {
	running := v2ray.IsV2RayRunning()
	t := touch.GenerateTouch()
	if !running { //如果没有运行，把connectedServer删掉，防止前端错误渲染
		t.ConnectedServer = nil
	}
	tools.ResponseSuccess(ctx, gin.H{
		"running": running,
		"touch":   t,
	})
}

func DeleteTouch(ctx *gin.Context) {
	var ws configure.Whiches
	err := ctx.ShouldBindJSON(&ws)
	if err != nil {
		tools.ResponseError(ctx, errors.New("参数有误"))
		return
	}
	err = service.DeleteWhich(ws.Get())
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	GetTouch(ctx)
}
