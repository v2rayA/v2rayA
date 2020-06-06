package controller

import (
	"github.com/mzz2017/v2rayA/common"
	"github.com/mzz2017/v2rayA/core/v2ray"
	"github.com/mzz2017/v2rayA/global"
	"github.com/mzz2017/v2rayA/db/configure"
	"github.com/gin-gonic/gin"
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
	err := v2ray.StopAndDisableV2rayService()
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	global.Plugins.CloseAll()
	common.ResponseSuccess(ctx, gin.H{"lastConnectedServer": configure.GetConnectedServer()})
}
