package service

import (
	"V2RayA/global"
	"V2RayA/model/transparentProxy"
	"V2RayA/model/v2ray"
	"V2RayA/persistence/configure"
	"errors"
)

func Disconnect() (err error) {
	err = transparentProxy.StopTransparentProxy(global.Iptables)
	if err != nil {
		return
	}
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
	if which == nil {
		return errors.New("which不能为nil")
	}
	//定位Server
	tsr, err := which.LocateServer()
	if err != nil {
		return
	}
	//根据找到的Server更新V2Ray的配置
	err = v2ray.UpdateV2RayConfigAndRestart(&tsr.VmessInfo)
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
