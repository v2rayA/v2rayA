package configure

import "github.com/v2rayA/v2rayA/db"

// CustomInbound represents a user-defined inbound proxy port.
// Protocol must be "socks" or "http".
// Tag must be unique and is used as the inbound tag in v2ray config.
// Outbound specifies the outbound group this inbound is bound to.
// OutboundType specifies the binding mode: "direct" or "routingA".
// RoutingARules contains the RoutingA rule text when OutboundType is "routingA".
type CustomInbound struct {
	Tag           string `json:"tag"`
	Protocol      string `json:"protocol"` // "socks" or "http"
	Port          int    `json:"port"`
	Outbound      string `json:"outbound"`      // bound outbound group name
	OutboundType  string `json:"outboundType"`  // "direct" or "routingA"
	RoutingARules string `json:"routingARules"` // RoutingA rules text (when outboundType="routingA")
}

// GetCustomInbounds returns all custom inbound configs stored in DB.
func GetCustomInbounds() []CustomInbound {
	var result []CustomInbound
	_ = db.Get("system", "customInbounds", &result)
	if result == nil {
		result = []CustomInbound{}
	}
	return result
}

// SetCustomInbounds persists the custom inbound list to DB.
func SetCustomInbounds(inbounds []CustomInbound) error {
	return db.Set("system", "customInbounds", inbounds)
}

// GetCustomInboundsByOutbound returns all custom inbounds bound to the given outbound group.
func GetCustomInboundsByOutbound(outbound string) []CustomInbound {
	all := GetCustomInbounds()
	var result []CustomInbound
	for _, ci := range all {
		if ci.Outbound == outbound {
			result = append(result, ci)
		}
	}
	return result
}
