package controller

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
)

func PutDnsRules(ctx *gin.Context) {
	var rules []configure.DnsRule
	if err := ctx.ShouldBindJSON(&rules); err != nil {
		common.ResponseError(ctx, logError(fmt.Errorf("bad request: %w", err)))
		return
	}
	for i, rule := range rules {
		if rule.Server == "" {
			common.ResponseError(ctx, logError(fmt.Errorf("rule[%d]: server cannot be empty", i)))
			return
		}
		if rule.Outbound == "" {
			common.ResponseError(ctx, logError(fmt.Errorf("rule[%d]: outbound cannot be empty", i)))
			return
		}
	}
	if err := configure.SetDnsRules(rules); err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	common.ResponseSuccess(ctx, nil)
}

func GetDnsRules(ctx *gin.Context) {
	common.ResponseSuccess(ctx, gin.H{
		"rules": configure.GetDnsRulesNotNil(),
	})
}
