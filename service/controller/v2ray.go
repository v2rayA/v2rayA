package controller

import (
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/plugin"
)

func PostV2ray(ctx *gin.Context) {
	cs := configure.GetConnectedServer()
	if cs == nil {
		common.ResponseError(ctx, logError(nil, "cannot start V2Ray without server connected"))
		return
	}
	csr, err := cs.LocateServer()
	if err != nil {
		return
	}
	err = v2ray.UpdateV2RayConfig(&csr.VmessInfo)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	common.ResponseSuccess(ctx, gin.H{"connectedServer": cs})
}

func DeleteV2ray(ctx *gin.Context) {
	err := v2ray.StopV2rayService()
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	plugin.GlobalPlugins.CloseAll()
	common.ResponseSuccess(ctx, gin.H{"lastConnectedServer": configure.GetConnectedServer()})
}
