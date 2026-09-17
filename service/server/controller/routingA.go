package controller

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/v2rayA/RoutingA"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
)

func GetRoutingA(ctx *gin.Context) {
	common.ResponseSuccess(ctx, gin.H{
		"routingA": configure.GetRoutingA(),
	})
}
func PutRoutingA(ctx *gin.Context) {
	var data struct {
		RoutingA string `json:"routingA"`
	}
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		common.ResponseError(ctx, badRequest("routingA", "request body must be {\"routingA\": string}"))
		return
	}
	// remove hardcode replacement and try parsing
	lines := strings.Split(data.RoutingA, "\n")
	hardcodeReplacement := regexp.MustCompile(`\$\$.+?\$\$`)
	for i := range lines {
		hardcodes := hardcodeReplacement.FindAllString(lines[i], -1)
		for _, hardcode := range hardcodes {
			lines[i] = strings.Replace(lines[i], hardcode, "", 1)
		}
	}
	_, err = RoutingA.Parse(strings.Join(lines, "\n"))
	if err != nil {
		common.ResponseError(ctx, logError(fmt.Errorf("invalid RoutingA rules: %w", err)))
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

	err = configure.SetRoutingA(&data.RoutingA)
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
