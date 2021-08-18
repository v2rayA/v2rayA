package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/core/vmessInfo"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/global"
	"github.com/v2rayA/v2rayA/server/service"
	"net"
	"strconv"
	"strings"
)

func getHostnameFromHost(host string) string {
	i := strings.LastIndex(host, ":")
	if _, err := strconv.Atoi(host[i+1:]); err != nil {
		return host
	}
	h, _, err := net.SplitHostPort(host)
	if err != nil {
		return host
	}
	return h
}

func PutPorts(ctx *gin.Context) {
	var data configure.Ports
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		common.ResponseError(ctx, logError(nil, "bad request"))
		return
	}
	origin := service.GetPorts()
	err = service.SetPorts(&data)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	if data.VlessGrpc > 0 && origin.VlessGrpc <= 0 {
		// turn VLESS-GRPC on
		host := "example.com"
		if h := getHostnameFromHost(ctx.Request.Host); h != "localhost" && h != "127.0.0.1" && h != "0.0.0.0" {
			host = h
		}
		link, err := genLinkForVlessGrpc(host)
		if err != nil {
			common.ResponseError(ctx, err)
			return
		}
		common.ResponseSuccess(ctx, gin.H{
			"vlessGrpcLink": link,
		})
		return
	}
	common.ResponseSuccess(ctx, nil)
}

func GetPorts(ctx *gin.Context) {
	var data struct {
		configure.Ports
		VlessGrpcLink *string `json:"vlessGrpcLink"`
	}
	data.Ports = service.GetPorts()
	if data.VlessGrpc > 0 {
		host := "example.com"
		if h := getHostnameFromHost(ctx.Request.Host); h != "localhost" && h != "127.0.0.1" && h != "0.0.0.0" {
			host = h
		}
		link, err := genLinkForVlessGrpc(host)
		if err != nil {
			common.ResponseError(ctx, err)
			return
		}
		data.VlessGrpcLink = &link
	}
	common.ResponseSuccess(ctx, data)
}

func genLinkForVlessGrpc(hostname string) (link string, err error) {
	id, err := v2ray.GenerateIdFromAccounts()
	if err != nil {
		return "", fmt.Errorf("failed to generate link for VLESS-GRPC inbound: %w", err)
	}
	cert := "/etc/v2raya/vlessGrpc.crt"
	if len(global.GetEnvironmentConfig().VlessGrpcInboundCertKey) >= 2 {
		cert = global.GetEnvironmentConfig().VlessGrpcInboundCertKey[0]
	}
	_, commonName, err := common.GetCertInfo(cert)
	if err != nil {
		return "", err
	}
	p := service.GetPorts()
	info := vmessInfo.VmessInfo{
		Ps:       hostname + " - v2rayA",
		Add:      hostname,
		Port:     strconv.Itoa(p.VlessGrpc),
		ID:       id,
		Net:      "grpc",
		Host:     commonName,
		Type:     "v2rayA_VLESS_GRPC",
		TLS:      "tls",
		Alpn:     "h2",
		V:        "2",
		Protocol: "vless",
	}
	return info.ExportToURL(), nil
}
