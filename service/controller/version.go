package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/core/v2ray/where"
	"net/http"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/core/v2ray/asset/gfwlist"
	"github.com/v2rayA/v2rayA/global"
)

func GetVersion(ctx *gin.Context) {
	var dohValid string
	var vlessValid bool
	var iptablesMode string

	ver, err := where.GetV2rayServiceVersion()
	if err == nil {
		vlessValid, _ = common.VersionGreaterEqual(ver, "4.27.0")
		err = v2ray.CheckDohSupported(ver)
	}
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
		"vlessValid":    vlessValid,
		"iptablesMode":  iptablesMode, //仅代表是否支持tproxy，真实iptables所使用的表还要看是否是增强模式
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
