package v2ray

import (
	"V2RayA/global"
	"V2RayA/model/iptables"
	"V2RayA/model/vmessInfo"
	"V2RayA/persistence/configure"
	"V2RayA/tools"
	"errors"
	"fmt"
	"github.com/json-iterator/go"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"
)

func IsV2RayProcessExists() bool {
	out, err := exec.Command("sh", "-c", "ps -ef|awk '{print $9,$8}'|grep v2ray$").CombinedOutput()
	if err != nil || (strings.Contains(string(out), "invalid option") && strings.Contains(strings.ToLower(string(out)), "busybox")) {
		out, err = exec.Command("sh", "-c", "ps|awk '{print $4,$5}'|grep v2ray$").CombinedOutput()
	}
	return err == nil && len(strings.TrimSpace(string(out))) > 0 && !strings.Contains(string(out), "-version")
}

func IsV2RayRunning() bool {
	switch global.ServiceControlMode {
	case global.DockerMode:
		b, err := ioutil.ReadFile(GetConfigPath())
		if err != nil {
			return false
		}
		return len(gjson.GetBytes(b, "inbounds").Array()) > 0
	case global.CommonMode:
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
	var out []byte
	switch global.ServiceControlMode {
	case global.DockerMode:
		//看inbounds是不是空的，是的话就补上
		tmplJson := NewTemplate()
		var b []byte
		b, err = ioutil.ReadFile(GetConfigPath())
		if err != nil {
			return
		}
		err = jsoniter.Unmarshal(b, &tmplJson)
		if err != nil {
			return
		}
		if len(tmplJson.Inbounds) <= 0 {
			// 读入模板json
			rawJson := NewTemplate()
			raw := []byte(TemplateJson)
			err = jsoniter.Unmarshal(raw, &rawJson)
			if err != nil {
				return errors.New("读入模板json出错，请检查templateJson变量是否是正确的json格式")
			}
			tmplJson.Inbounds = rawJson.Inbounds
			b, _ = jsoniter.Marshal(tmplJson)
			err = WriteV2rayConfig(b)
			if err != nil {
				return
			}
		}

		_ = tools.KillAll("v2ray", true)
		//8秒等待v2ray启动
		startTime := time.Now()
		for {
			if time.Now().Sub(startTime) > 8*time.Second {
				return errors.New("请勿在Docker模式下频繁更换配置，请等待一段时间后再试")
			}
			<-time.After(100 * time.Millisecond)
			if IsV2RayProcessExists() {
				return nil
			}
		}
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
	case global.CommonMode:
		_ = tools.KillAll("v2ray", true)
		v2wd, _ := GetV2rayWorkingDir()
		v2ctlDir, _ := GetV2ctlDir()
		//cd到v2ctl的目录，防止找不到v2ctl
		wd, _ := os.Getwd()
		cmd := fmt.Sprintf("cd %v && nohup %v/v2ray --config=%v > %v/v2ray.log 2>&1 &", v2ctlDir, v2wd, GetConfigPath(), wd)
		out, err = exec.Command("sh", "-c", cmd).CombinedOutput()
		if err != nil {
			err = errors.New(err.Error() + string(out))
		}
	}
	if err != nil {
		return
	}
	<-time.After(100 * time.Millisecond)
	if !IsV2RayRunning() {
		return errors.New("v2ray启动失败")
	}
	return
}

func WriteV2rayConfig(content []byte) (err error) {
	err = ioutil.WriteFile(GetConfigPath(), content, os.ModeAppend)
	if err != nil {
		return errors.New("WriteV2rayConfig: " + err.Error())
	}
	return
}

/*更新v2ray配置并重启*/
func UpdateV2RayConfigAndRestart(vmessInfo *vmessInfo.VmessInfo) (err error) {
	//读配置，转换为v2ray配置并写入
	tmpl := NewTemplate()
	err = tmpl.FillWithVmessInfo(*vmessInfo)
	if err != nil {
		return
	}
	err = WriteV2rayConfig(tmpl.ToConfigBytes())
	if err != nil {
		return
	}
	err = RestartV2rayService()
	if err != nil {
		return
	}
	time.Sleep(100 * time.Millisecond)
	if configure.GetSettingNotNil().Transparent != configure.TransparentClose && CheckTProxySupported() == nil {
		_ = iptables.DeleteRules()
		err = iptables.WriteRules()
	}
	return
}

func UpdateV2rayWithConnectedServer() (err error) {
	cs := configure.GetConnectedServer()
	if cs == nil { //没有连接，把v2ray配置更新一下好了
		return pretendToStopV2rayService()
	}
	sr, err := cs.LocateServer()
	if err != nil {
		return
	}
	tmpl := NewTemplate()
	err = tmpl.FillWithVmessInfo(sr.VmessInfo)
	if err != nil {
		return
	}
	err = WriteV2rayConfig(tmpl.ToConfigBytes())
	if err != nil {
		return
	}
	if IsV2RayRunning() {
		err = RestartV2rayService()
		if configure.GetSettingNotNil().Transparent != configure.TransparentClose && CheckTProxySupported() == nil {
			time.Sleep(100 * time.Millisecond)
			_ = iptables.DeleteRules()
			err = iptables.WriteRules()
		}
	}
	return
}

/*清空inbounds规则来假停v2ray*/
func pretendToStopV2rayService() (err error) {
	tmplJson := NewTemplate()
	b, err := ioutil.ReadFile(GetConfigPath())
	if err != nil {
		return
	}
	err = jsoniter.Unmarshal(b, &tmplJson)
	if err != nil {
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

func StopV2rayService() (err error) {
	var out []byte
	switch global.ServiceControlMode {
	case global.DockerMode:
		return pretendToStopV2rayService()
	case global.CommonMode:
		err = tools.KillAll("v2ray", true)
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
	if IsV2RayRunning() {
		return errors.New("v2ray停止失败")
	}
	return
}

func RestartAndEnableV2rayService() (err error) {
	err = RestartV2rayService()
	if err != nil {
		return
	}
	return EnableV2rayService()
}

func StopAndDisableV2rayService() (err error) {
	err = StopV2rayService()
	if err != nil {
		return
	}
	return DisableV2rayService()
}
