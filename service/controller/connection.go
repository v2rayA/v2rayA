package controller

import (
	"V2RayA/persistence/configure"
	"V2RayA/service"
	"V2RayA/common"
	"errors"
	"github.com/gin-gonic/gin"
)

func PostConnection(ctx *gin.Context) {
	var which configure.Which
	err := ctx.ShouldBindJSON(&which)
	if err != nil {
		common.ResponseError(ctx, errors.New("bad request"))
		return
	}
	lastConnectedServer := configure.GetConnectedServer()
	err = service.Connect(&which)
	if err != nil {
		err := errors.New("Fail in connecting: " + err.Error())
		common.ResponseError(ctx, err)
		return
	}
	common.ResponseSuccess(ctx, gin.H{"connectedServer": configure.GetConnectedServer(), "lastConnectedServer": lastConnectedServer})
}

func DeleteConnection(ctx *gin.Context) {
	cs := configure.GetConnectedServer()
	err := service.Disconnect()
	if err != nil {
		common.ResponseError(ctx, err)
		return
	}
	common.ResponseSuccess(ctx, gin.H{"lastConnectedServer": cs})
}
