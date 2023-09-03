package service

import (
	"errors"
	"fmt"

	"github.com/v2rayA/v2rayA/core/ipforward"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"github.com/v2rayA/v2rayA/core/v2ray/service"
	"github.com/v2rayA/v2rayA/core/v2ray/where"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

var V2OnlyFeatureError = fmt.Errorf("v2fly/v2ray-core only feature")

func StopV2ray() (err error) {
	v2ray.ProcessManager.Stop(true)
	return nil
}
func StartV2ray() (err error) {
	if err = checkSupport(nil); err != nil {
		return err
	}
	//configure the ip forward
	setting := GetSetting()
	if setting.IpForward != ipforward.IsIpForwardOn() {
		e := ipforward.WriteIpForward(setting.IpForward)
		if e != nil {
			log.Warn("Connect: %v", e)
		}
	}
	if css := configure.GetConnectedServers(); css.Len() == 0 {
		return fmt.Errorf("failed: no server is selected. please select at least one server")
	}
	return v2ray.UpdateV2RayConfig()
}

func Disconnect(which configure.Which, clearOutbound bool) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to disconnect: %w", err)
		}
	}()
	lastConnected := configure.GetConnectedServersByOutbound(which.Outbound)
	if clearOutbound {
		err = configure.ClearConnects(which.Outbound)
	} else {
		err = configure.RemoveConnect(which)
	}
	if err != nil {
		return
	}
	//update the v2ray config and restart v2ray
	if v2ray.ProcessManager.Running() {
		defer func() {
			if err != nil && lastConnected != nil && v2ray.ProcessManager.Running() {
				_ = configure.OverwriteConnects(lastConnected)
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
	if !asset.DoesV2rayAssetExist("geoip.dat") || !asset.DoesV2rayAssetExist("geosite.dat") {
		return fmt.Errorf("geoip.dat or geosite.dat file does not exists")
	}
	if setting.RulePortMode == configure.GfwlistMode || setting.Transparent == configure.TransparentGfwlist {
		if !asset.DoesV2rayAssetExist("LoyalsoldierSite.dat") {
			return fmt.Errorf("GFWList file does not exists. Try updating GFWList please")
		}
	}
	return nil
}

func checkSupport(toAppend []*configure.Which) (err error) {
	setting := GetSetting()
	if err = checkAssetsExist(setting); err != nil {
		return err
	}
	variant, err := service.CheckV5()
	if err != nil {
		return err
	}
	if variant != where.V2ray {
		outbound2cnt := map[string]int{}
		for _, wt := range append(toAppend, configure.GetConnectedServers().Get()...) {
			outbound2cnt[wt.Outbound]++
			if outbound2cnt[wt.Outbound] > 1 {
				return fmt.Errorf("cannot select multiple servers: %w", V2OnlyFeatureError)
			}
		}
	}
	return nil
}

func Connect(which *configure.Which) (err error) {
	log.Trace("Connect: begin")
	defer log.Trace("Connect: done")
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to connect: %w", err)
		}
	}()
	if which == nil {
		return fmt.Errorf("which can not be nil")
	}
	setting := GetSetting()
	if err = checkSupport([]*configure.Which{which}); err != nil {
		if !errors.Is(err, V2OnlyFeatureError) {
			return err
		}
		if err = configure.ClearConnects(which.Outbound); err != nil {
			return err
		}
	}
	//configure the ip forward
	if setting.IpForward != ipforward.IsIpForwardOn() {
		e := ipforward.WriteIpForward(setting.IpForward)
		if e != nil {
			log.Warn("Connect: %v", e)
		}
	}
	//locate server
	currentConnected := configure.GetConnectedServersByOutbound(which.Outbound)
	defer func() {
		// if error occurs, restore the result of connecting
		if err != nil && currentConnected != nil && v2ray.ProcessManager.Running() {
			_ = configure.OverwriteConnects(currentConnected)
			_ = v2ray.UpdateV2RayConfig()
		}
	}()
	//save the result of connecting to database
	if err = configure.AddConnect(*which); err != nil {
		return
	}
	//update the v2ray config and start/restart v2ray
	if v2ray.ProcessManager.Running() {
		if err = v2ray.UpdateV2RayConfig(); err != nil {
			return
		}
	}
	return
}
