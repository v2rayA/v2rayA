package controller

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/v2rayA/RoutingA"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"github.com/v2rayA/v2rayA/server/service"
)

// applyCustomInbounds stores the new list and reloads the running core with it.
// A core that refuses the new list used to stay stopped with the rejected
// inbound stored, so every later start failed too; put the old list back and
// bring the core up with it instead.
func applyCustomInbounds(previous, next []configure.CustomInbound) error {
	err := service.ApplyCoreConfig(func() func() error {
		return func() error { return configure.SetCustomInbounds(previous) }
	}, func() error {
		return configure.SetCustomInbounds(next)
	})
	var failure *service.ApplyCoreConfigError
	if !errors.As(err, &failure) {
		return err
	}
	if failure.RestoreStoreErr != nil {
		log.Warn("applyCustomInbounds: failed to restore the previous inbounds: %v", failure.RestoreStoreErr)
	}
	if failure.RestoreUpdateErr != nil {
		log.Warn("applyCustomInbounds: failed to restart the core with the previous inbounds: %v", failure.RestoreUpdateErr)
	}
	if failure.Restored() {
		return fmt.Errorf("the core could not restart with the new inbound, the previous ones are back: %w", failure.UpdateErr)
	}
	return fmt.Errorf("the core could not restart with the new inbound and could not be restored either: %w", failure.UpdateErr)
}

func GetCustomInbound(ctx *gin.Context) {
	inbounds := configure.GetCustomInbounds()
	common.ResponseSuccess(ctx, gin.H{"inbounds": inbounds})
}

func PostCustomInbound(ctx *gin.Context) {
	release, ok := beginMutation(ctx)
	if !ok {
		return
	}
	defer release()

	var ci configure.CustomInbound
	if err := ctx.ShouldBindJSON(&ci); err != nil {
		common.ResponseError(ctx, badRequest("custom inbound", fmt.Errorf("request body is not a valid custom inbound object: %v", err)))
		return
	}
	if ci.Protocol != "socks" && ci.Protocol != "http" {
		common.ResponseError(ctx, common.Coded("CUSTOM_INBOUND_INVALID", logError(fmt.Errorf("protocol %q is not supported; use socks or http", ci.Protocol)), map[string]interface{}{
			"field": "protocol",
			"value": ci.Protocol,
		}))
		return
	}
	if ci.Port <= 0 || ci.Port > 65535 {
		common.ResponseError(ctx, common.Coded("CUSTOM_INBOUND_INVALID", logError(fmt.Errorf("port %d is out of range; use 1-65535", ci.Port)), map[string]interface{}{
			"field": "port",
			"value": ci.Port,
		}))
		return
	}
	ports := configure.GetPortsNotNil()
	if ci.Port == ports.Socks5 || ci.Port == ports.Http || ci.Port == ports.Socks5WithPac ||
		ci.Port == ports.HttpWithPac || ci.Port == ports.Vmess || ci.Port == ports.Api.Port {
		common.ResponseError(ctx, common.Coded("CUSTOM_INBOUND_INVALID", logError(fmt.Errorf("port %d is already in use by a configured port", ci.Port)), map[string]interface{}{
			"field": "port",
			"value": ci.Port,
		}))
		return
	}
	if ci.Tag == "" {
		common.ResponseError(ctx, logError(fmt.Errorf("tag is required")))
		return
	}
	// Proxy authentication is optional, but half of it is a configuration
	// mistake that would silently leave the port open.
	if (ci.Username == "") != (ci.Password == "") {
		common.ResponseError(ctx, common.Coded("CUSTOM_INBOUND_INVALID", logError(fmt.Errorf("inbound %q needs both a username and a password, or neither", ci.Tag)), map[string]interface{}{
			"field": "username",
		}))
		return
	}

	// Validate outbound binding
	if ci.Outbound == "" {
		common.ResponseError(ctx, logError(fmt.Errorf("inbound %q needs an outbound group", ci.Tag)))
		return
	}
	if ci.OutboundType != "direct" && ci.OutboundType != "routingA" {
		common.ResponseError(ctx, common.Coded("CUSTOM_INBOUND_INVALID", logError(fmt.Errorf("outboundType %q is not supported; use direct or routingA", ci.OutboundType)), map[string]interface{}{
			"field": "outboundType",
			"value": ci.OutboundType,
		}))
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
			common.ResponseError(ctx, logError(fmt.Errorf("inbound %q uses routingA but its RoutingA rules are empty", ci.Tag)))
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
			common.ResponseError(ctx, common.Coded("CUSTOM_INBOUND_INVALID", logError(fmt.Errorf("port %d is already in use by '%s'", ci.Port, existing.Tag)), map[string]interface{}{
				"field": "port",
				"value": ci.Port,
			}))
			return
		}
	}
	previous := append([]configure.CustomInbound(nil), inbounds...)
	inbounds = append(inbounds, ci)
	if err := applyCustomInbounds(previous, inbounds); err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	common.ResponseSuccess(ctx, gin.H{"inbounds": inbounds})
}

func DeleteCustomInbound(ctx *gin.Context) {
	release, ok := beginMutation(ctx)
	if !ok {
		return
	}
	defer release()

	var req struct {
		Tag string `json:"tag"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil || req.Tag == "" {
		common.ResponseError(ctx, badRequest("tag", "request body must be an object with a non-empty \"tag\""))
		return
	}
	inbounds := configure.GetCustomInbounds()
	previous := append([]configure.CustomInbound(nil), inbounds...)
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
	if err := applyCustomInbounds(previous, newList); err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	common.ResponseSuccess(ctx, gin.H{"inbounds": newList})
}
