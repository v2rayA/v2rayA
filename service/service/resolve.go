package service

import (
	"github.com/json-iterator/go"
	"log"
	"net/url"
	"regexp"
	"strings"
	"v2rayA/common"
	"v2rayA/core/nodeData"
	"v2rayA/core/vmessInfo"
)

/*
根据传入的 vmess://xxxxx 解析出NodeData
*/
func ResolveVmessURL(vmess string) (data *nodeData.NodeData, err error) {
	if len(vmess) < 8 || strings.ToLower(vmess[:8]) != "vmess://" {
		err = newError("this address is not begin with vmess://")
		return
	}
	var info vmessInfo.VmessInfo
	// 进行base64解码，并unmarshal到VmessInfo上
	raw, err := common.Base64StdDecode(vmess[8:])
	if err != nil {
		// 不是json格式，尝试以vmess://BASE64(Security:ID@Add:Port)?remarks=Ps&obfsParam=Host&Path=Path&obfs=Net&tls=TLS解析
		var u *url.URL
		u, err = url.Parse(vmess)
		if err != nil {
			return
		}
		re := regexp.MustCompile(`.*:(.+)@(.+):(\d+)`)
		s := strings.Split(vmess[8:], "?")[0]
		s, err = common.Base64StdDecode(s)
		subMatch := re.FindStringSubmatch(s)
		if subMatch == nil {
			err = newError("unrecognized vmess address")
			return
		}
		q := u.Query()
		info = vmessInfo.VmessInfo{
			ID:            subMatch[1],
			Add:           subMatch[2],
			Port:          subMatch[3],
			Ps:            q.Get("remarks"),
			Host:          q.Get("obfsParam"),
			Path:          q.Get("path"),
			Net:           q.Get("obfs"),
			Aid:           q.Get("aid"),
			TLS:           map[string]string{"1": "tls"}[q.Get("tls")],
			V:             "2",
			AllowInsecure: false,
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
	// 对错误vmess进行力所能及的修正
	if strings.HasPrefix(info.Host, "/") && info.Path == "" {
		info.Path = info.Host
		info.Host = ""
	}
	if info.Aid == "" {
		info.Aid = "6"
	}
	data = new(nodeData.NodeData)
	data.VmessInfo = info
	return
}

/*
根据传入的 ss://xxxxx 解析出NodeData
*/
func ResolveSSURL(u string) (data *nodeData.NodeData, err error) {
	if len(u) < 5 || strings.ToLower(u[:5]) != "ss://" {
		err = newError("this address is not begin with ss://")
		return
	}
	// 该函数尝试对ss://链接进行解析
	resolveFormat := func(content string) (subMatch []string, ok bool) {
		// 尝试按ss://method:password@server:port#name格式进行解析
		content = strings.TrimSuffix(content, "#")
		re := regexp.MustCompile(`(.+):(.+)@(.+?):(\d+)(#.+)?`)
		subMatch = re.FindStringSubmatch(content)
		if len(subMatch) == 0 {
			// 尝试按ss://BASE64(method:password)@server:port#name格式进行解析
			re = regexp.MustCompile(`(.+)()@(.+?):(\d+)(#.+)?`) //留个空组，确保subMatch长度统一
			subMatch = re.FindStringSubmatch(content)
			if len(subMatch) > 0 {
				raw, err := common.Base64StdDecode(subMatch[1])
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
	content := u[5:]
	//看是不是有#，有的话说明name没有被base64
	sp := strings.Split(content, "#")
	var name string
	if len(sp) == 2 {
		content = sp[0]
		var e error
		name, e = common.Base64URLDecode(sp[1])
		if e != nil {
			name, e = common.Base64StdDecode(sp[1])
		}
		_name, e := url.QueryUnescape(name)
		if e == nil {
			name = _name
		}
	}
	var (
		subMatch []string
		ok       bool
	)
	// 尝试解析ss://链接，失败则先base64解码
	if subMatch, ok = resolveFormat(content); !ok {
		// 进行base64解码，并unmarshal到VmessInfo上
		content, err = common.Base64StdDecode(content)
		if err != nil {
			content, err = common.Base64URLDecode(content)
			if err != nil {
				return
			}
		}
		subMatch, ok = resolveFormat(content)
	}
	if !ok {
		err = newError("unrecognized ss address")
		return
	}
	info := vmessInfo.VmessInfo{
		Protocol: "ss",
		Net:      subMatch[1],
		ID:       subMatch[2],
		Add:      subMatch[3],
		Port:     subMatch[4],
		Ps:       subMatch[5],
	}
	if len(name) > 0 {
		info.Ps = name
	}
	// 填充模板并处理结果
	data = new(nodeData.NodeData)
	//t, err := v2ray.NewTemplateFromVmessInfo(info)
	//if err == nil {
	//	b := t.ToConfigBytes()
	//	data.Config = string(b)
	//}
	data.VmessInfo = info
	return
}

/*
根据传入的 ss://xxxxx 解析出NodeData
*/
func ResolveSSRURL(u string) (data *nodeData.NodeData, err error) {
	if len(u) < 6 || strings.ToLower(u[:6]) != "ssr://" {
		err = newError("this address is not begin with ssr://")
		return
	}
	// 该函数尝试对ssr://链接进行解析
	resolveFormat := func(content string) (v vmessInfo.VmessInfo, ok bool) {
		arr := strings.Split(content, "/?")
		if strings.Contains(content, ":") && len(arr) < 2 {
			content += "/?remarks=&protoparam=&obfsparam="
			arr = strings.Split(content, "/?")
		} else if len(arr) != 2 {
			return v, false
		}
		pre := strings.Split(arr[0], ":")
		if len(pre) > 6 {
			//如果长度多于6，说明host中包含字符:，重新合并前几个分组到host去
			pre[len(pre)-6] = strings.Join(pre[:len(pre)-5], ":")
			pre = pre[len(pre)-6:]
		} else if len(pre) < 6 {
			return v, false
		}
		q, err := url.ParseQuery(arr[1])
		if err != nil {
			return v, false
		}
		pswd, _ := common.Base64URLDecode(pre[5])
		add, _ := common.Base64URLDecode(pre[0])
		remarks, _ := common.Base64URLDecode(q.Get("remarks"))
		protoparam, _ := common.Base64URLDecode(q.Get("protoparam"))
		obfsparam, _ := common.Base64URLDecode(q.Get("obfsparam"))
		v = vmessInfo.VmessInfo{
			Ps:       remarks,
			Add:      add,
			Port:     pre[1],
			ID:       pswd,
			Net:      pre[3],
			Type:     pre[2],
			Host:     protoparam,
			Path:     obfsparam,
			TLS:      pre[4],
			Protocol: "ssr",
		}
		return v, true
	}
	content := u[6:]
	var (
		info vmessInfo.VmessInfo
		ok   bool
	)
	// 尝试解析ssr://链接，失败则先base64解码
	if info, ok = resolveFormat(content); !ok {
		// 进行base64解码，并unmarshal到VmessInfo上
		content, err = common.Base64StdDecode(content)
		if err != nil {
			content, err = common.Base64URLDecode(content)
			if err != nil {
				return
			}
		}
		info, ok = resolveFormat(content)
	}
	if !ok {
		err = newError("unrecognized ssr address")
		return
	}
	// 填充模板并处理结果
	data = new(nodeData.NodeData)
	data.VmessInfo = info
	return
}

func ResolveTrojanURL(u string) (data *nodeData.NodeData, err error) {
	//	trojan://password@server:port#escape(remarks)
	if !strings.HasPrefix(u, "trojan://") {
		err = newError("this address is not begin with trojan://")
		return
	}
	t, err := url.Parse(u)
	if err != nil {
		err = newError("invalid trojan format")
		return
	}
	allowInsecure := t.Query().Get("allowInsecure")
	data = new(nodeData.NodeData)
	data.VmessInfo = vmessInfo.VmessInfo{
		Ps:            t.Fragment,
		Add:           t.Hostname(),
		Port:          t.Port(),
		ID:            t.User.String(),
		Host:          t.Query().Get("peer"),
		AllowInsecure: allowInsecure == "1" || allowInsecure == "true",
		Protocol:      "trojan",
	}
	log.Println(data.VmessInfo)
	return
}
func ResolvePingTunnelURL(u string) (data *nodeData.NodeData, err error) {
	if len(u) < 13 || strings.ToLower(u[:13]) != "pingtunnel://" {
		err = newError("this address is not begin with pingtunnel://")
		return
	}
	u = u[13:]
	u, err = common.Base64StdDecode(u)
	if err != nil {
		log.Println(u)
		err = newError(err)
		return
	}
	arr := strings.Split(u, "#")
	var ps string
	if len(arr) == 2 {
		ps, _ = url.QueryUnescape(arr[1])
	}
	u = arr[0]
	re := regexp.MustCompile(`(.+):(.+)`)
	subMatch := re.FindStringSubmatch(u)
	if len(subMatch) < 3 {
		return nil, newError("wrong format of pingtunnel")
	}
	data = new(nodeData.NodeData)
	passwd, err := common.Base64URLDecode(subMatch[2])
	if err != nil {
		log.Println(subMatch[2])
		err = newError(err)
		return
	}
	data.VmessInfo = vmessInfo.VmessInfo{
		Ps:       ps,
		Add:      subMatch[1],
		ID:       passwd,
		Protocol: "pingtunnel",
	}
	return
}

var ErrorEmptyAddress = newError("ResolveURL error: empty address")

func ResolveURL(u string) (n *nodeData.NodeData, err error) {
	u = strings.TrimSpace(u)
	if len(u) <= 0 {
		err = ErrorEmptyAddress
		return
	}
	if strings.HasPrefix(u, "vmess://") {
		n, err = ResolveVmessURL(u)
	} else if strings.HasPrefix(u, "ss://") {
		n, err = ResolveSSURL(u)
	} else if strings.HasPrefix(u, "ssr://") {
		n, err = ResolveSSRURL(u)
	} else if strings.HasPrefix(u, "pingtunnel://") {
		n, err = ResolvePingTunnelURL(u)
	} else if strings.HasPrefix(u, "trojan://") {
		n, err = ResolveTrojanURL(u)
	} else {
		err = newError("not supported protocol. we only support ss, ssr and vmess now: " + u)
		return
	}
	if err != nil {
		return
	}
	return
}
