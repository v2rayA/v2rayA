package service

import (
	"fmt"
	"github.com/v2rayA/v2rayA/core/ipforward"
	"github.com/v2rayA/v2rayA/core/specialMode"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/plugin"
	"log"
)

func StopV2ray() (err error) {
	plugin.GlobalPlugins.CloseAll()
	specialMode.StopDNSSupervisor()
	err = v2ray.StopV2rayService()
	if err != nil {
		return
	}
	return
}
func StartV2ray() (err error) {
	if css := configure.GetConnectedServers(); len(css) == 0 {
		return fmt.Errorf("failed: no server is connected. connect a server instead")
	}
	return v2ray.UpdateV2RayConfig()
}

func Disconnect(outbound string) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to disconnect: %w", err)
		}
	}()
	lastConnected := configure.GetConnectedServer(outbound)
	err = configure.ClearConnected(outbound)
	if err != nil {
		return
	}
	//update the v2ray config and restart v2ray
	if v2ray.IsV2RayRunning() || len(configure.GetOutbounds()) <= 1 {
		defer func() {
			if err != nil && lastConnected != nil && v2ray.IsV2RayRunning() {
				_ = configure.SetConnect(lastConnected)
				_ = v2ray.UpdateV2RayConfig()
			}
		}()
		if err = v2ray.UpdateV2RayConfig(); err != nil {
			return
		}
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

func Connect(which *configure.Which) (err error) {
	log.Println("Connect: begin")
	defer log.Println("Connect: done")
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to connect: %w", err)
		}
	}()
	setting := GetSetting()
	if err = checkAssetsExist(setting); err != nil {
		return
	}
	if which == nil {
		return newError("which can not be nil")
	}
	//configure the ip forward
	if setting.IntranetSharing != ipforward.IsIpForwardOn() {
		e := ipforward.WriteIpForward(setting.IntranetSharing)
		if e != nil {
			log.Println("[warning]", e)
		}
	}
	//locate server
	currentConnected := configure.GetConnectedServer(which.Outbound)
	defer func() {
		// if error occurs, restore the result of connecting
		if err != nil && currentConnected != nil && v2ray.IsV2RayRunning() {
			_ = configure.SetConnect(currentConnected)
			_ = v2ray.UpdateV2RayConfig()
		}
	}()
	//save the result of connecting to database
	err = configure.SetConnect(which)
	//update the v2ray config and start/restart v2ray
	if v2ray.IsV2RayRunning() || len(configure.GetOutbounds()) <= 1 {
		if err = v2ray.UpdateV2RayConfig(); err != nil {
			return
		}
	}
	return
}
