package v2ray

import (
	"fmt"
	"strings"
	"time"

	"github.com/v2rayA/v2rayA/common/cmds"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/core/iptables"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

// cleanupResidualTransparentProxyRules cleans up any residual iptables/nftables rules
// that may have been left behind after an abnormal termination (e.g., kill -9, system crash, panic).
// It uses "2>/dev/null || true" to ensure no errors are raised if rules/chains don't exist.
func cleanupResidualTransparentProxyRules() {
	commands := `
iptables -w 2 -t mangle -F TP_OUT 2>/dev/null || true
iptables -w 2 -t mangle -D OUTPUT -j TP_OUT 2>/dev/null || true
iptables -w 2 -t mangle -X TP_OUT 2>/dev/null || true
iptables -w 2 -t mangle -F TP_PRE 2>/dev/null || true
iptables -w 2 -t mangle -D PREROUTING -j TP_PRE 2>/dev/null || true
iptables -w 2 -t mangle -X TP_PRE 2>/dev/null || true
iptables -w 2 -t mangle -F TP_RULE 2>/dev/null || true
iptables -w 2 -t mangle -X TP_RULE 2>/dev/null || true
iptables -w 2 -t mangle -F TP_MARK 2>/dev/null || true
iptables -w 2 -t mangle -X TP_MARK 2>/dev/null || true
iptables -w 2 -t nat -F TP_OUT 2>/dev/null || true
iptables -w 2 -t nat -D OUTPUT -j TP_OUT 2>/dev/null || true
iptables -w 2 -t nat -X TP_OUT 2>/dev/null || true
iptables -w 2 -t nat -F TP_PRE 2>/dev/null || true
iptables -w 2 -t nat -D PREROUTING -j TP_PRE 2>/dev/null || true
iptables -w 2 -t nat -X TP_PRE 2>/dev/null || true
iptables -w 2 -t nat -F TP_RULE 2>/dev/null || true
iptables -w 2 -t nat -X TP_RULE 2>/dev/null || true
iptables -w 2 -F DROP_SPOOFING 2>/dev/null || true
iptables -w 2 -D INPUT -j DROP_SPOOFING 2>/dev/null || true
iptables -w 2 -D FORWARD -j DROP_SPOOFING 2>/dev/null || true
iptables -w 2 -X DROP_SPOOFING 2>/dev/null || true
ip rule del fwmark 0x40/0xc0 table 100 2>/dev/null || true
ip route del local 0.0.0.0/0 dev lo table 100 2>/dev/null || true
nft delete table inet v2raya 2>/dev/null || true
`
	cmds.ExecCommands(commands, false)
}

func deleteTransparentProxyRulesKeepSystemProxy() {
	stopTinyTun()
	iptables.CloseWatcher()
	if !conf.GetEnvironmentConfig().Lite {
		removeResolvHijacker()
		iptables.Tproxy.GetCleanCommands().Run(false)
		iptables.Redirect.GetCleanCommands().Run(false)
		iptables.DropSpoofing.GetCleanCommands().Run(false)
	}
	time.Sleep(30 * time.Millisecond)
}

func deleteTransparentProxyRules() {
	deleteTransparentProxyRulesKeepSystemProxy()
	iptables.SystemProxy.GetCleanCommands().Run(false)
}

func writeTransparentProxyRules(tmpl *Template) (err error) {
	defer func() {
		if err != nil {
			log.Warn("writeTransparentProxyRules: %v", err)
			deleteTransparentProxyRules()
		}
	}()
	cleanupResidualTransparentProxyRules()
	setting := configure.GetSettingNotNil()
	switch setting.TransparentType {
	case configure.TransparentTun:
		return startTinyTun(tmpl)
	case configure.TransparentTproxy:
		if err = iptables.Tproxy.GetSetupCommands().Run(true); err != nil {
			if strings.Contains(err.Error(), "TPROXY") && strings.Contains(err.Error(), "No chain") {
				err = fmt.Errorf("you does not compile xt_TPROXY in kernel")
			}
			return fmt.Errorf("not support \"tproxy\" mode of transparent proxy: %w", err)
		}
		iptables.SetWatcher(iptables.Tproxy)
	case configure.TransparentRedirect:
		if err = iptables.Redirect.GetSetupCommands().Run(true); err != nil {
			return fmt.Errorf("not support \"redirect\" mode of transparent proxy: %w", err)
		}
		iptables.SetWatcher(iptables.Redirect)
	case configure.TransparentSystemProxy:
		if err = iptables.SystemProxy.GetSetupCommands().Run(true); err != nil {
			return fmt.Errorf("not support \"system proxy\" mode of transparent proxy: %w", err)
		}
	default:
		return fmt.Errorf("undefined \"%v\" mode of transparent proxy", setting.TransparentType)
	}

	if ShouldLocalDnsListen() {
		if couldListenLocalhost, e := CouldLocalDnsListen(); couldListenLocalhost {
			if e != nil {
				log.Warn("only listen at 127.2.0.17: %v", e)
			}
			resetResolvHijacker()
		} else {
			log.Warn("writeTransparentProxyRules: %v", e)
		}
	}
	return nil
}

func IsTransparentOn(setting *configure.Setting) bool {
	if setting == nil {
		setting = configure.GetSettingNotNil()
	}
	if setting.Transparent == configure.TransparentClose {
		return false
	}
	if conf.GetEnvironmentConfig().Lite &&
		(setting.TransparentType == configure.TransparentTproxy ||
			setting.TransparentType == configure.TransparentRedirect) {
		return false
	}
	return true
}

