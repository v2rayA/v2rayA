package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
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
	// if is turning VLESS-GRPC on
	if data.Vmess > 0 && origin.Vmess <= 0 {
		link, err := getLinkForVmess(ctx.Request.Host)
		if err != nil {
			common.ResponseError(ctx, err)
			return
		}
		common.ResponseSuccess(ctx, gin.H{
			"vmessLink": link,
		})
		return
	}
	common.ResponseSuccess(ctx, nil)
}

func getLinkForVmess(host string) (string, error) {
	return _getLinkForVmess(getHostnameFromHost(host))
}

func GetPorts(ctx *gin.Context) {
	var data struct {
		configure.Ports
		VmessLink *string `json:"vmessLink"`
	}
	data.Ports = service.GetPorts()
	if data.Vmess > 0 {
		link, err := getLinkForVmess(ctx.Request.Host)
		if err != nil {
			common.ResponseError(ctx, err)
			return
		}
		data.VmessLink = &link
	}
	common.ResponseSuccess(ctx, data)
}

func _getLinkForVmess(hostname string) (link string, err error) {
	id, err := v2ray.GenerateIdFromAccounts()
	if err != nil {
		return "", fmt.Errorf("failed to generate link for VLESS-GRPC inbound: %w", err)
	}
	p := service.GetPorts()
	info := serverObj.V2Ray{
		Ps:       "VMess | v2rayA",
		Add:      hostname,
		Port:     strconv.Itoa(p.Vmess),
		ID:       id,
		Aid:      "0",
		Net:      "tcp",
		V:        "2",
		Protocol: "vmess",
	}
	return info.ExportToURL(), nil
}
