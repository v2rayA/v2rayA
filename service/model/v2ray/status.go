package v2ray

import (
	"V2RayA/global"
	"V2RayA/model/dnsPoison/entity"
	"V2RayA/model/shadowsocksr"
	"V2RayA/model/v2ray/asset"
	"V2RayA/model/vmessInfo"
	"V2RayA/persistence/configure"
	"V2RayA/tools/ports"
	"bytes"
	"errors"
	"github.com/json-iterator/go"
	"log"
	"net"
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
		out, err := exec.Command("sh", "-c", "service v2ray status|head -n 5|grep running").CombinedOutput()
		if err != nil || strings.Contains(string(out), "not running") {
			return false
		}
	case global.SystemctlMode:
		out, err := exec.Command("sh", "-c", "systemctl status v2ray|head -n 5|grep running").Output()
		return err == nil && len(out) > 0
	}
	return true
}
func RestartV2rayService() (err error) {
	setting := configure.GetSettingNotNil()
	if (setting.Transparent == configure.TransparentGfwlist || setting.PacMode == configure.GfwlistMode) && !asset.IsGFWListExists() {
		return errors.New("cannot find GFWList files. update GFWList and try again")
	}
	//关闭transparentProxy，防止v2ray在启动DOH时需要解析域名
	var out []byte
	switch global.ServiceControlMode {
	case global.ServiceMode:
		out, err = exec.Command("sh", "-c", "service v2ray restart").CombinedOutput()
		if err != nil {
			err = errors.New(err.Error() + string(out))
		}
	case global.SystemctlMode:
		out, err = exec.Command("sh", "-c", "systemctl restart v2ray").Output()
		if err != nil {
			err = errors.New(err.Error() + string(out))
		}
	case global.UniversalMode:
		_ = killV2ray()
		v2wd, _ := asset.GetV2rayWorkingDir()
		v2ctlDir, _ := asset.GetV2ctlDir()
		global.V2RayPID, err = os.StartProcess(v2wd+"/v2ray", []string{"--config=" + asset.GetConfigPath()}, &os.ProcAttr{
			Dir: v2ctlDir, //防止找不到v2ctl
			Env: os.Environ(),
			Files: []*os.File{
				os.Stdin,
				os.Stdout,
				os.Stderr,
			},
		})
		if err != nil {
			err = errors.New(err.Error() + string(out))
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
	var sPortOpen string
	for _, v := range tmplJson.Inbounds {
		if v.Port != 0 {
			bPortOpen = true
			sPortOpen = strconv.Itoa(v.Port)
			break
		}
	}
	startTime := time.Now()
	for {
		if bPortOpen {
			if p, which := ports.IsPortOccupied(sPortOpen, "tcp", true); p && strings.Contains(which, "v2ray") {
				break
			}
		} else {
			if IsV2RayRunning() {
				time.Sleep(1000 * time.Millisecond) //距离v2ray进程启动到端口绑定可能需要一些时间
				break
			}
		}
		if time.Since(startTime) > 15*time.Second {
			return errors.New("v2ray-core does not start normally, there may be a problem with the configuration file or the required port is occupied")
		}
		time.Sleep(500 * time.Millisecond)
	}
	return
}

/*
传入nil，以当前连接节点更新v2ray配置，视情况重启

传入非nil，以传入VmessInfo更新v2ray配置并重启
*/
func UpdateV2RayConfig(v *vmessInfo.VmessInfo) (err error) {
	CheckAndStopTransparentProxy()
	defer CheckAndSetupTransparentProxy(true)
	//iptables.SpoofingFilter.GetCleanCommands().Clean()
	//defer iptables.SpoofingFilter.GetSetupCommands().Setup(nil)
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
			log.Println(err)
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

	global.SSRs.ClearAll()
	entity.StopDNSPoison()

	if v == nil && !IsV2RayRunning() {
		//没有运行就不需要重新启动了
		return
	}
	err = RestartV2rayService()
	if err != nil {
		return
	}
	if len(tmpl.Outbounds) > 0 && tmpl.Outbounds[0].Protocol == "socks" {
		//说明是ss或ssr，启动ssr server
		// 尝试将address解析成ip
		if v == nil {
			v = &sr.VmessInfo
		}
		if net.ParseIP(v.Add) == nil {
			addrs, e := net.LookupHost(v.Add)
			if e == nil && len(addrs) > 0 {
				v.Add = addrs[0]
			}
		}
		ss := new(shadowsocksr.SSR)
		err = ss.Serve(global.GetEnvironmentConfig().SSRListenPort, v.Net, v.ID, v.Add, v.Port, v.TLS, v.Path, v.Type, v.Host)
		if err != nil {
			return
		}
		global.SSRs.Append(*ss)
	}
	if configure.GetSettingNotNil().Transparent != configure.TransparentClose && !global.SupportTproxy {
		//redirect+poison增强方案
		entity.SetupDnsPoisonWithExtraInfo(extraInfo)
	}
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
			msg := "fail in stopping v2ray"
			if err != nil && len(strings.TrimSpace(err.Error())) > 0 {
				msg += ": " + err.Error()
			}
			err = errors.New(msg)
		}
	}()
	var out []byte
	switch global.ServiceControlMode {
	case global.UniversalMode:
		return killV2ray()
	case global.ServiceMode:
		out, err = exec.Command("sh", "-c", "service v2ray stop").CombinedOutput()
		if err != nil {
			err = errors.New(err.Error() + string(out))
		}
	case global.SystemctlMode:
		out, err = exec.Command("sh", "-c", "systemctl stop v2ray").CombinedOutput()
		if err != nil {
			err = errors.New(err.Error() + string(out))
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
