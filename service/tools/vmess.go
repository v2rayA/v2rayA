package tools

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"v2rayW/models"
)

func ResolveVmessURL(vmess string) (nodeData *models.NodeData, err error) {
	if !strings.HasPrefix(vmess, "vmess://") {
		err = errors.New("this address is not begin with vmess://")
		return
	}
	// 进行base64解码，并unmarshal到VmessInfo上
	raw, err := base64.StdEncoding.DecodeString(vmess[8:])
	if err != nil {
		return
	}
	var info models.VmessInfo
	err = json.Unmarshal(raw, &info)
	if err != nil {
		return
	}
	// 填充模板并处理结果
	tmpl := models.NewTemplate()
	err = tmpl.ImportFromURL(info)

	nodeData = new(models.NodeData)
	b, err := json.Marshal(tmpl)
	nodeData.Config = string(b)
	nodeData.VmessInfo = info
	return
}
