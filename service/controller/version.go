package controller

import (
	"V2RayA/global"
	"V2RayA/core/gfwlist"
	"V2RayA/core/v2ray"
	"V2RayA/common"
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
	common.ResponseSuccess(ctx, gin.H{
		"version":       global.Version,
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
	g, err := gfwlist.GetRemoteGFWListUpdateTime(http.DefaultClient)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	common.ResponseSuccess(ctx, gin.H{"remoteGFWListVersion": g.UpdateTime.Local().Format("2006-01-02")})
}
