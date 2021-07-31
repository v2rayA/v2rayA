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
	err = service.Connect(&which)
	if err != nil {
		log.Println(err)
		common.ResponseError(ctx, logError(err, "failed to connect"))
		return
	}
	GetTouch(ctx)
}

func DeleteConnection(ctx *gin.Context) {
	var which configure.Which
	err := ctx.ShouldBindJSON(&which)
	if err != nil {
		common.ResponseError(ctx, logError(nil, "bad request"))
		return
	}
	err = service.Disconnect(which.Outbound)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	GetTouch(ctx)
}

func PostV2ray(ctx *gin.Context) {
	err := service.StartV2ray()
	if err != nil {
		common.ResponseError(ctx, logError(err, "failed to start v2ray-core"))
		return
	}
	GetTouch(ctx)
}

func DeleteV2ray(ctx *gin.Context) {
	err := service.StopV2ray()
	if err != nil {
		common.ResponseError(ctx, logError(err, "failed to stop v2ray-core"))
		return
	}
	GetTouch(ctx)
}
