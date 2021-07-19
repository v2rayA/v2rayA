package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/server/service"
	"log"
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
		log.Println(err)
		common.ResponseError(ctx, logError(err, "failed to connect"))
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
