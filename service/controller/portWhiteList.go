package controller

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
	"v2rayA/common"
	"v2rayA/core/v2ray"
	"v2rayA/global"
	"v2rayA/db/configure"
	"v2rayA/service"
)

func GetPortWhiteList(ctx *gin.Context) {
	common.ResponseSuccess(ctx, configure.GetPortWhiteListNotNil())
}

func PutPortWhiteList(ctx *gin.Context) {
	var data configure.PortWhiteList
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		common.ResponseError(ctx, logError(nil, "bad request"))
		return
	}
	if !data.Valid() {
		common.ResponseError(ctx, logError(nil, "invalid format of port"))
		return
	}
	err = configure.SetPortWhiteList(&data)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	common.ResponseSuccess(ctx, nil)
}

func PostPortWhiteList(ctx *gin.Context) {
	var data struct {
		RequestPort string `json:"requestPort"`
	}
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		common.ResponseError(ctx, logError(nil, "bad request"))
		return
	}
	p := service.GetPortsDefault()
	listenPort := strings.Split(global.GetEnvironmentConfig().Address, ":")[1]
	wpl := configure.PortWhiteList{
		TCP: []string{"1:1023", data.RequestPort, listenPort, strconv.Itoa(p.Http), strconv.Itoa(p.HttpWithPac), strconv.Itoa(p.Socks5)},
		UDP: []string{"1:1023", strconv.Itoa(p.Socks5)},
	}
	err = configure.SetPortWhiteList(&wpl)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	v2ray.CheckAndStopTransparentProxy()
	v2ray.CheckAndSetupTransparentProxy(true)
	common.ResponseSuccess(ctx, nil)
}
