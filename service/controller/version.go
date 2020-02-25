package controller

import (
	"V2RayA/global"
	"V2RayA/model/v2ray"
	"V2RayA/service"
	"V2RayA/tools"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetVersion(ctx *gin.Context) {
	err := v2ray.CheckDohSupported()
	var dohValid string
	var iptablesMode string
	if err == nil {
		dohValid = "yes"
	} else {
		dohValid = err.Error()
	}
	if global.SupportTproxy {
		iptablesMode = "tproxy"
	} else {
		iptablesMode = "redirect"
	}
	tools.ResponseSuccess(ctx, gin.H{
		"version":       global.Version,
		"dockerMode":    global.ServiceControlMode == global.DockerMode,
		"foundNew":      global.FoundNew,
		"remoteVersion": global.RemoteVersion,
		"serviceValid":  v2ray.IsV2rayServiceValid(),
		"dohValid":      dohValid,
		"iptablesMode":  iptablesMode,
	})
}

func GetRemoteGFWListVersion(ctx *gin.Context) {
	//c, err := httpClient.GetHttpClientAutomatically()
	//if err != nil {
	//	tools.ResponseError(ctx, err)
	//	return
	//}
	t, err := service.GetRemoteGFWListUpdateTime(http.DefaultClient)
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	tools.ResponseSuccess(ctx, gin.H{"remoteGFWListVersion": t.Format("2006-01-02")})
}
