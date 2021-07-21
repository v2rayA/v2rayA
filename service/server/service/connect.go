package service

import (
	"github.com/v2rayA/v2rayA/core/ipforward"
	"github.com/v2rayA/v2rayA/core/specialMode"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/plugin"
	"log"
	"os"
)

func Disconnect() (err error) {
	plugin.GlobalPlugins.CloseAll()
	specialMode.StopDNSSupervisor()
	err = v2ray.StopV2rayService()
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
	if setting.RulePortMode == configure.GfwlistMode || setting.Transparent == configure.TransparentGfwlist {
		if !asset.LoyalsoldierSiteDatExists() {
			return newError("GFWList file not exists. Try updating GFWList please")
		}
	}
	return nil
}

const resolvConf = "/etc/resolv.conf"

func writeResolvConf() {
	os.WriteFile(resolvConf, []byte("nameserver 223.6.6.6"), 0644)
}

func checkResolvConf() {
	if _, err := os.Stat(resolvConf); os.IsNotExist(err) {
		writeResolvConf()
	}
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
	checkResolvConf()
	//配置ip转发
	if setting.IntranetSharing != ipforward.IsIpForwardOn() {
		e := ipforward.WriteIpForward(setting.IntranetSharing)
		if e != nil {
			log.Println("[warning]", e)
		}
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
	//v2ray.EnableV2rayService()
	return
}
