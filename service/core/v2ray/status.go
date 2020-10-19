package v2ray

import (
	"bytes"
	"fmt"
	netstat2 "github.com/cakturk/go-netstat/netstat"
	"github.com/json-iterator/go"
	"github.com/v2rayA/v2rayA/common/netTools/netstat"
	"github.com/v2rayA/v2rayA/common/ntp"
	"github.com/v2rayA/v2rayA/core/dnsPoison/entity"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"github.com/v2rayA/v2rayA/core/v2ray/where"
	"github.com/v2rayA/v2rayA/core/vmessInfo"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/global"
	"github.com/v2rayA/v2rayA/plugin"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func IsV2RayProcessExists() bool {
	out, err := exec.Command("sh", "-c", "ps -e -o args").CombinedOutput()
	if err != nil || (strings.Contains(string(out), "invalid option") && strings.Contains(strings.ToLower(string(out)), "busybox")) {
		out, err = exec.Command("sh", "-c", "ps|awk '{print $4,$5}'").CombinedOutput()
	}
	if err == nil {
		//模拟grep -E
		regex := regexp.MustCompile(`\bv2ray\b`)
		lines := bytes.Split(out, []byte{'\n'})
		for _, line := range lines {
			if regex.Match(line) && !bytes.Contains(line, []byte("-version")) {
				return true
			}
		}
	}
	return false
}

func IsV2RayRunning() bool {
	switch global.ServiceControlMode {
	case global.UniversalMode:
		return IsV2RayProcessExists()
	case global.ServiceMode:
		out, err := exec.Command("sh", "-c", "service v2ray status|grep running").CombinedOutput()
		if err != nil || strings.Contains(string(out), "not running") {
			return false
		}
	case global.SystemctlMode:
		out, err := exec.Command("sh", "-c", "systemctl status v2ray|grep Active|grep running").Output()
		return err == nil && len(out) > 0
	}
	return true
}

func testprint() string {
	var buffer strings.Builder
	e, _ := netstat2.TCPSocks(func(entry *netstat2.SockTabEntry) bool {
		return true
	})
	e2, _ := netstat2.TCP6Socks(func(entry *netstat2.SockTabEntry) bool {
		return true
	})
	buffer.WriteString(fmt.Sprintf("%-6v%-25v%-25v%-15v%-6v%-9v%v\n", "Proto", "Local Address", "Foreign Address", "State", "User", "Inode", "PID/Program name"))
	for _, v := range e {
		var pstr string
		if v.Process != nil {
			pstr = v.Process.String()
		}
		buffer.WriteString(fmt.Sprintf(
			"%-6v%-25v%-25v%-15v%-6v%-9v%v\n",
			"tcp",
			v.LocalAddr.IP.String()+"/"+strconv.Itoa(int(v.LocalAddr.Port)),
			v.RemoteAddr.IP.String()+"/"+strconv.Itoa(int(v.RemoteAddr.Port)),
			v.State.String(),
			v.UID,
			"",
			pstr,
		))
	}
	for _, v := range e2 {
		var pstr string
		if v.Process != nil {
			pstr = v.Process.String()
		}
		buffer.WriteString(fmt.Sprintf(
			"%-6v%-25v%-25v%-15v%-6v%-9v%v\n",
			"tcp6",
			v.LocalAddr.IP.String()+"/"+strconv.Itoa(int(v.LocalAddr.Port)),
			v.RemoteAddr.IP.String()+"/"+strconv.Itoa(int(v.RemoteAddr.Port)),
			v.State.String(),
			v.UID,
			"",
			pstr,
		))
	}
	return buffer.String()
}

func RestartV2rayService() (err error) {
	if ok, t, err := ntp.IsDatetimeSynced(); err == nil && !ok {
		return newError("Please sync datetime first. Your datetime is ", time.Now().Local().Format(ntp.DisplayFormat), ", and the correct datetime is ", t.Local().Format(ntp.DisplayFormat))
	}
	setting := configure.GetSettingNotNil()
	if (setting.Transparent == configure.TransparentGfwlist || setting.PacMode == configure.GfwlistMode) && !asset.IsGFWListExists() {
		return newError("cannot find GFWList files. update GFWList and try again")
	}
	//关闭transparentProxy，防止v2ray在启动DOH时需要解析域名
	var out []byte
	switch global.ServiceControlMode {
	case global.ServiceMode:
		out, err = exec.Command("sh", "-c", "service v2ray restart").CombinedOutput()
		if err != nil {
			err = newError(string(out)).Base(err)
		}
	case global.SystemctlMode:
		out, err = exec.Command("sh", "-c", "systemctl restart v2ray").Output()
		if err != nil {
			err = newError(string(out)).Base(err)
		}
	case global.UniversalMode:
		_ = killV2ray()
		v2wd, _ := where.GetV2rayWorkingDir()
		v2ctlDir, _ := asset.GetV2ctlDir()
		log.Println(v2wd+"/v2ray", "--config="+asset.GetConfigPath(), v2ctlDir)
		global.V2RayPID, err = os.StartProcess(v2wd+"/v2ray", []string{v2wd + "/v2ray", "--config=" + asset.GetConfigPath()}, &os.ProcAttr{
			Dir: v2ctlDir, //防止找不到v2ctl
			Env: os.Environ(),
			Files: []*os.File{
				os.Stdin,
				os.Stdout,
				os.Stderr,
			},
		})
		if err != nil {
			err = newError(string(out)).Base(err)
		}
	}
	if err != nil {
		return
	}
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
			is, err = netstat.IsProcessListenPort("v2ray", port)
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
			return newError("v2ray-core does not start normally, there may be a problem of the configuration file or the required port is occupied")
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
	//iptables.SpoofingFilter.GetCleanCommands().Clean()
	//defer iptables.SpoofingFilter.GetSetupCommands().Setup(nil)
	plugin.GlobalPlugins.CloseAll()
	entity.StopDNSPoison()
	//读配置，转换为v2ray配置并写入
	var (
		tmpl      Template
		sr        *configure.ServerRaw
		extraInfo *entity.ExtraInfo
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
		tmpl, extraInfo, err = NewTemplateFromVmessInfo(sr.VmessInfo)
	} else {
		tmpl, extraInfo, err = NewTemplateFromVmessInfo(*v)
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
	if v.Protocol != "" && v.Protocol != "vmess" && v.Protocol != "vless" {
		// 说明是plugin，启动plugin client
		var plu plugin.Plugin
		plu, err = plugin.NewPlugin(global.GetEnvironmentConfig().PluginListenPort, *v)
		if err != nil {
			return
		}
		plugin.GlobalPlugins.Append(plu)
	}

	entity.CheckAndSetupDnsPoisonWithExtraInfo(extraInfo)
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
	var out []byte
	switch global.ServiceControlMode {
	case global.UniversalMode:
		return killV2ray()
	case global.ServiceMode:
		out, err = exec.Command("sh", "-c", "service v2ray stop").CombinedOutput()
		if err != nil {
			err = newError(string(out)).Base(err)
		}
	case global.SystemctlMode:
		out, err = exec.Command("sh", "-c", "systemctl stop v2ray").CombinedOutput()
		if err != nil {
			err = newError(string(out)).Base(err)
		}
	}
	return
}

func StopAndDisableV2rayService() (err error) {
	err = StopV2rayService()
	if err != nil {
		return
	}
	return DisableV2rayService()
}
