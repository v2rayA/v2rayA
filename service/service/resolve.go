package service

import (
	"V2RayA/model/nodeData"
	tmpl "V2RayA/model/v2rayTmpl"
	"V2RayA/model/vmessInfo"
	"V2RayA/tools"
	"github.com/json-iterator/go"
	"errors"
	"net/url"
	"regexp"
	"strings"
)

/*
根据传入的 vmess://xxxxx 解析出NodeData
*/
func ResolveVmessURL(vmess string) (data *nodeData.NodeData, err error) {
	if len(vmess) < 8 || strings.ToLower(vmess[:8]) != "vmess://" {
		err = errors.New("this address is not begin with vmess://")
		return
	}
	var info vmessInfo.VmessInfo
	// 进行base64解码，并unmarshal到VmessInfo上
	raw, err := tools.Base64StdDecode(vmess[8:])
	if err != nil {
		// 不是json格式，尝试以vmess://BASE64(Security:ID@Add:Port)?remarks=Ps&obfsParam=Host&Path=Path&obfs=Net&tls=TLS解析
		var u *url.URL
		u, err = url.Parse(vmess)
		if err != nil {
			return
		}
		re := regexp.MustCompile(`.+:(.+)@(.+):(\d+)`)
		s := strings.Split(vmess[8:], "?")[0]
		s, err = tools.Base64StdDecode(s)
		subMatch := re.FindStringSubmatch(s)
		if subMatch == nil {
			err = errors.New("无法识别的vmess链接")
			return
		}
		q := u.Query()
		info = vmessInfo.VmessInfo{
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
		err = jsoniter.Unmarshal([]byte(raw), &info)
		if err != nil {
			return
		}
	}
	// 填充模板并处理结果
	t := tmpl.NewTemplate()
	err = t.FillWithVmessInfo(info)
	data = new(nodeData.NodeData)
	b := t.ToConfigBytes()
	data.Config = string(b)
	data.VmessInfo = info
	return
}

/*
根据传入的 ss://xxxxx 解析出NodeData
*/
func ResolveSSURL(vmess string) (data *nodeData.NodeData, err error) {
	if len(vmess) < 5 || strings.ToLower(vmess[:5]) != "ss://" {
		err = errors.New("this address is not begin with ss://")
		return
	}
	// 该函数尝试对ss://链接进行解析
	resolveFormat := func(content string) (subMatch []string, ok bool) {
		// 尝试按ss://method:password@server:port#name格式进行解析
		re := regexp.MustCompile(`(.+):(.+)@(.+?):(\d+)(#.+)?`)
		subMatch = re.FindStringSubmatch(content)
		if len(subMatch) == 0 {
			// 尝试按ss://BASE64(method:password)@server:port#name格式进行解析
			re = regexp.MustCompile(`(.+)()@(.+?):(\d+)(#.+)?`) //留个空组，确保subMatch长度统一
			subMatch = re.FindStringSubmatch(content)
			if len(subMatch) > 0 {
				raw, err := tools.Base64StdDecode(subMatch[1])
				if err != nil {
					return
				}
				as := strings.Split(raw, ":")
				subMatch[1], subMatch[2] = as[0], as[1]
			}
		}
		if subMatch == nil {
			return
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
		content, err = tools.Base64StdDecode(content)
		if err != nil {
			return
		}
		subMatch, ok = resolveFormat(content)
	}
	if !ok {
		err = errors.New("不是合法的ss URL")
		return
	}
	info := vmessInfo.VmessInfo{
		Protocol: "shadowsocks",
		Net:      subMatch[1],
		ID:       subMatch[2],
		Add:      subMatch[3],
		Port:     subMatch[4],
		Ps:       subMatch[5],
	}
	// 填充模板并处理结果
	t := tmpl.NewTemplate()
	err = t.FillWithVmessInfo(info)

	data = new(nodeData.NodeData)
	b := t.ToConfigBytes()
	data.Config = string(b)
	data.VmessInfo = info
	return
}
func ResolveURL(u string) (n *nodeData.NodeData, err error) {
	if strings.HasPrefix(u, "vmess://") {
		n, err = ResolveVmessURL(u)
	} else if strings.HasPrefix(u, "ss://") {
		n, err = ResolveSSURL(u)
	} else {
		err = errors.New("不支持该协议，目前只支持ss和vmess协议")
		return
	}
	if err != nil {
		return
	}
	return
}
