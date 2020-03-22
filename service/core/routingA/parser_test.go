package routingA

import (
	"testing"
)

func TestParse(t *testing.T) {
	program := `
# 自定义outbound
outbound: httpout = http(address: 127.0.0.1, port: 8080, user: 'my-username', pass: 'my-password')
outbound: socksout = socks(address: 127.0.0.1, port: 10800, user: "my-username", pass: "my-password")

# 设置默认outbound，不填则为proxy
default: httpout

# 缺省情况下有proxy、block、direct三个outbound tag

# 域名规则
domain(domain: v2raya.mzz.pub) -> socksout
domain(full: dns.google) -> proxy
domain(contains: facebook) -> proxy
# IP规则
ip(127.0.0.1) -> direct
ip(192.168.0.0/16) -> direct

# 包含多个域名
domain(contains: google, domain: www.twitter.com, domain: mzz.pub) -> proxy
# 包含多个IP
ip(1.2.3.4, 9.9.9.9, 223.5.5.5) -> direct

# 扩展文件规则
extern(ip, geoip, private) -> direct
extern(domain, geosite, category-ads) -> block
# 也可写作
ip(geoip: private) -> direct
domain(geosite: category-ads) -> block

# and规则
domain(geosite: cn, geosite:speedtest) && port(80, 443) -> direct
ip(8.8.8.8) && network(tcp, udp) && port(1-1023, 8443) -> proxy
ip(1.1.1.1) && protocol(http) -> direct
`
	rules, err := Parse(program)
	if err != nil {
		t.Fatal(err)
	}
	for _, rule := range rules {
		switch rule := rule.(type) {
		case Routing:
			t.Log(rule.And, rule.Out)
		case Define:
			switch v := rule.Value.(type) {
			case string:
				t.Log(rule.Name, v)
			case Function:
				t.Log(rule.Name, v)
			}
		}
	}
}
