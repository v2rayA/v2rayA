package service

import (
	"fmt"

	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/ipforward"
	"github.com/v2rayA/v2rayA/kernel/v2ray"
	"github.com/v2rayA/v2rayA/kernel/v2ray/asset"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

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
		return common.Coded("NO_SERVER_SELECTED", fmt.Errorf("no server is selected; select at least one server first"), nil)
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
		return fmt.Errorf("geoip.dat or geosite.dat is missing from %s; put the files there or set --v2ray-assetsdir", asset.GetV2rayLocationAssetOverride())
	}
	if setting.RulePortMode == configure.GfwlistMode || setting.Transparent == configure.TransparentGfwlist {
		if !asset.DoesV2rayAssetExist("LoyalsoldierSite.dat") {
			return asset.GFWListMissingError()
		}
	}
	return nil
}

func checkSupport(toAppend []*configure.Which) (err error) {
	setting := GetSetting()
	if err = checkAssetsExist(setting); err != nil {
		return err
	}
	// Both v2ray-core and xray-core support load balancing now
	// No need to restrict multiple servers for any core type
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
		return fmt.Errorf("no server was given to connect to")
	}
	// Reject a malformed or stale selection before it is stored: AddConnect
	// below persists it, and a stored entry that cannot be located makes
	// every later connect and core start fail with the same error.
	if _, err = which.LocateServerRaw(); err != nil {
		return err
	}
	setting := GetSetting()
	// checkSupport only verifies the geo assets now; the load-balancing
	// restriction it used to report is gone, so any error it returns is fatal.
	if err = checkSupport([]*configure.Which{which}); err != nil {
		return err
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
		// if error occurs, restore the result of connecting. The stored list
		// is restored whether or not the core runs; regenerating the config
		// only makes sense when it does.
		if err != nil {
			if currentConnected != nil && currentConnected.Len() > 0 {
				_ = configure.OverwriteConnects(currentConnected)
			} else {
				_ = configure.ClearConnects(which.Outbound)
			}
			if v2ray.ProcessManager.Running() {
				_ = v2ray.UpdateV2RayConfig()
			}
		}
	}()
	//save the result of connecting to database
	if err = configure.AddConnect(*which); err != nil {
		return
	}
	//update the v2ray config and start/restart v2ray.
	//UpdateV2RayConfig starts the core when it is not running, so a connect
	//request always takes effect instead of silently only saving the selection.
	if err = v2ray.UpdateV2RayConfig(); err != nil {
		return
	}
	return
}

// ReplaceOutboundConnections atomically replaces members of one outbound group.
// It updates v2ray config once after DB changes, and rolls back on failure.
func ReplaceOutboundConnections(outbound string, touches []configure.Which) (err error) {
	log.Trace("ReplaceOutboundConnections: begin")
	defer log.Trace("ReplaceOutboundConnections: done")
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to replace outbound connections: %w", err)
		}
	}()

	if outbound == "" {
		outbound = "proxy"
	}

	// Normalize outbound and deduplicate touches.
	normalized := make([]configure.Which, 0, len(touches))
	seen := make(map[string]struct{})
	for i, wt := range touches {
		if wt.ID <= 0 {
			return fmt.Errorf("invalid touch id at index %d: %d", i, wt.ID)
		}
		switch wt.TYPE {
		case configure.ServerType:
			wt.Sub = 0
		case configure.SubscriptionServerType:
			if wt.Sub < 0 {
				return fmt.Errorf("invalid subscription index at index %d: %d", i, wt.Sub)
			}
		default:
			return fmt.Errorf("invalid touch type at index %d: %q", i, wt.TYPE)
		}
		wt.Outbound = outbound
		key := fmt.Sprintf("%s/%d/%d", wt.TYPE, wt.ID, wt.Sub)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, wt)
	}

	backup := configure.GetConnectedServersByOutbound(outbound)
	restore := func() {
		if backup != nil && backup.Len() > 0 {
			_ = configure.OverwriteConnects(backup)
		} else {
			_ = configure.ClearConnects(outbound)
		}
	}

	if err = configure.ClearConnects(outbound); err != nil {
		return err
	}
	for _, wt := range normalized {
		if err = configure.AddConnect(wt); err != nil {
			restore()
			return err
		}
	}

	if v2ray.ProcessManager.Running() {
		if err = v2ray.UpdateV2RayConfig(); err != nil {
			restore()
			_ = v2ray.UpdateV2RayConfig()
			return err
		}
	}

	return nil
}
