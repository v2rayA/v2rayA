package v2ray

import (
	"github.com/json-iterator/go"
	"github.com/v2rayA/v2rayA/common/netTools/netstat"
	"github.com/v2rayA/v2rayA/common/ntp"
	"github.com/v2rayA/v2rayA/core/specialMode"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"github.com/v2rayA/v2rayA/core/v2ray/where"
	"github.com/v2rayA/v2rayA/core/vmessInfo"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/global"
	"github.com/v2rayA/v2rayA/plugin"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

func IsV2RayRunning() bool {
	return global.V2RayPID != nil
}

func RestartV2rayService() (err error) {
	if ok, t, err := ntp.IsDatetimeSynced(); err == nil && !ok {
		return newError("Please sync datetime first. Your datetime is ", time.Now().Local().Format(ntp.DisplayFormat), ", and the correct datetime is ", t.Local().Format(ntp.DisplayFormat))
	}
	setting := configure.GetSettingNotNil()
	if (setting.Transparent == configure.TransparentGfwlist || setting.RulePortMode == configure.GfwlistMode) && !asset.IsGFWListExists() {
		return newError("cannot find GFWList files. update GFWList and try again")
	}
	//关闭transparentProxy，防止v2ray在启动DOH时需要解析域名
	_ = killV2ray()
	v2rayBinPath, err := where.GetV2rayBinPath()
	if err != nil {
		return
	}
	dir := path.Dir(v2rayBinPath)
	var params = []string{
		v2rayBinPath,
		"--config=" + asset.GetV2rayConfigPath(),
	}
	if confdir := asset.GetV2rayConfigDirPath(); confdir != "" {
		params = append(params, "--confdir="+confdir)
	}
	log.Println(strings.Join(params, " "))
	assetDir := asset.GetV2rayLocationAsset()
	global.V2RayPID, err = os.StartProcess(v2rayBinPath, params, &os.ProcAttr{
		Dir: dir, //防止找不到v2ctl
		Env: append(os.Environ(),
			"V2RAY_LOCATION_ASSET="+assetDir,
			"XRAY_LOCATION_ASSET="+assetDir,
		),
		Files: []*os.File{
			os.Stderr,
			os.Stdout,
		},
	})
	if err != nil {
		return newError().Base(err)
	}
	defer func() {
		if err != nil && global.V2RayPID != nil {
			_ = killV2ray()
		}
	}()
	//如果inbounds中开放端口，检测端口是否已就绪
	tmplJson := NewTemplate()
	var b []byte
	b, err = asset.GetConfigBytes()
	if err != nil {
		return
	}
	err = jsoniter.Unmarshal(b, &tmplJson)
	if err != nil {
		return
	}
	bPortOpen := false
	var port int
	for _, v := range tmplJson.Inbounds {
		if v.Port != 0 {
			if v.Settings != nil && v.Settings.Network != "" && !strings.Contains(v.Settings.Network, "tcp") {
				continue
			}
			bPortOpen = true
			port = v.Port
			break
		}
	}
	//defer func() {
	//	log.Println(port)
	//	log.Println("\n" + netstat.Print([]string{"tcp", "tcp6"}))
	//	log.Println("\n" + testprint())
	//}()
	startTime := time.Now()
	for {
		if bPortOpen {
			var is bool
			is, err = netstat.IsProcessListenPort(path.Base(v2rayBinPath), port)
			if err != nil {
				return
			}
			if is {
				break
			}
		} else {
			if IsV2RayRunning() {
				time.Sleep(1000 * time.Millisecond) //距离v2ray进程启动到端口绑定可能需要一些时间
				break
			}
		}

		if time.Since(startTime) > 15*time.Second {
			return newError("v2ray-core does not start normally, check the log for more information")
		}
		time.Sleep(1000 * time.Millisecond)
	}
	return
}

/*
传入nil，以当前连接节点更新v2ray配置，视情况重启

传入非nil，以传入VmessInfo更新v2ray配置并重启
*/
func UpdateV2RayConfig(v *vmessInfo.VmessInfo) (err error) {
	CheckAndStopTransparentProxy()
	defer func() {
		if e := CheckAndSetupTransparentProxy(true); e != nil {
			err = newError(e).Base(err)
		}
	}()
	plugin.GlobalPlugins.CloseAll()
	specialMode.StopDNSSupervisor()
	//读配置，转换为v2ray配置并写入
	var (
		tmpl      Template
		sr        *configure.ServerRaw
	)
	if v == nil {
		cs := configure.GetConnectedServer()
		if cs == nil { //没有连接，把v2ray配置更新一下好了
			return pretendToStopV2rayService()
		}
		sr, err = cs.LocateServer()
		if err != nil {
			return
		}
		tmpl, err = NewTemplateFromVmessInfo(sr.VmessInfo)
	} else {
		tmpl, err = NewTemplateFromVmessInfo(*v)
	}
	if err != nil {
		return
	}
	err = WriteV2rayConfig(tmpl.ToConfigBytes())
	if err != nil {
		return
	}

	if v == nil && !IsV2RayRunning() {
		//没有运行就不需要重新启动了
		return
	}
	if v == nil {
		v = &sr.VmessInfo
	}
	if occupied, port, pname := tmpl.CheckInboundPortsOccupied(); occupied {
		return newError("Port ", port, " is occupied by ", pname)
	}
	err = RestartV2rayService()
	if err != nil {
		return
	}

	// try launching plugin client
	var plu plugin.Plugin
	plu, err = plugin.NewPlugin(global.GetEnvironmentConfig().PluginListenPort, *v)
	if err != nil {
		return
	}
	if plu != nil {
		plugin.GlobalPlugins.Append(plu)
	}

	specialMode.CheckAndSetupDNSSupervisor()
	return
}

/*清空inbounds规则来假停v2ray*/
func pretendToStopV2rayService() (err error) {
	tmplJson := NewTemplate()
	b, err := asset.GetConfigBytes()
	if err != nil {
		return
	}

	err = jsoniter.Unmarshal(b, &tmplJson)
	if err != nil {
		log.Println(string(b), err)
		return
	}
	tmplJson.Inbounds = make([]Inbound, 0)
	b, _ = jsoniter.Marshal(tmplJson)
	err = WriteV2rayConfig(b)
	if err != nil {
		return
	}
	if IsV2RayRunning() {
		if occupied, port, pname := tmplJson.CheckInboundPortsOccupied(); occupied {
			return newError("Port ", port, " is occupied by ", pname)
		}
		err = RestartV2rayService()
	}
	return
}

func killV2ray() (err error) {
	if global.V2RayPID != nil {
		err = global.V2RayPID.Kill()
		if err != nil {
			return
		}
		_, err = global.V2RayPID.Wait()
		if err != nil {
			return
		}
		global.V2RayPID = nil
	}
	return
}

func StopV2rayService() (err error) {
	defer CheckAndStopTransparentProxy()
	defer func() {
		if IsV2RayRunning() {
			msg := "failed to stop v2ray"
			if err != nil && len(strings.TrimSpace(err.Error())) > 0 {
				msg += ": " + err.Error()
			}
			err = newError(msg)
		}
	}()
	return killV2ray()
}
