package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/common/resolv"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/v2ray"
	"github.com/v2rayA/v2rayA/server/service"
)

// DnsConfigResponse 是新 DNS 配置的 API 响应结构。
type DnsConfigResponse struct {
	UseNewModule bool                `json:"useNewModule"`
	Listener     string              `json:"listener"`
	Rules        []configure.DnsRule `json:"rules"`
}

type dnsSettingsRequest struct {
	Rules   *[]configure.DnsRule `json:"rules"`
	DnsMode *configure.DnsMode   `json:"dnsMode"`
	NodeDns *string              `json:"nodeDns"`
}

// PutDnsRules accepts the legacy rules array or the DNS dialog's combined
// changes, so replacing a selected source needs only one validation and reload.
func PutDnsRules(ctx *gin.Context) {
	release, ok := beginMutation(ctx)
	if !ok {
		return
	}
	defer release()
	raw, err := ctx.GetRawData()
	if err != nil {
		common.ResponseError(ctx, badRequest("DNS settings", err))
		return
	}
	raw = bytes.TrimSpace(raw)
	previousSetting := configure.GetSettingNotNil()
	data := *previousSetting
	var rules []configure.DnsRule
	combined := len(raw) > 0 && raw[0] == '{'
	writeSetting := false
	if combined {
		var input dnsSettingsRequest
		if err = json.Unmarshal(raw, &input); err == nil {
			if input.DnsMode != nil {
				data.DnsMode = *input.DnsMode
				writeSetting = true
			}
			err = configure.ValidateDnsMode(data.DnsMode)
			if data.DnsServiceEnabled() {
				if input.Rules != nil {
					rules = *input.Rules
				}
				if input.NodeDns != nil {
					data.NodeDns = *input.NodeDns
					writeSetting = true
				}
			}
		}
	} else {
		err = json.Unmarshal(raw, &rules)
	}
	if err != nil {
		common.ResponseError(ctx, badRequest("DNS settings", err))
		return
	}
	for i, rule := range rules {
		// 检查上游地址：优先使用 Upstream 字段，回退到 Server
		upstream := rule.Upstream
		if upstream == "" {
			upstream = rule.Server
		}
		if upstream == "" {
			common.ResponseError(ctx, logError(fmt.Errorf("DNS rule %d has no upstream server", i+1)))
			return
		}
		if err := v2ray.CheckDnsUpstream(upstream); err != nil {
			common.ResponseError(ctx, badRequest("DNS rules", fmt.Errorf("DNS rule %d: %w", i+1, err)))
			return
		}
		// 同步新旧字段：确保 Server 和 Upstream 至少有一个有值
		if rule.Server == "" && rule.Upstream != "" {
			rule.Server = rule.Upstream
		}
		if rule.Upstream == "" && rule.Server != "" {
			rule.Upstream = rule.Server
		}
		// 同步 Domain 和 Domains
		if rule.Domains == "" && rule.Domain != "" {
			rule.Domains = rule.Domain
		}
		if rule.Domain == "" && rule.Domains != "" {
			rule.Domain = rule.Domains
		}
		if rule.Outbound == "" {
			rule.Outbound = "direct"
		}
		rules[i] = rule
	}

	var migrated []configure.DnsRule
	if rules != nil {
		migrated = configure.MigrateDnsRules(rules)
	}
	proposed := migrated
	if proposed == nil {
		proposed = configure.GetDnsRulesNotNil()
	}
	var endpoint *resolv.IPDNSEndpoint
	if !combined || data.DnsServiceEnabled() {
		endpoint, err = v2ray.SelectNodeDNS(&data, proposed)
		if err != nil {
			common.ResponseError(ctx, logError(err))
			return
		}
		if writeSetting && data.NodeDns != "" && data.NodeDns != "auto" {
			data.NodeDns = endpoint.URL
		}
	}
	var setting *configure.Setting
	if writeSetting {
		configure.MigrateSetting(&data)
		setting = &data
	}
	err = service.ApplyCoreConfig(func() func() error {
		var previousRules []configure.DnsRule
		if migrated != nil {
			previousRules = configure.GetDnsRulesNotNil()
		}
		var restoreSetting *configure.Setting
		if writeSetting {
			restoreSetting = previousSetting
		}
		return func() error { return configure.SetDnsSettings(previousRules, restoreSetting) }
	}, func() error {
		return configure.SetDnsSettings(migrated, setting)
	}, endpoint)
	if err != nil {
		var failure *service.ApplyCoreConfigError
		if errors.As(err, &failure) {
			err = failure.UpdateErr
			if failure.RestoreStoreErr != nil {
				err = fmt.Errorf("%w; restoring DNS rules failed: %v", err, failure.RestoreStoreErr)
			} else if failure.RestoreUpdateErr != nil {
				err = fmt.Errorf("%w; restarting with previous DNS rules failed: %v", err, failure.RestoreUpdateErr)
			}
		}
		common.ResponseError(ctx, logError(err))
		return
	}
	common.ResponseSuccess(ctx, nil)
}

// GetDnsRules 处理 GET /api/dns 请求，返回 DNS 规则配置和状态。
// 返回新格式配置，同时保持向后兼容。
func GetDnsRules(ctx *gin.Context) {
	rules := configure.GetDnsRulesNotNil()
	migrated := configure.MigrateDnsRules(rules)

	// 获取当前设置以返回监听地址等信息
	setting := service.GetSetting()

	listener := ""
	if setting != nil {
		listener = setting.DnsListenAddr
	}

	common.ResponseSuccess(ctx, DnsConfigResponse{
		UseNewModule: true,
		Listener:     listener,
		Rules:        migrated,
	})
}

func GetNodeDNSOptions(ctx *gin.Context) {
	common.ResponseSuccess(ctx, v2ray.CollectNodeDNSOptions(service.GetSetting(), configure.GetDnsRulesNotNil()))
}

// PostNodeDNSOptions previews sources from unsaved rules without mutating them.
func PostNodeDNSOptions(ctx *gin.Context) {
	var input struct {
		Rules []configure.DnsRule `json:"rules"`
	}
	if err := ctx.ShouldBindJSON(&input); err != nil {
		common.ResponseError(ctx, badRequest("DNS rules", err))
		return
	}
	common.ResponseSuccess(ctx, v2ray.CollectNodeDNSOptions(service.GetSetting(), input.Rules))
}
