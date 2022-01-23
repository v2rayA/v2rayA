package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/core/v2ray/asset/dat"
	"github.com/v2rayA/v2rayA/core/v2ray/service"
	"net/http"
)

func GetVersion(ctx *gin.Context) {
	var lite int
	if conf.GetEnvironmentConfig().Lite {
		lite = 1
	}
	v5 := service.CheckV5() == nil
	common.ResponseSuccess(ctx, gin.H{
		"version":       conf.Version,
		"foundNew":      conf.FoundNew,
		"remoteVersion": conf.RemoteVersion,
		"serviceValid":  service.IsV2rayServiceValid(),
		"v5":            v5,
		"lite":          lite,
	})
}

func GetRemoteGFWListVersion(ctx *gin.Context) {
	//c, err := httpClient.GetHttpClientAutomatically()
	//if err != nil {
	//	tools.ResponseError(ctx, err)
	//	return
	//}
	g, err := dat.GetRemoteGFWListUpdateTime(http.DefaultClient)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	common.ResponseSuccess(ctx, gin.H{"remoteGFWListVersion": g.UpdateTime.Local().Format("2006-01-02")})
}
