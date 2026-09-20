package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/ipforward"
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
	previousIpForward := ipforward.IsIpForwardOn()
	ipForwardChanged := false
	if setting.IpForward != previousIpForward {
		if e := ipforward.WriteIpForward(setting.IpForward); e != nil {
			log.Warn("UpdateSetting: %v", e)
		} else {
			ipForwardChanged = true
		}
	}
	var previous *configure.Setting
	sideEffectsRestored := true
	err = ApplyCoreConfig(func() func() error {
		previous = configure.GetSettingNotNil()
		return func() error {
			log.SetLogLevel(previous.LogLevel)
			if ipForwardChanged {
				if e := ipforward.WriteIpForward(previousIpForward); e != nil {
					sideEffectsRestored = false
					log.Warn("UpdateSetting: failed to restore ip forwarding: %v", e)
				}
			}
			return configure.SetSetting(previous)
		}
	}, func() error {
		if e := configure.SetSetting(setting); e != nil {
			return e
		}
		log.SetLogLevel(setting.LogLevel)
		return nil
	})
	if err != nil {
		var failure *ApplyCoreConfigError
		if !errors.As(err, &failure) {
			return err
		}
		if failure.RestoreStoreErr != nil {
			log.Warn("UpdateSetting: failed to restore the previous setting: %v", failure.RestoreStoreErr)
		}
		if failure.RestoreUpdateErr != nil {
			log.Warn("UpdateSetting: failed to restart the core with the previous setting: %v", failure.RestoreUpdateErr)
		}
		restored := failure.Restored() && sideEffectsRestored
		message := "invalid config: the core could not restart with the new settings, the previous ones are back: %w"
		if !restored {
			message = "invalid config: the core could not restart with the new settings and the previous ones could not be restored either, the core is stopped: %w"
		}
		invalidConfigErr := fmt.Errorf(message, failure.UpdateErr)
		return common.Coded("INVALID_CONFIG", invalidConfigErr, map[string]interface{}{"detail": failure.UpdateErr.Error(), "restored": restored})
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
