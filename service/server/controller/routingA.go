package controller

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/server/service"
)

func GetRoutingA(ctx *gin.Context) {
	common.ResponseSuccess(ctx, gin.H{
		"routingA": configure.GetRoutingA(),
		"source":   configure.GetRoutingASource(),
	})
}
func PostRoutingAImport(ctx *gin.Context) {
	var source configure.RoutingASource
	if err := ctx.ShouldBindJSON(&source); err != nil {
		common.ResponseError(ctx, badRequest("source", "invalid RoutingA source"))
		return
	}
	text, err := service.FetchRoutingA(source)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	common.ResponseSuccess(ctx, gin.H{"routingA": text})
}
func PutRoutingA(ctx *gin.Context) {
	release, ok := beginMutation(ctx)
	if !ok {
		return
	}
	defer release()
	var data struct {
		RoutingA string                   `json:"routingA"`
		Source   configure.RoutingASource `json:"source"`
	}
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		common.ResponseError(ctx, badRequest("routingA", "request body must be {\"routingA\": string}"))
		return
	}
	if data.Source.URL != "" && (data.Source.IntervalHours < 1 || data.Source.IntervalHours > 8760) {
		common.ResponseError(ctx, badRequest("intervalHours", "RoutingA update interval must be between 1 and 8760 hours"))
		return
	}
	if data.Source.URL != "" {
		if err := service.ValidateRoutingASource(data.Source); err != nil {
			common.ResponseError(ctx, badRequest("source", err))
			return
		}
		data.Source.URL = strings.TrimSpace(data.Source.URL)
	}
	if data.Source.URL == "" {
		data.Source = configure.RoutingASource{IntervalHours: 24}
	}
	lines := strings.Split(data.RoutingA, "\n")
	if err := service.ValidateRoutingA(data.RoutingA); err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}

	// Check for deprecated inbound definitions in RoutingA rules
	hasInboundDef := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "inbound(") || strings.HasPrefix(trimmed, "inbound (") {
			hasInboundDef = true
			break
		}
	}

	err = service.SaveRoutingA(data.RoutingA, data.Source)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}

	if hasInboundDef {
		common.ResponseSuccess(ctx, gin.H{
			"warning": "RoutingA 中定义入站(inbound)的功能已弃用，生成的 JSON 配置将不会包含对应的入站端口。请使用自定义入站设置中的 RoutingA 规则功能替代。",
		})
		return
	}
	common.ResponseSuccess(ctx, nil)
}
