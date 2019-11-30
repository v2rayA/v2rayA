package service

import (
	"V2RayA/model/ipforward"
	"V2RayA/model/v2ray"
	"V2RayA/persistence/configure"
	"errors"
)

func GetSetting() *configure.Setting {
	s := configure.GetSetting()
	if s == nil {
		s = configure.NewSetting()
		_ = configure.SetSetting(s)
	}
	return s
}

func UpdateSetting(setting *configure.Setting) (err error) {
	switch setting.PacMode {
	case configure.GfwlistMode:
		if !v2ray.IsH2yExists() {
			return errors.New("未发现GFWList文件，请更新GFWList后再试")
		}
	case configure.CustomMode:
		if !v2ray.IsCustomExists() {
			return errors.New("未发现custom.dat文件，功能正在开发")
		}
	}
	if setting.Transparent == configure.TransparentGfwlist && !v2ray.IsH2yExists() {
		return errors.New("未发现GFWList文件，请更新GFWList后再试")
	}
	if setting.Transparent != configure.TransparentClose {
		if setting.IpForward != ipforward.IsIpForwardOn() {
			err = ipforward.WriteIpForward(setting.IpForward)
			if err != nil {
				return
			}
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
		err = v2ray.UpdateV2RayConfigAndRestart(&tsr.VmessInfo)
		if err != nil {
			return
		}
	}
	return
}
