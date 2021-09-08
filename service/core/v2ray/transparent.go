package v2ray

import (
	"fmt"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/core/iptables"
	"github.com/v2rayA/v2rayA/core/specialMode"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"strings"
	"time"
)

func DeleteTransparentProxyRules() {
	removeResolvHijacker()
	iptables.CloseWatcher()
	iptables.Tproxy.GetCleanCommands().Clean()
	iptables.Redirect.GetCleanCommands().Clean()
	iptables.DropSpoofing.GetCleanCommands().Clean()
	time.Sleep(100 * time.Millisecond)
}

func WriteTransparentProxyRules(preprocess func(c *iptables.SetupCommands)) (err error) {
	defer func() {
		if err != nil {
			log.Warn("WriteTransparentProxyRules: %v", err)
			DeleteTransparentProxyRules()
		}
	}()
	if specialMode.ShouldUseSupervisor() {
		if err = iptables.DropSpoofing.GetSetupCommands().Setup(preprocess); err != nil {
			log.Warn("DropSpoofing can't be enable: %v", err)
			return err
		}
	}
	setting := configure.GetSettingNotNil()
	if setting.TransparentType == configure.TransparentTproxy {
		if err = iptables.Tproxy.GetSetupCommands().Setup(preprocess); err != nil {
			if strings.Contains(err.Error(), "TPROXY") && strings.Contains(err.Error(), "No chain") {
				err = fmt.Errorf("you does not compile xt_TPROXY in kernel")
			}
			return fmt.Errorf("not support \"tproxy\" mode of transparent proxy: %w", err)
		}
		iptables.SetWatcher(&iptables.Tproxy)
	} else if setting.TransparentType == configure.TransparentRedirect {
		if err = iptables.Redirect.GetSetupCommands().Setup(preprocess); err != nil {
			return fmt.Errorf("not support \"redirect\" mode of transparent proxy: %w", err)
		}
		iptables.SetWatcher(&iptables.Redirect)
	}
	if specialMode.ShouldLocalDnsListen() {
		if couldListenLocalhost, e := specialMode.CouldLocalDnsListen(); couldListenLocalhost {
			if e != nil {
				log.Warn("only listen at 127.2.0.17: %v", e)
			}
			resetResolvHijacker()
		} else if specialMode.ShouldUseFakeDns() {
			return fmt.Errorf("fakedns cannot be enabled: %w", e)
		} else {
			log.Warn("WriteTransparentProxyRules: %v", e)
		}
	}
	return nil
}

func CheckAndSetupTransparentProxy(checkRunning bool, setting *configure.Setting) (err error) {
	if conf.GetEnvironmentConfig().Lite {
		return nil
	}
	if setting != nil {
		setting.FillEmpty()
	} else {
		setting = configure.GetSettingNotNil()
	}
	if (!checkRunning || ProcessManager.Running()) && setting.Transparent != configure.TransparentClose {
		DeleteTransparentProxyRules()
		err = WriteTransparentProxyRules(func(c *iptables.SetupCommands) {

		})
	}
	return
}

func CheckAndStopTransparentProxy() {
	if conf.GetEnvironmentConfig().Lite {
		return
	}
	DeleteTransparentProxyRules()
}
