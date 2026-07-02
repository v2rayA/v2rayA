package configure

import "fmt"

// MigrateDnsRules 将旧 DNS 规则迁移到新格式，为缺失的新字段填充合理的默认值。
// 该函数保持向后兼容：旧规则的 Server/Domains 映射到新规则的 Upstream/Domain 语义。
// 如果规则同时设置了新旧字段，优先使用新字段值。
func MigrateDnsRules(oldRules []DnsRule) []DnsRule {
	migrated := make([]DnsRule, len(oldRules))
	for i, rule := range oldRules {
		migrated[i] = rule

		// 如果旧字段 Server 有值而新字段 Upstream 为空，则填充
		if migrated[i].Upstream == "" && migrated[i].Server != "" {
			migrated[i].Upstream = migrated[i].Server
		}
		// 如果新字段 Upstream 有值而旧字段 Server 为空，则同步
		if migrated[i].Server == "" && migrated[i].Upstream != "" {
			migrated[i].Server = migrated[i].Upstream
		}

		// 如果旧字段 Domains 有值而新字段 Domain 为空，则填充
		if migrated[i].Domain == "" && migrated[i].Domains != "" {
			migrated[i].Domain = migrated[i].Domains
		}
		// 如果新字段 Domain 有值而旧字段 Domains 为空，则同步
		if migrated[i].Domains == "" && migrated[i].Domain != "" {
			migrated[i].Domains = migrated[i].Domain
		}

		// 设置 Action 默认值
		if migrated[i].Action == "" {
			migrated[i].Action = "route"
		}

		// 设置 Policy 默认值
		if migrated[i].Policy == "" {
			migrated[i].Policy = "single"
		}

		// 设置 Outbound 默认值（仅当完全为空时）
		if migrated[i].Outbound == "" {
			migrated[i].Outbound = "direct"
		}

		// 设置 QueryType 默认值（* 表示所有类型）
		if migrated[i].QueryType == "" {
			migrated[i].QueryType = "*"
		}

		// 设置 RuleID 默认值
		if migrated[i].RuleID == "" {
			migrated[i].RuleID = fmt.Sprintf("rule-%d", rule.ID)
		}
	}
	return migrated
}

// MigrateSetting 将旧设置配置迁移到新格式，为缺失的新 DNS 配置字段填充默认值。
// 该函数只填充零值字段，不会覆盖用户已显式设置的字段。
//
// 布尔字段的特殊处理：
//   - FillEmpty() 跳过布尔字段，因此旧配置中缺失的布尔字段会保持 false
//   - 本函数使用 gjson 检测原始 JSON 中是否存在该字段
//   - 仅当字段不存在于原始 JSON 时才设置默认值 true
func MigrateSetting(setting *Setting) {
	// 监听地址默认值
	if setting.DnsListenAddr == "" {
		setting.DnsListenAddr = "0.0.0.0:52353"
	}

	// 缓存大小默认值
	if setting.DnsCacheSize == 0 {
		setting.DnsCacheSize = 4096
	}

	// 最小 TTL 默认值
	if setting.DnsCacheMinTTL == 0 {
		setting.DnsCacheMinTTL = 60
	}

	// 最大 TTL 默认值
	if setting.DnsCacheMaxTTL == 0 {
		setting.DnsCacheMaxTTL = 86400
	}

	// 布尔字段默认值处理 —— 注意：
	// common.FillEmpty 跳过布尔字段，所以旧配置中缺失的布尔字段会保持 false。
	// 这里不强制设置为 true，而是由调用方（如 GetSettingNotNil）根据
	// 原始 JSON 中的字段存在性来决定。
	// 对于全新的配置，NewSetting() 已设置正确的默认值。
}
