package controller

import (
	"V2RayA/global"
	"V2RayA/service"
	"V2RayA/tools"
	"github.com/gin-gonic/gin"
)

func Version(ctx *gin.Context) {
	tools.ResponseSuccess(ctx, gin.H{
		"version":       global.Version,
		"dockerMode":    global.ServiceControlMode == global.DockerMode,
		"foundNew":      global.FoundNew,
		"remoteVersion": global.RemoteVersion,
	})
}

func GetRemoteGFWListVersion(ctx *gin.Context) {
	c, err := tools.GetHttpClientAutomatically()
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	t, err := service.GetRemoteGFWListUpdateTime(c)
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	tools.ResponseSuccess(ctx, gin.H{"remoteGFWListVersion": t.Format("2006-01-02")})
}
