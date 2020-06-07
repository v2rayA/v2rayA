package service

import (
	"github.com/mzz2017/v2rayA/core/v2ray"
	"github.com/mzz2017/v2rayA/core/v2ray/asset/gfwlist"
	"github.com/mzz2017/v2rayA/db/configure"
	"github.com/mzz2017/v2rayA/global"
	"log"
)

func Disconnect() (err error) {
	global.Plugins.CloseAll()
	err = v2ray.StopV2rayService()
	if err != nil {
		return
	}
	err = v2ray.DisableV2rayService()
	if err != nil {
		return
	}
	err = configure.ClearConnected()
	if err != nil {
		return
	}
	return
}

func checkAssetsExist(setting *configure.Setting) error {
	//FIXME: non-fully check
	if setting.PacMode == configure.GfwlistMode || setting.Transparent == configure.TransparentGfwlist {
		if !gfwlist.LoyalsoldierSiteDatExists() {
			return newError("GFWList file not exists. Try updating GFWList please")
		}
	}
	return nil
}

func Connect(which *configure.Which) (err error) {
	log.Println("Connect: begin")
	defer log.Println("Connect: done")
	setting := GetSetting()
	if err = checkAssetsExist(setting); err != nil {
		return
	}
	if which == nil {
		return newError("which can not be nil")
	}
	//定位Server
	tsr, err := which.LocateServer()
	if err != nil {
		log.Println(err)
		return
	}
	cs := configure.GetConnectedServer()
	defer func() {
		if err != nil && cs != nil && v2ray.IsV2RayRunning() {
			_ = configure.SetConnect(cs)
		}
	}()
	//unset connectedServer to avoid refresh in advance
	_ = configure.ClearConnected()
	//根据找到的Server更新V2Ray的配置
	err = v2ray.UpdateV2RayConfig(&tsr.VmessInfo)
	if err != nil {
		return
	}
	//保存节点连接成功的结果
	err = configure.SetConnect(which)
	if err != nil {
		return
	}
	return v2ray.EnableV2rayService()
}
