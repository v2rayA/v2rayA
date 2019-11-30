package controller

import (
	"V2RayA/persistence/configure"
	"V2RayA/service"
	"V2RayA/tools"
	"errors"
	"github.com/gin-gonic/gin"
)

func PostConnection(ctx *gin.Context) {
	var which configure.Which
	err := ctx.ShouldBindJSON(&which)
	if err != nil {
		tools.ResponseError(ctx, errors.New("参数有误"))
		return
	}
	lastConnectedServer := configure.GetConnectedServerNotNil()
	err = service.Connect(&which)
	if err != nil {
		tools.ResponseError(ctx, errors.New("连接失败："+err.Error()))
		return
	}
	tools.ResponseSuccess(ctx, gin.H{"connectedServer": configure.GetConnectedServerNotNil(), "lastConnectedServer": lastConnectedServer})
}

func DeleteConnection(ctx *gin.Context) {
	cs := configure.GetConnectedServerNotNil()
	err := service.Disconnect()
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	tools.ResponseSuccess(ctx, gin.H{"lastConnectedServer": cs})
}
