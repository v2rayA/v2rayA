package v2ray

import (
	"net"

	"github.com/v2rayA/v2rayA/core/coreObj"
	"github.com/v2rayA/v2rayA/core/v2ray/where"
	"github.com/v2rayA/v2rayA/db/configure"
)

// The health check traverses the active proxy outbound, even when public proxy
// ports are disabled or routing rules would send the probe destination directly.
func (t *Template) appendSubscriptionMonitor() error {
	if t.Variant != where.Xray {
		return nil
	}
	connected := configure.GetConnectedServersByOutbound("proxy").Get()
	if len(connected) != 1 || connected[0].TYPE != configure.SubscriptionServerType {
		return nil
	}
	sub := configure.GetSubscription(connected[0].Sub)
	if sub == nil || !sub.Monitor {
		return nil
	}
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	t.SubscriptionMonitorPort = l.Addr().(*net.TCPAddr).Port
	l.Close()
	const tag = "subscription-monitor"
	t.Inbounds = append(t.Inbounds, coreObj.Inbound{
		Listen: "127.0.0.1", Port: t.SubscriptionMonitorPort, Protocol: "http", Tag: tag,
	})
	t.Routing.Rules = append([]coreObj.RoutingRule{{
		Type: "field", InboundTag: []string{tag}, OutboundTag: "proxy",
	}}, t.Routing.Rules...)
	return nil
}
