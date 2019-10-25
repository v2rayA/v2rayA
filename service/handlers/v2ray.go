package handlers

import (
	"V2RayA/config"
	"V2RayA/tools"
	"errors"
	"github.com/gin-gonic/gin"
)

func PostV2ray(ctx *gin.Context) {
	tr := config.GetTouchRaw()
	if tr.ConnectedServer == nil {
		tools.ResponseError(ctx, errors.New("不能启动V2Ray, 请选择一个节点连接"))
		return
	}
	err := tools.RestartV2rayService()
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	tools.ResponseSuccess(ctx, gin.H{"connectedServer": tr.ConnectedServer})
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
	tools.ResponseSuccess(ctx, gin.H{"lastConnectedServer": config.GetTouchRaw().ConnectedServer})
}
