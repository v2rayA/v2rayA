package controller

import (
	"V2RayA/global"
	"V2RayA/model/v2ray"
	"V2RayA/persistence/configure"
	"V2RayA/tools"
	"errors"
	"github.com/gin-gonic/gin"
)

func PostV2ray(ctx *gin.Context) {
	cs := configure.GetConnectedServer()
	if cs == nil {
		tools.ResponseError(ctx, errors.New("不能启动V2Ray, 请选择一个节点连接"))
		return
	}
	csr, err := cs.LocateServer()
	if err != nil {
		return
	}
	err = v2ray.UpdateV2RayConfig(&csr.VmessInfo)
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	tools.ResponseSuccess(ctx, gin.H{"connectedServer": cs})
}

func DeleteV2ray(ctx *gin.Context) {
	if global.ServiceControlMode == global.DockerMode {
		tools.ResponseError(ctx, errors.New("Docker模式下无法关闭V2Ray，但可以断开节点连接"))
		return
	}
	err := v2ray.StopAndDisableV2rayService()
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	global.SSRs.ClearAll()
	tools.ResponseSuccess(ctx, gin.H{"lastConnectedServer": configure.GetConnectedServer()})
}
