package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/core/serverObj"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/db/configure"
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
		common.ResponseError(ctx, logError("bad request"))
		return
	}
	origin := service.GetPorts()
	err = service.SetPorts(&data)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	if certKey := conf.GetEnvironmentConfig().VlessGrpcInboundCertKey; len(certKey) >= 2 {
		// if is turning VLESS-GRPC on
		if data.VlessGrpc > 0 && origin.VlessGrpc <= 0 {
			link, err := getLinkForVlessGrpc(ctx.Request.Host, certKey[0])
			if err != nil {
				common.ResponseError(ctx, err)
				return
			}
			common.ResponseSuccess(ctx, gin.H{
				"vlessGrpcLink": link,
			})
			return
		}
	}
	common.ResponseSuccess(ctx, nil)
}

func getLinkForVlessGrpc(host string, cert string) (string, error) {
	hostname := "example.com"
	if h := getHostnameFromHost(host); h != "localhost" && h != "127.0.0.1" && h != "0.0.0.0" {
		hostname = h
	}
	names, err := common.GetCertInfo(cert)
	if err != nil {
		return "", err
	}
	link, err := _getLinkForVlessGrpc(hostname, names[0])
	if err != nil {
		return "", err
	}
	return link, nil
}

func GetPorts(ctx *gin.Context) {
	var data struct {
		configure.Ports
		VlessGrpcLink *string `json:"vlessGrpcLink"`
	}
	data.Ports = service.GetPorts()
	if certKey := conf.GetEnvironmentConfig().VlessGrpcInboundCertKey; len(certKey) >= 2 {
		if data.VlessGrpc > 0 {
			link, err := getLinkForVlessGrpc(ctx.Request.Host, certKey[0])
			if err != nil {
				common.ResponseError(ctx, err)
				return
			}
			data.VlessGrpcLink = &link
		}
	}
	common.ResponseSuccess(ctx, data)
}

func _getLinkForVlessGrpc(hostname string, sni string) (link string, err error) {
	id, err := v2ray.GenerateIdFromAccounts()
	if err != nil {
		return "", fmt.Errorf("failed to generate link for VLESS-GRPC inbound: %w", err)
	}
	p := service.GetPorts()
	info := serverObj.V2Ray{
		Ps:       "VLESS-GRPC | v2rayA",
		Add:      hostname,
		Port:     strconv.Itoa(p.VlessGrpc),
		ID:       id,
		Net:      "grpc",
		Host:     sni,
		Path:     "v2rayA_VLESS_GRPC",
		TLS:      "tls",
		Alpn:     "h2",
		V:        "2",
		Protocol: "vless",
	}
	return info.ExportToURL(), nil
}
