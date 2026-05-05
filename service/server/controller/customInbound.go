package controller

import (
	"fmt"
	"net"

	"github.com/gin-gonic/gin"
	"github.com/v2rayA/RoutingA"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
	"regexp"
	"strings"
)

func GetCustomInbound(ctx *gin.Context) {
	inbounds := configure.GetCustomInbounds()
	common.ResponseSuccess(ctx, gin.H{"inbounds": inbounds})
}

func PostCustomInbound(ctx *gin.Context) {
	var ci configure.CustomInbound
	if err := ctx.ShouldBindJSON(&ci); err != nil {
		common.ResponseError(ctx, logError("bad request"))
		return
	}
	if ci.Protocol != "socks" && ci.Protocol != "http" {
		common.ResponseError(ctx, logError(fmt.Errorf("protocol must be socks or http")))
		return
	}
	if ci.Port <= 0 || ci.Port > 65535 {
		common.ResponseError(ctx, logError(fmt.Errorf("invalid port")))
		return
	}
	if ci.Tag == "" {
		common.ResponseError(ctx, logError(fmt.Errorf("tag is required")))
		return
	}
	if net.ParseIP("0.0.0.0:"+fmt.Sprint(ci.Port)) == nil {
		// basic port check already done above
	}

	// Validate outbound binding
	if ci.Outbound == "" {
		common.ResponseError(ctx, logError(fmt.Errorf("outbound group is required")))
		return
	}
	if ci.OutboundType != "direct" && ci.OutboundType != "routingA" {
		common.ResponseError(ctx, logError(fmt.Errorf("outboundType must be 'direct' or 'routingA'")))
		return
	}

	// Verify the outbound group exists
	outbounds := configure.GetOutbounds()
	outboundExists := false
	for _, ob := range outbounds {
		if ob == ci.Outbound {
			outboundExists = true
			break
		}
	}
	if !outboundExists {
		common.ResponseError(ctx, logError(fmt.Errorf("outbound group '%s' does not exist", ci.Outbound)))
		return
	}

	// If outboundType is "routingA", validate the RoutingA rules
	if ci.OutboundType == "routingA" {
		if ci.RoutingARules == "" {
			common.ResponseError(ctx, logError(fmt.Errorf("routingA rules are required when outboundType is 'routingA'")))
			return
		}
		// Parse and validate RoutingA rules
		lines := strings.Split(ci.RoutingARules, "\n")
		hardcodeReplacement := regexp.MustCompile(`\$\$.+?\$\$`)
		for i := range lines {
			hardcodes := hardcodeReplacement.FindAllString(lines[i], -1)
			for _, hardcode := range hardcodes {
				lines[i] = strings.Replace(lines[i], hardcode, "", 1)
			}
		}
		_, err := RoutingA.Parse(strings.Join(lines, "\n"))
		if err != nil {
			common.ResponseError(ctx, logError(fmt.Errorf("invalid RoutingA rules: %w", err)))
			return
		}
	}

	inbounds := configure.GetCustomInbounds()
	// check duplicate tag and port
	for _, existing := range inbounds {
		if existing.Tag == ci.Tag {
			common.ResponseError(ctx, logError(fmt.Errorf("tag '%s' already exists", ci.Tag)))
			return
		}
		if existing.Port == ci.Port {
			common.ResponseError(ctx, logError(fmt.Errorf("port %d is already in use by '%s'", ci.Port, existing.Tag)))
			return
		}
	}
	inbounds = append(inbounds, ci)
	if err := configure.SetCustomInbounds(inbounds); err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	common.ResponseSuccess(ctx, gin.H{"inbounds": inbounds})
}

func DeleteCustomInbound(ctx *gin.Context) {
	var req struct {
		Tag string `json:"tag"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil || req.Tag == "" {
		common.ResponseError(ctx, logError("bad request"))
		return
	}
	inbounds := configure.GetCustomInbounds()
	newList := inbounds[:0]
	found := false
	for _, ci := range inbounds {
		if ci.Tag == req.Tag {
			found = true
			continue
		}
		newList = append(newList, ci)
	}
	if !found {
		common.ResponseError(ctx, logError(fmt.Errorf("tag '%s' not found", req.Tag)))
		return
	}
	if err := configure.SetCustomInbounds(newList); err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	common.ResponseSuccess(ctx, gin.H{"inbounds": newList})
}
