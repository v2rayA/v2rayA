package handlers

import (
	"V2RayA/config"
	"V2RayA/tools"
	"github.com/gin-gonic/gin"
)

func PostV2ray(ctx *gin.Context) {
	err := tools.RestartV2rayService()
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	tools.ResponseSuccess(ctx, gin.H{"connectedServer": config.GetTouchRaw().ConnectedServer})
}

func DeleteV2ray(ctx *gin.Context) {
	err := tools.StopV2rayService()
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	err = tools.DisableV2rayService()
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	tools.ResponseSuccess(ctx, gin.H{"connectedServer": config.GetTouchRaw().ConnectedServer})
}
