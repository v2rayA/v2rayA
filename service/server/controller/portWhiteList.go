package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/global"
	"github.com/v2rayA/v2rayA/server/service"
	"net"
	"strconv"
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
	p := service.GetPorts()
	_, listenPort, _ := net.SplitHostPort(global.GetEnvironmentConfig().Address)
	tcpList := []string{"1:1023", data.RequestPort, listenPort}
	udpList := []string{"1:1023"}
	if p.Http > 0 {
		tcpList = append(tcpList, strconv.Itoa(p.Http))
	}
	if p.HttpWithPac > 0 {
		tcpList = append(tcpList, strconv.Itoa(p.HttpWithPac))
	}
	if p.VlessGrpc > 0 {
		tcpList = append(tcpList, strconv.Itoa(p.VlessGrpc))
	}
	if p.Socks5 > 0 {
		tcpList = append(tcpList, strconv.Itoa(p.Socks5))
		udpList = append(udpList, strconv.Itoa(p.Socks5))
	}
	wpl := configure.PortWhiteList{
		TCP: tcpList,
		UDP: udpList,
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
