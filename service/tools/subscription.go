package tools

import (
	"V2RayA/models"
	"bytes"
	"log"
	"net/http"
	"strings"
)

func ResolveSubscription(source string) (infos []*models.NodeData, err error) {
	// get请求source
	res, err := http.Get(source)
	if err != nil {
		return
	}
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(res.Body)
	defer res.Body.Close()
	// base64解码, raw是多行vmess
	raw, err := Base64StdDecode(buf.String())
	if err != nil {
		return
	}
	// 切分raw
	rows := strings.Split(strings.TrimSpace(raw), "\n")
	// 解析
	infos = make([]*models.NodeData, 0)
	for _, row := range rows {
		var data *models.NodeData
		data, err = ResolveURL(row)
		if err != nil {
			log.Println(row, err)
			err = nil
			continue
		}
		infos = append(infos, data)
	}
	return
}
