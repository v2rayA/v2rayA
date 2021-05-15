package service

import (
	"github.com/v2rayA/v2rayA/core/ipforward"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/core/v2ray/asset/gfwlist"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/plugin"
	"log"
	"net"
	"os"
)

func Disconnect() (err error) {
	plugin.GlobalPlugins.CloseAll()
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
	if setting.PacMode == configure.GfwlistMode || setting.Transparent == configure.TransparentGfwlist {
		if !gfwlist.LoyalsoldierSiteDatExists() {
			return newError("GFWList file not exists. Try updating GFWList please")
		}
	}
	return nil
}

const resolvConf = "/etc/resolv.conf"

func writeResolvConf() {
	os.WriteFile(resolvConf, []byte("nameserver 223.5.5.5"), 0644)
}

func checkResolvConf() {
	if _, err := os.Stat(resolvConf); os.IsNotExist(err) {
		writeResolvConf()
	} else {
		errCnt := 0
		maxTry := 2
		for {
			addrs, err := net.LookupHost("apple.com")
			if len(addrs) == 0 || err != nil {
				errCnt++
				if errCnt <= maxTry {
					continue
				}
			}
			break
		}
		if errCnt >= maxTry {
			log.Println("[warning] There may be no network or dns manager conflicting with v2rayA. If problems occur, paste your file /etc/resolv.conf for help.")
			writeResolvConf()
		}
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
		err = ipforward.WriteIpForward(setting.IntranetSharing)
		if err != nil {
			return
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
