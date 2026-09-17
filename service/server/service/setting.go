package service

import (
	"fmt"
	"time"

	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/ipforward"
	"github.com/v2rayA/v2rayA/kernel/v2ray"
	"github.com/v2rayA/v2rayA/kernel/v2ray/asset"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

func GetSetting() *configure.Setting {
	s := configure.GetSettingNotNil()
	if s == nil {
		s = configure.NewSetting()
		_ = configure.SetSetting(s)
	}
	if s.LogLevel == "" {
		s.LogLevel = conf.GetEnvironmentConfig().LogLevel
	}
	return s
}

func UpdateSetting(setting *configure.Setting) (err error) {
	if setting.LogLevel == "" {
		setting.LogLevel = conf.GetEnvironmentConfig().LogLevel
	}
	if (setting.Transparent == configure.TransparentGfwlist || setting.RulePortMode == configure.GfwlistMode) && !asset.DoesV2rayAssetExist("LoyalsoldierSite.dat") {
		return asset.GFWListMissingError()
	}
	if setting.IpForward != ipforward.IsIpForwardOn() {
		e := ipforward.WriteIpForward(setting.IpForward)
		if e != nil {
			log.Warn("UpdateSetting: %v", e)
		}
	}
	previous := configure.GetSettingNotNil()
	err = configure.SetSetting(setting)
	if err != nil {
		return
	}
	log.SetLogLevel(setting.LogLevel)
	// If v2ray is running and has connections, rewrite config and restart connection to make changes to transparent proxy, TCPFastOpen, etc. take effect immediately.
	css := configure.GetConnectedServers()
	if v2ray.ProcessManager.Running() && css.Len() > 0 {
		err = v2ray.UpdateV2RayConfig()
		if err != nil {
			// Put the working setting back and bring the core up with it: a
			// rejected setting used to leave the core stopped, and the next
			// save reported success while nothing was running.
			log.SetLogLevel(previous.LogLevel)
			if e := configure.SetSetting(previous); e != nil {
				log.Warn("UpdateSetting: failed to restore the previous setting: %v", e)
			} else if e := v2ray.UpdateV2RayConfig(); e != nil {
				log.Warn("UpdateSetting: failed to restart the core with the previous setting: %v", e)
			}
			invalidConfigErr := fmt.Errorf("invalid config: the core could not restart with the new settings, the previous ones are back: %w", err)
			return common.Coded("INVALID_CONFIG", invalidConfigErr, map[string]interface{}{"detail": err.Error()})
		}
	}
	if setting.GFWListAutoUpdateMode == configure.AutoUpdateAtIntervals {
		conf.TickerUpdateGFWList.Reset(configure.IntervalHours(setting.GFWListAutoUpdateIntervalHour))
	} else {
		conf.TickerUpdateGFWList.Reset(24 * time.Hour * 365 * 100)
	}
	if setting.SubscriptionAutoUpdateMode == configure.AutoUpdateAtIntervals {
		conf.TickerUpdateSubscription.Reset(configure.IntervalHours(setting.SubscriptionAutoUpdateIntervalHour))
	} else {
		conf.TickerUpdateSubscription.Reset(24 * time.Hour * 365 * 100)
	}
	return
}
