package tools

import (
	"V2RayA/global"
	"V2RayA/models/v2ray"
	"V2RayA/models/v2rayTmpl"
	"V2RayA/models/vmessInfo"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

func RestartV2rayService() (err error) {
	switch global.ServiceControlMode {
	case v2ray.Docker:
		_, err = exec.Command("sh", "-c", "killall -9 v2ray").CombinedOutput()
		//8秒等待v2ray启动
		startTime := time.Now()
		for {
			if time.Now().Sub(startTime) > 8*time.Second {
				return errors.New("请勿在Docker模式下频繁更换配置，请等待一段时间后再试")
			}
			<-time.After(200 * time.Millisecond)
			if IsV2RayRunning() {
				return nil
			}
		}
	case v2ray.Service:
		_, err = exec.Command("sh", "-c", "service v2ray restart").CombinedOutput()
	case v2ray.Systemctl:
		_, err = exec.Command("sh", "-c", "systemctl restart v2ray").Output()
	case v2ray.Common:
		_, _ = exec.Command("sh", "-c", "killall -9 v2ray").CombinedOutput()
		_, err = exec.Command("sh", "-c", "v2ray --config=/etc/v2ray/config.json").CombinedOutput()
	}
	if err != nil {
		if global.ServiceControlMode == v2ray.Docker {
			log.Println("建议检查killall命令是否可用")
		}
		return
	}
	if !IsV2RayRunning() {
		return errors.New("v2ray启动失败")
	}
	return
}

func StopV2rayService() (err error) {
	switch global.ServiceControlMode {
	case v2ray.Docker: //docker中无需stop service
		return nil
	case v2ray.Common:
		_, err = exec.Command("sh", "-c", "killall -9 v2ray").CombinedOutput()
	case v2ray.Service:
		_, err = exec.Command("sh", "-c", "service v2ray stop").CombinedOutput()
	case v2ray.Systemctl:
		_, err = exec.Command("sh", "-c", "systemctl stop v2ray").CombinedOutput()
	}
	if IsV2RayRunning() {
		return errors.New("v2ray停止失败")
	}
	return
}

func EnableV2rayService() (err error) {
	switch global.ServiceControlMode {
	case v2ray.Docker, v2ray.Common: //docker, common中无需enable service
	case v2ray.Service:
		_, err = exec.Command("sh", "-c", "update-rc.d v2ray enable").CombinedOutput()
	case v2ray.Systemctl:
		_, err = exec.Command("sh", "-c", "systemctl enable v2ray").Output()
	}
	return
}

func DisableV2rayService() (err error) {
	switch global.ServiceControlMode {
	case v2ray.Docker, v2ray.Common: //docker, common中无需disable service
	case v2ray.Service:
		_, err = exec.Command("sh", "-c", "update-rc.d v2ray disable").CombinedOutput()
	case v2ray.Systemctl:
		_, err = exec.Command("sh", "-c", "systemctl disable v2ray").Output()
	}
	return
}

func WriteV2rayConfig(content []byte) (err error) {
	return ioutil.WriteFile("/etc/v2ray/config.json", content, os.ModeAppend)
}

func IsV2RayRunning() bool {
	switch global.ServiceControlMode {
	case v2ray.Docker, v2ray.Common:
		out, err := exec.Command("sh", "-c", "ps -ef|grep v2ray").CombinedOutput()
		return err == nil && strings.Contains(string(out), "v2ray -config=")
	case v2ray.Service:
		out, err := exec.Command("sh", "-c", "service v2ray status|head -n 5|grep running").CombinedOutput()
		if err != nil || strings.Contains(string(out), "not running") {
			return false
		}
	case v2ray.Systemctl:
		out, err := exec.Command("sh", "-c", "systemctl status v2ray|head -n 5|grep running").Output()
		return err == nil && len(out) > 0
	}
	return true
}

/*更新v2ray配置并重启*/
func UpdateV2RayConfigAndRestart(vmessInfo *vmessInfo.VmessInfo) (err error) {
	//读配置，转换为v2ray配置并写入
	tmpl := v2rayTmpl.NewTemplate()
	err = tmpl.FillWithVmessInfo(*vmessInfo)
	if err != nil {
		return
	}
	err = WriteV2rayConfig(tmpl.ToConfigBytes())
	if err != nil {
		return
	}
	return RestartV2rayService()
}

/*清空inbounds规则来假停v2ray, 不重启服务*/
func PretendToStopV2rayService() (err error) {
	tmplJson := v2rayTmpl.Template{}
	b, err := ioutil.ReadFile("/etc/v2ray/config.json")
	if err != nil {
		return
	}
	err = json.Unmarshal(b, &tmplJson)
	if err != nil {
		return
	}
	tmplJson.Inbounds = make([]v2rayTmpl.Inbound, 0)
	b, _ = json.Marshal(tmplJson)
	return WriteV2rayConfig(b)
}
