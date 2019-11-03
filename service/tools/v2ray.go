package tools

import (
	"V2RayA/global"
	"V2RayA/models/v2rayTmpl"
	"V2RayA/models/vmessInfo"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

func RestartV2rayService() (err error) {
	if global.IsInDocker {
		_, err = exec.Command("sh", "-c", "killall -9 v2ray").CombinedOutput()
		return
	}
	_, err = exec.Command("sh", "-c", "service v2ray restart").CombinedOutput()
	if err != nil {
		log.Println(1, err)
		_, err = exec.Command("sh", "-c", "systemctl restart v2ray").Output()
	}
	if err != nil {
		return
	}
	if !IsV2RayRunning() {
		return errors.New("v2ray启动失败")
	}
	return
}

func StopV2rayService() (err error) { //TODO: 在stop的地方替换config，并在启动时还原
	if global.IsInDocker {
		return nil //docker中无法stop service
	}
	out, err := exec.Command("sh", "-c", "service v2ray stop").CombinedOutput()
	if err != nil {
		log.Println(2, string(out), err)
		_, err = exec.Command("sh", "-c", "systemctl stop v2ray").CombinedOutput()
	}
	if err != nil {
		log.Println(2, string(out), err)
		_, err = exec.Command("sh", "-c", "killall -9 v2ray").CombinedOutput()
	}
	if IsV2RayRunning() {
		return errors.New("v2ray停止失败")
	}
	return
}

func EnableV2rayService() (err error) {
	if global.IsInDocker {
		return nil //docker中无需enable service
	}
	_, err = exec.Command("sh", "-c", "update-rc.d v2ray enable").CombinedOutput()
	if err != nil {
		log.Println(3, err)
		_, err = exec.Command("sh", "-c", "systemctl enable v2ray").Output()
	}
	return
}

func DisableV2rayService() (err error) {
	if global.IsInDocker {
		return nil //docker中无需disable service
	}
	_, err = exec.Command("sh", "-c", "update-rc.d v2ray disable").CombinedOutput()
	if err != nil {
		log.Println(4, err)
		_, err = exec.Command("sh", "-c", "systemctl disable v2ray").Output()
	}
	return
}

func WriteV2rayConfig(content []byte) (err error) {
	return ioutil.WriteFile("/etc/v2ray/config.json", content, os.ModeAppend)
}

func IsV2RayRunning() bool {
	if global.IsInDocker {
		out, err := exec.Command("sh", "-c", "ps|grep v2ray").CombinedOutput()
		return err == nil && strings.Contains(string(out), "v2ray -config=")
	}
	out, err := exec.Command("sh", "-c", "service v2ray status|head -n 5|grep running").CombinedOutput()
	if strings.Contains(string(out), "not running") {
		return false
	}
	if err != nil {
		log.Println(5, string(out), err)
		out, err = exec.Command("sh", "-c", "systemctl status v2ray|head -n 5|grep running").Output()
	}
	return err == nil && len(out) > 0
}
func UpdateV2RayConfig(vmessInfo *vmessInfo.VmessInfo) (err error) {
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
