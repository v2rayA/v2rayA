package service

import (
	"V2RayA/core/v2ray"
	"V2RayA/global"
	"V2RayA/persistence/configure"
	"errors"
	"log"
)

func Disconnect() (err error) {
	global.SSRs.ClearAll()
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

func Connect(which *configure.Which) (err error) {
	log.Println("Connect: begin")
	defer log.Println("Connect: done")
	if which == nil {
		return errors.New("which can not be nil")
	}
	//定位Server
	tsr, err := which.LocateServer()
	if err != nil {
		log.Println(err)
		return
	}
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
