package routingA

import (
	"log"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
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
	}
	for i, test := range tests {
		Parse(test)
		t.Log("test", i,"passed")
	}
	Parse(strings.Join(tests, "\n"))
	t.Log("all tests passed")
}
