package controller

import (
	"v2rayA/common"
	"v2rayA/core/v2ray"
	"v2rayA/core/v2ray/asset/gfwlist"
	"v2rayA/global"
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
