package controller

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/kernel/v2ray"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/server/service"
)

func GetOutbounds(ctx *gin.Context) {
	outbounds := configure.GetOutbounds()
	common.ResponseSuccess(ctx, gin.H{
		"outbounds": outbounds,
	})
}

func PostOutbound(ctx *gin.Context) {
	var data struct {
		Outbound string `json:"outbound"`
	}
	if err := ctx.ShouldBindJSON(&data); err != nil || data.Outbound == "" {
		common.ResponseError(ctx, logError("bad request"))
		return
	}
	if err := configure.AddOutbound(data.Outbound); err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	GetOutbounds(ctx)
}

func GetOutbound(ctx *gin.Context) {
	setting := configure.GetOutboundSetting(ctx.Query("outbound"))
	common.ResponseSuccess(ctx, gin.H{
		"setting": setting,
	})
}

func PutOutbound(ctx *gin.Context) {
	var data struct {
		Outbound string                    `json:"outbound"`
		Setting  configure.OutboundSetting `json:"setting"`
	}
	if err := ctx.ShouldBindJSON(&data); err != nil || data.Outbound == "" {
		common.ResponseError(ctx, logError("bad request"))
		return
	}
	if err := configure.SetOutboundSetting(data.Outbound, data.Setting); err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	if v2ray.ProcessManager.Running() && configure.GetConnectedServers().Len() > 0 {
		err := v2ray.UpdateV2RayConfig()
		if err != nil {
			common.ResponseError(ctx, fmt.Errorf("invalid config: %w", err))
		}
	}
	common.ResponseSuccess(ctx, nil)
}

func DeleteOutbound(ctx *gin.Context) {
	updatingMu.Lock()
	if updating {
		common.ResponseError(ctx, processingErr)
		updatingMu.Unlock()
		return
	}
	updating = true
	updatingMu.Unlock()
	defer func() {
		updatingMu.Lock()
		updating = false
		updatingMu.Unlock()
	}()

	var data struct {
		Outbound string `json:"outbound"`
	}
	if err := ctx.ShouldBindJSON(&data); err != nil || data.Outbound == "" {
		common.ResponseError(ctx, logError("bad request"))
		return
	}
	if data.Outbound == "proxy" {
		common.ResponseError(ctx, logError("outbound \"proxy\" cannot be deleted"))
		return
	}

	// Check if any custom inbound is bound to this outbound group
	boundInbounds := configure.GetCustomInboundsByOutbound(data.Outbound)
	if len(boundInbounds) > 0 {
		names := make([]string, len(boundInbounds))
		for i, ci := range boundInbounds {
			names[i] = fmt.Sprintf("%s (port %d, %s)", ci.Tag, ci.Port, ci.Protocol)
		}
		common.ResponseError(ctx, logError(fmt.Errorf(
			"cannot delete outbound group '%s': the following custom inbounds are bound to it:\n%s\nPlease unbind them first",
			data.Outbound,
			strings.Join(names, "\n"),
		)))
		return
	}

	if w := configure.GetConnectedServersByOutbound(data.Outbound); w != nil {
		if err := service.Disconnect(configure.Which{Outbound: data.Outbound}, true); err != nil {
			common.ResponseError(ctx, logError(err))
			return
		}
	}
	if err := configure.RemoveOutbound(data.Outbound); err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	GetOutbounds(ctx)
}

func PutOutboundConnections(ctx *gin.Context) {
	updatingMu.Lock()
	if updating {
		common.ResponseError(ctx, processingErr)
		updatingMu.Unlock()
		return
	}
	updating = true
	updatingMu.Unlock()
	defer func() {
		updatingMu.Lock()
		updating = false
		updatingMu.Unlock()
	}()

	var data struct {
		Outbound string `json:"outbound"`
		Touches  []struct {
			ID        int    `json:"id"`
			TYPE      string `json:"_type"`
			TYPEAlias string `json:"type"`
			Sub       *int   `json:"sub"`
			Outbound  string `json:"outbound"`
		} `json:"touches"`
	}
	if err := ctx.ShouldBindJSON(&data); err != nil {
		common.ResponseError(ctx, logError("bad request"))
		return
	}

	normalizeTouchType := func(raw string) (configure.TouchType, bool) {
		normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(raw, "_", ""), "-", ""))
		switch normalized {
		case "server":
			return configure.ServerType, true
		case "subscriptionserver":
			return configure.SubscriptionServerType, true
		default:
			return "", false
		}
	}

	whiches := make([]configure.Which, 0, len(data.Touches))
	for i, w := range data.Touches {
		rawType := w.TYPE
		if rawType == "" {
			rawType = w.TYPEAlias
		}
		typ, ok := normalizeTouchType(rawType)
		if !ok {
			common.ResponseError(ctx, logError(fmt.Errorf("bad request: invalid touch type at index %d: %q", i, rawType)))
			return
		}
		if w.ID <= 0 {
			common.ResponseError(ctx, logError(fmt.Errorf("bad request: invalid touch id at index %d: %d", i, w.ID)))
			return
		}
		sub := 0
		if w.Sub != nil {
			sub = *w.Sub
		}
		if typ == configure.SubscriptionServerType && sub < 0 {
			common.ResponseError(ctx, logError(fmt.Errorf("bad request: invalid sub index at index %d: %d", i, sub)))
			return
		}
		if typ == configure.ServerType {
			sub = 0
		}
		whiches = append(whiches, configure.Which{
			TYPE:     typ,
			ID:       w.ID,
			Sub:      sub,
			Outbound: data.Outbound,
		})
	}

	if err := service.ReplaceOutboundConnections(data.Outbound, whiches); err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	getTouch(ctx)
}
