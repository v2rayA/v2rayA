package service

import (
	"fmt"
	"time"

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
			return fmt.Errorf("invalid config: the core could not restart with the new settings: %w", err)
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
