package controller

import (
	"V2RayA/global"
	"V2RayA/core/v2ray"
	"V2RayA/persistence/configure"
	"V2RayA/common"
	"errors"
	"github.com/gin-gonic/gin"
)

func PostV2ray(ctx *gin.Context) {
	cs := configure.GetConnectedServer()
	if cs == nil {
		common.ResponseError(ctx, errors.New("cannot start V2Ray without server connected"))
		return
	}
	csr, err := cs.LocateServer()
	if err != nil {
		return
	}
	err = v2ray.UpdateV2RayConfig(&csr.VmessInfo)
	if err != nil {
		common.ResponseError(ctx, err)
		return
	}
	common.ResponseSuccess(ctx, gin.H{"connectedServer": cs})
}

func DeleteV2ray(ctx *gin.Context) {
	err := v2ray.StopAndDisableV2rayService()
	if err != nil {
		common.ResponseError(ctx, err)
		return
	}
	global.SSRs.ClearAll()
	common.ResponseSuccess(ctx, gin.H{"lastConnectedServer": configure.GetConnectedServer()})
}
