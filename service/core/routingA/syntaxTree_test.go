package routingA

import (
	"log"
	"testing"
)

func TestGenerateSyntaxTree(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	tests := []string{
		`outbound:httpout=http(address:127.0.0.1,port:8080,user:'my-username',pass:'my-password')`,
		`outbound:socksout=socks(address:127.0.0.1,port:10800,user:"my-username",pass:"my-password")`,
		`outbound:test=socks(127.0.0.1,port:10800,user:"my-username",pass:"my-password")`,
		`default:httpout`,
		`domain(domain:v2raya.mzz.pub)->socksout`,
		`domain(full:dns.google)->proxy`,
		`domain(contains:.google.)->proxy`,
		`ip(127.0.0.1)->direct`,
		`ip(192.168.0.0/16)->direct`,
		`# one line comment`,
		``,
		``,
		`extern(ip,geoip,private)->direct`,
		`extern(domain,geosite,category-ads)->block`,
		`ip(8.8.8.8)&&network(tcp,udp)&&port(1-1023,8443)->proxy`,
		``,
		`ip(1.1.1.1)&&protocol(http)->direct`,
		`
# 自定义outbound
outbound: httpout = http(address: 127.0.0.1, port: 8080, user: 'my-username', pass: 'my-password')
outbound: socksout = socks(address: 127.0.0.1, port: 10800, user: "my-username", pass: "my-password")
outbound:   test=   socks ( 127.0.0.1,   port: 10800, user: "my-username",pass:"my-password" )
outbound:   test2=   socks ( test,   port: 10800, user: "my-username",pass:"my-password" )

# 设置默认outbound，不填则为proxy
default  : httpout

# 缺省情况下有proxy、block、direct三个outbound tag

# 普通的路由规则书写方式
domain(domain: v2raya.mzz.pub) -> socksout
domain(full:   dns.google)   ->   proxy
domain(contains: .google.) -> proxy
ip(127.0.0.1) -> direct
ip(192.168.0.0/16) -> direct
extern(ip, geoip, private) -> direct
extern(domain, geosite, category-ads) -> block

# and规则
ip(8.8.8.8) && network(tcp, udp) && port(1-1023, 8443) -> proxy
ip(1.1.1.1) && protocol(http) -> direct
`,
	}
	for _, test := range tests {
		S, err := generateSyntaxTree(test)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(string(S.sym), "=>", S.val)
	}
	t.Log("all tests passed")
}
