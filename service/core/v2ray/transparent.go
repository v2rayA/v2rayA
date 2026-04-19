package v2ray

import (
	"fmt"
	"strings"
	"time"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/core/iptables"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

func deleteTransparentProxyRules() {
	stopTinyTun()
	iptables.CloseWatcher()
	if !conf.GetEnvironmentConfig().Lite {
		removeResolvHijacker()
		iptables.Tproxy.GetCleanCommands().Run(false)
		iptables.Redirect.GetCleanCommands().Run(false)
		iptables.DropSpoofing.GetCleanCommands().Run(false)
	}
	iptables.SystemProxy.GetCleanCommands().Run(false)
	time.Sleep(30 * time.Millisecond)
}

func writeTransparentProxyRules(tmpl *Template) (err error) {
	defer func() {
		if err != nil {
			log.Warn("writeTransparentProxyRules: %v", err)
			deleteTransparentProxyRules()
		}
	}()
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

