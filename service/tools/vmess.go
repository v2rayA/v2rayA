package tools

import (
	"encoding/json"
	"errors"
	"log"
	"net/url"
	"regexp"
	"strings"
	"V2RayA/models"
)

/*
根据传入的 vmess://xxxxx 解析出NodeData
*/
func ResolveVmessURL(vmess string) (nodeData *models.NodeData, err error) {
	if len(vmess) < 8 || strings.ToLower(vmess[:8]) != "vmess://" {
		err = errors.New("this address is not begin with vmess://")
		return
	}
	var info models.VmessInfo
	// 进行base64解码，并unmarshal到VmessInfo上
	raw, err := Base64StdDecode(vmess[8:])
	if err != nil {
		// 不是json格式，尝试以vmess://BASE64(Security:ID@Add:Port)?remarks=Ps&obfsParam=Host&Path=Path&obfs=Net&tls=TLS解析
		var u *url.URL
		u, err = url.Parse(vmess)
		if err != nil {
			return
		}
		re := regexp.MustCompile(`.+:(.+)@(.+):(\d+)`)
		s := strings.Split(vmess[8:], "?")[0]
		s, err = Base64StdDecode(s)
		subMatch := re.FindStringSubmatch(s)
		if subMatch == nil {
			err = errors.New("无法识别的vmess链接")
			return
		}
		q := u.Query()
		info = models.VmessInfo{
			ID:   subMatch[1],
			Add:  subMatch[2],
			Port: subMatch[3],
			Ps:   q.Get("remarks"),
			Host: q.Get("obfsParam"),
			Path: q.Get("path"),
			Net:  q.Get("obfs"),
			Aid:  q.Get("aid"),
			TLS:  map[string]string{"1": "tls"}[q.Get("tls")],
			V:    "2",
		}
		if info.Net == "websocket" {
			info.Net = "ws"
		}
	} else {
		err = json.Unmarshal([]byte(raw), &info)
		if err != nil {
			return
		}
	}
	// 填充模板并处理结果
	tmpl := models.NewTemplate()
	err = tmpl.FillWithVmessInfo(info)

	nodeData = new(models.NodeData)
	b, err := json.Marshal(tmpl)
	nodeData.Config = string(b)
	nodeData.VmessInfo = info
	return
}

/*
根据传入的 ss://xxxxx 解析出NodeData
*/
func ResolveSSURL(vmess string) (nodeData *models.NodeData, err error) {
	if len(vmess) < 5 || strings.ToLower(vmess[:5]) != "ss://" {
		err = errors.New("this address is not begin with ss://")
		return
	}
	// 该函数尝试对ss://链接进行解析
	resolveFormat := func(content string) (subMatch []string, ok bool) {
		// 尝试按ss://method:password@server:port#name格式进行解析
		re := regexp.MustCompile(`(.+):(.+)@(\d+\.\d+\.\d+\.\d+):(\d+)(#.+)?`)
		subMatch = re.FindStringSubmatch(content)
		if len(subMatch) == 0 {
			// 尝试按ss://BASE64(method:password)@server:port#name格式进行解析
			re = regexp.MustCompile(`(.+)()@(\d+\.\d+\.\d+\.\d+):(\d+)(#.+)?`) //留个空组，确保subMatch长度统一
			subMatch = re.FindStringSubmatch(content)
			if len(subMatch) > 0 {
				raw, err := Base64StdDecode(subMatch[1])
				if err != nil {
					return
				}
				as := strings.Split(raw, ":")
				subMatch[1], subMatch[2] = as[0], as[1]
			}
		}
		if len(subMatch[5]) > 0 {
			subMatch[5] = subMatch[5][1:]
		}
		return subMatch, len(subMatch) > 0
	}
	content := vmess[5:]
	var (
		subMatch []string
		ok       bool
	)
	// 尝试解析ss://链接，失败则先base64解码
	if subMatch, ok = resolveFormat(content); !ok {
		// 进行base64解码，并unmarshal到VmessInfo上
		content, err = Base64StdDecode(content)
		if err != nil {
			return
		}
		subMatch, ok = resolveFormat(content)
	}
	if !ok {
		err = errors.New("不是合法的ss URL")
		return
	}
	log.Println(content, subMatch)
	info := models.VmessInfo{
		Protocol: "shadowsocks",
		Type:     subMatch[1],
		ID:       subMatch[2],
		Add:      subMatch[3],
		Port:     subMatch[4],
		Ps:       subMatch[5],
	}
	log.Println(info)
	// 填充模板并处理结果
	tmpl := models.NewTemplate()
	err = tmpl.FillWithVmessInfo(info)

	nodeData = new(models.NodeData)
	b, err := json.Marshal(tmpl)
	nodeData.Config = string(b)
	nodeData.VmessInfo = info
	return
}
