package controller

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/v2ray"
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
		common.ResponseError(ctx, badRequest("outbound", "request body must be a JSON object with a non-empty \"outbound\" string"))
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
		common.ResponseError(ctx, badRequest("outbound", "request body must be a JSON object with a non-empty \"outbound\" string"))
		return
	}
	if err := configure.SetOutboundSetting(data.Outbound, data.Setting); err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	if v2ray.ProcessManager.Running() && configure.GetConnectedServers().Len() > 0 {
		err := v2ray.UpdateV2RayConfig()
		if err != nil {
			invalidConfigErr := fmt.Errorf("invalid config: %w", err)
			common.ResponseError(ctx, common.Coded("INVALID_CONFIG", invalidConfigErr, map[string]interface{}{"detail": err.Error()}))
			return
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
		common.ResponseError(ctx, badRequest("outbound", "request body must be a JSON object with a non-empty \"outbound\" string"))
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
		common.ResponseError(ctx, badRequest("outbound connections", "request body must be {\"outbound\": string, \"touches\": [...]}"))
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
			common.ResponseError(ctx, logError(fmt.Errorf("touches[%d]._type %q is not valid; expected server or subscriptionServer", i, rawType)))
			return
		}
		if w.ID <= 0 {
			common.ResponseError(ctx, logError(fmt.Errorf("touches[%d].id %d must be positive", i, w.ID)))
			return
		}
		sub := 0
		if w.Sub != nil {
			sub = *w.Sub
		}
		if typ == configure.SubscriptionServerType && sub < 0 {
			common.ResponseError(ctx, logError(fmt.Errorf("touches[%d].sub %d must be non-negative", i, sub)))
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

// PutOutboundSelection chooses the member a group routes through alone, or
// returns the group to balancing when `which` is null. The member must be
// connected in that group.
func PutOutboundSelection(ctx *gin.Context) {
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
		Outbound string           `json:"outbound"`
		Which    *configure.Which `json:"which"`
	}
	if err := ctx.ShouldBindJSON(&data); err != nil || data.Outbound == "" {
		common.ResponseError(ctx, badRequest("outbound selection", "request body must be {\"outbound\": string, \"which\": {...} | null}"))
		return
	}
	link := ""
	if data.Which != nil {
		data.Which.Outbound = data.Outbound
		if data.Which.TYPE == configure.ServerType {
			data.Which.Sub = 0
		}
		member := false
		if members := configure.GetConnectedServersByOutbound(data.Outbound); members != nil {
			for _, m := range members.Get() {
				if m.EqualTo(*data.Which) {
					member = true
					break
				}
			}
		}
		if !member {
			common.ResponseError(ctx, logError(fmt.Errorf("the node is not a member of outbound %q", data.Outbound)))
			return
		}
		sr, err := data.Which.LocateServerRaw()
		if err != nil {
			common.ResponseError(ctx, logError(err))
			return
		}
		link = sr.ServerObj.ExportToURL()
	}
	setting := configure.GetOutboundSetting(data.Outbound)
	setting.Selected = link
	if err := configure.SetOutboundSetting(data.Outbound, setting); err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	if v2ray.ProcessManager.Running() && configure.GetConnectedServers().Len() > 0 {
		if err := v2ray.UpdateV2RayConfig(); err != nil {
			common.ResponseError(ctx, common.Coded("INVALID_CONFIG", fmt.Errorf("invalid config: %w", err), map[string]interface{}{"detail": err.Error()}))
			return
		}
	}
	getTouch(ctx)
}
