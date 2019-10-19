package tools

import (
	"bytes"
	"encoding/base64"
	"net/http"
	"strings"
	"v2rayW/models"
)

func ResolveSubscription(source string) (infos []*models.NodeData, err error) {
	// get请求source
	res, err := http.Get(source)
	if err != nil {
		return
	}
	defer res.Body.Close()
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(res.Body)
	// base64解码, raw是多行vmess
	raw, err := base64.StdEncoding.DecodeString(buf.String())
	if err != nil {
		return
	}
	// 切分raw
	rows := strings.Split(string(raw), "\n")
	// 解析
	infos = make([]*models.NodeData, 0)
	for _, row := range rows {
		var data *models.NodeData
		data, err = ResolveVmessURL(row)
		if err != nil {
			return
		}
		infos = append(infos, data)
	}
	return
}
