package service

import (
	"github.com/v2rayA/v2rayA/core/ipforward"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/global"
	"log"
	"time"
)

func GetSetting() *configure.Setting {
	s := configure.GetSettingNotNil()
	if s == nil {
		s = configure.NewSetting()
		_ = configure.SetSetting(s)
	}
	return s
}

func UpdateSetting(setting *configure.Setting) (err error) {
	if (setting.Transparent == configure.TransparentGfwlist || setting.RulePortMode == configure.GfwlistMode) && !asset.IsGFWListExists() {
		return newError("cannot find GFWList files. update GFWList and try again")
	}
	if setting.IntranetSharing != ipforward.IsIpForwardOn() {
		e := ipforward.WriteIpForward(setting.IntranetSharing)
		if e != nil {
			log.Println("[warning]", e)
		}
	}
	err = configure.SetSetting(setting)
	if err != nil {
		return
	}
	//如果v2ray正在运行且有连接，则重写配置并重启连接，使得对PAC模式、TCPFastOpen等配置的修改立即生效
	cs := configure.GetConnectedServer()
	if cs != nil && v2ray.IsV2RayRunning() {
		tsr, _ := cs.LocateServer()
		err = v2ray.UpdateV2RayConfig(&tsr.VmessInfo)
		if err != nil {
			return
		}
	}
	if setting.GFWListAutoUpdateMode == configure.AutoUpdateAtIntervals {
		global.TickerUpdateGFWList.Reset(time.Duration(setting.GFWListAutoUpdateIntervalHour) * time.Hour)
	} else {
		global.TickerUpdateGFWList.Reset(24 * time.Hour * 365 * 100)
	}
	if setting.SubscriptionAutoUpdateMode == configure.AutoUpdateAtIntervals {
		global.TickerUpdateSubscription.Reset(time.Duration(setting.SubscriptionAutoUpdateIntervalHour) * time.Hour)
	} else {
		global.TickerUpdateSubscription.Reset(24 * time.Hour * 365 * 100)
	}
	return
}
