package controller

import (
	"v2rayA/common"
	"v2rayA/persistence/configure"
	"v2rayA/service"
	"github.com/gin-gonic/gin"
)

func PostConnection(ctx *gin.Context) {
	var which configure.Which
	err := ctx.ShouldBindJSON(&which)
	if err != nil {
		common.ResponseError(ctx, logError(nil, "bad request"))
		return
	}
	lastConnectedServer := configure.GetConnectedServer()
	err = service.Connect(&which)
	if err != nil {
		common.ResponseError(ctx, logError(err, "fail in connecting"))
		return
	}
	common.ResponseSuccess(ctx, gin.H{"connectedServer": configure.GetConnectedServer(), "lastConnectedServer": lastConnectedServer})
}

func DeleteConnection(ctx *gin.Context) {
	cs := configure.GetConnectedServer()
	err := service.Disconnect()
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	common.ResponseSuccess(ctx, gin.H{"lastConnectedServer": cs})
}
