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
	lastConnectedServer := configure.GetConnectedServer()
	err = service.Connect(&which)
	if err != nil {
		err := errors.New("连接失败：" + err.Error())
		tools.ResponseError(ctx, err)
		return
	}
	tools.ResponseSuccess(ctx, gin.H{"connectedServer": configure.GetConnectedServer(), "lastConnectedServer": lastConnectedServer})
}

func DeleteConnection(ctx *gin.Context) {
	cs := configure.GetConnectedServer()
	err := service.Disconnect()
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	tools.ResponseSuccess(ctx, gin.H{"lastConnectedServer": cs})
}
