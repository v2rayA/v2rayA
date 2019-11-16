package service

import (
	"V2RayA/model/v2ray"
	"V2RayA/persistence/configure"
	"errors"
)

func GetSetting() *configure.Setting {
	s := configure.GetSetting()
	if s == nil {
		s = configure.NewSetting()
	}
	return s
}

func UpdateSetting(setting *configure.Setting) (err error) {
	//TODO: 检查参数合法性
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
	err = configure.SetSetting(setting)
	if err != nil {
		return
	}
	//如果当前有连接，则重写配置并重启连接，使得对PAC的修改立即生效
	if cs := configure.GetConnectedServer(); cs != nil {
		tsr, _ := cs.LocateServer()
		err = v2ray.UpdateV2RayConfigAndRestart(&tsr.VmessInfo)
		if err != nil {
			return
		}

	}
	return
}
