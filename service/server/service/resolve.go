package service

import (
	"github.com/json-iterator/go"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/core/vmessInfo"
	"github.com/v2rayA/v2rayA/infra/nodeData"
	"log"
	"net/url"
	"regexp"
	"strings"
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
		raw, err = common.Base64URLDecode(vmess[8:])
	}
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
		if err != nil {
			s, err = common.Base64URLDecode(s)
		}
		subMatch := re.FindStringSubmatch(s)
		if subMatch == nil {
			err = newError("unrecognized vmess address")
			return
		}
		q := u.Query()
		ps := q.Get("remarks")
		if ps == "" {
			ps = q.Get("remark")
		}
		obfs := q.Get("obfs")
		obfsParam := q.Get("obfsParam")
		path := q.Get("path")
		if obfs == "kcp" || obfs == "mkcp" {
			m := make(map[string]string)
			//迎合v2rayN的格式定义
			_ = jsoniter.Unmarshal([]byte(obfsParam), &m)
			path = m["seed"]
			obfsParam = ""
		}
		aid := q.Get("alterId")
		if aid == "" {
			aid = q.Get("aid")
		}
		info = vmessInfo.VmessInfo{
			ID:            subMatch[1],
			Add:           subMatch[2],
			Port:          subMatch[3],
			Ps:            ps,
			Host:          obfsParam,
			Path:          path,
			Net:           obfs,
			Aid:           aid,
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
		info.Aid = "1"
	}
	data = new(nodeData.NodeData)
	data.VmessInfo = info
	return
}

/*
根据传入的 vless://xxxxx 解析出NodeData
*/
func ResolveVlessURL(vless string) (data *nodeData.NodeData, err error) {
	if !strings.HasPrefix(vless, "vless://") {
		err = newError("this address is not begin with vless://")
		return
	}
	u, err := url.Parse(vless)
	if err != nil {
		return
	}
	data = new(nodeData.NodeData)
	data.VmessInfo = vmessInfo.VmessInfo{
		Ps:       u.Fragment,
		Add:      u.Hostname(),
		Port:     u.Port(),
		ID:       u.User.String(),
		Net:      u.Query().Get("type"),
		Type:     u.Query().Get("headerType"),
		Host:     u.Query().Get("sni"),
		Path:     u.Query().Get("path"),
		TLS:      u.Query().Get("security"),
		Flow:     u.Query().Get("flow"),
		Protocol: "vless",
	}
	if data.VmessInfo.Net == "" {
		data.VmessInfo.Net = "tcp"
	}
	if data.VmessInfo.Type == "" {
		data.VmessInfo.Type = "none"
	}
	if data.VmessInfo.Host == "" {
		data.VmessInfo.Host = u.Query().Get("host")
	}
	if data.VmessInfo.TLS == "" {
		data.VmessInfo.TLS = "none"
	}
	if data.VmessInfo.Flow == "" {
		data.VmessInfo.Flow = "xtls-rprx-direct"
	}
	if data.VmessInfo.Type == "mkcp" || data.VmessInfo.Type == "kcp" {
		data.VmessInfo.Path = u.Query().Get("seed")
	}
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
	resolveFormat := func(content string) (v *vmessInfo.VmessInfo, ok bool) {
		// 尝试按ss://BASE64(method:password)@server:port/?plugin=xxxx#name格式进行解析
		u, err := url.Parse(content)
		if err != nil {
			return nil, false
		}
		username := u.User.String()
		username, _ = common.Base64URLDecode(username)
		arr := strings.Split(username, ":")
		if len(arr) != 2 {
			return nil, false
		}
		method := arr[0]
		password := arr[1]
		var obfs, path, host string
		plugin := u.Query().Get("plugin")
		arr = strings.Split(plugin, ";")
		for i := 1; i < len(arr); i++ {
			a := strings.Split(arr[i], "=")
			switch a[0] {
			case "obfs":
				obfs = a[1]
			case "obfs-path":
				path = a[1]
			case "obfs-host":
				host = a[1]
			}
		}
		return &vmessInfo.VmessInfo{
			Net:      method,
			ID:       password,
			Add:      u.Hostname(),
			Port:     u.Port(),
			Ps:       u.Fragment,
			Type:     obfs,
			Path:     path,
			Host:     host,
			Protocol: "ss",
		}, true
	}
	var (
		v  *vmessInfo.VmessInfo
		ok bool
	)
	content := u
	// 尝试解析ss://链接，失败则先base64解码
	if v, ok = resolveFormat(content); !ok {
		// 进行base64解码，并unmarshal到VmessInfo上
		t := content[5:]
		var l, r string
		if ind := strings.Index(t, "#"); ind > -1 {
			l = t[:ind]
			r = t[ind+1:]
		} else {
			l = t
		}
		l, err = common.Base64StdDecode(l)
		if err != nil {
			l, err = common.Base64URLDecode(l)
			if err != nil {
				return
			}
		}
		t = "ss://" + l
		if len(r) > 0 {
			t += "#" + r
		}
		v, ok = resolveFormat(t)
	}
	if !ok {
		err = newError("unrecognized ss address")
		return
	}
	// 填充模板并处理结果
	data = new(nodeData.NodeData)
	data.VmessInfo = *v
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
	if !strings.HasPrefix(u, "trojan://") && !strings.HasPrefix(u, "trojan-go://") {
		err = newError("this address is not begin with trojan:// or trojan-go://")
		return
	}
	t, err := url.Parse(u)
	if err != nil {
		err = newError("invalid trojan format")
		return
	}
	allowInsecure := t.Query().Get("allowInsecure")
	data = new(nodeData.NodeData)
	sni := t.Query().Get("peer")
	if sni == "" {
		sni = t.Query().Get("sni")
	}
	data.VmessInfo = vmessInfo.VmessInfo{
		Ps:            t.Fragment,
		Add:           t.Hostname(),
		Port:          t.Port(),
		ID:            t.User.String(),
		Host:          sni,
		AllowInsecure: allowInsecure == "1" || allowInsecure == "true",
		Protocol:      "trojan",
	}
	if t.Scheme == "trojan-go" {
		data.VmessInfo.Protocol = "trojan-go"
		data.VmessInfo.Type = t.Query().Get("encryption")
		data.VmessInfo.Host = sni + "," + t.Query().Get("host")
		data.VmessInfo.Path = t.Query().Get("path")
		data.VmessInfo.Net = t.Query().Get("type")
		data.VmessInfo.TLS = "tls"
	}
	return
}
func ResolvePingTunnelURL1(u string) (data *nodeData.NodeData, err error) {
	if len(u) < 13 || strings.ToLower(u[:13]) != "pingtunnel://" {
		err = newError("this address is not begin with pingtunnel://")
		return
	}
	u = u[13:]
	u, err = common.Base64StdDecode(u)
	if err != nil {
		log.Println(u)
		err = newError().Base(err)
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
		err = newError().Base(err)
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

func ResolvePingTunnelURL2(u string) (data *nodeData.NodeData, err error) {
	if !strings.HasPrefix(u, "ping-tunnel://") {
		err = newError("this address is not begin with pingtunnel://")
		return
	}
	U, err := url.Parse(u)
	if err != nil {
		return
	}
	data = new(nodeData.NodeData)
	data.VmessInfo = vmessInfo.VmessInfo{
		Ps:       U.Fragment,
		Add:      U.Host,
		ID:       U.User.String(),
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
	} else if strings.HasPrefix(u, "vless://") {
		n, err = ResolveVlessURL(u)
	} else if strings.HasPrefix(u, "ss://") {
		n, err = ResolveSSURL(u)
	} else if strings.HasPrefix(u, "ssr://") {
		n, err = ResolveSSRURL(u)
	} else if strings.HasPrefix(u, "pingtunnel://") {
		n, err = ResolvePingTunnelURL1(u)
	} else if strings.HasPrefix(u, "ping-tunnel://") {
		n, err = ResolvePingTunnelURL2(u)
	} else if strings.HasPrefix(u, "trojan://") || strings.HasPrefix(u, "trojan-go://") {
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
