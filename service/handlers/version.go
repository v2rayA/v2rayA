package handlers

import (
	"V2RayA/global"
	"V2RayA/models/v2ray"
	"V2RayA/tools"
	"github.com/gin-gonic/gin"
)

func Version(ctx *gin.Context) {
	tools.ResponseSuccess(ctx, gin.H{
		"version":    "v0.2",
		"isInDocker": global.ServiceControlMode == v2ray.Docker,
	})
}

func GetRemoteGFWListVersion(ctx *gin.Context) {
	c, err := tools.GetHttpClientAutomatically()
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	t, err := tools.GetRemoteGFWListUpdateTime(c)
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	tools.ResponseSuccess(ctx, gin.H{"remoteGFWListVersion": t.Format("2006-01-02")})
}
