package routingA

import (
	"testing"
)

func TestNewFunction(t *testing.T) {
	tests := []string{
		`outbound: socksout = socks(address: 127.0.0.1, port: 10800, user: "my-username", pass: "my-password")`,
		`outbound:   test=   socks ( 127.0.0.1,   port: 10800, user: "my-username",pass:"my-password" )`,
		`outbound:   test2=   socks ( test,   port: 10800, user: "my-username",pass:"my-password" )`,
		`outbound:test3=socks()`,
	}
	for _, test := range tests {
		S, _ := generateSyntaxTree(test)
		f := newFunction(S.children[0].children[0].children[2].children[2])
		t.Log(f.Name, f.Params, f.NamedParams)
	}
}

func TestNewOutbound(t *testing.T) {
	tests := []string{
		`outbound: socksout = socks(address: 127.0.0.1, port: 10800, user: "my-username", pass: "my-password")`,
		`outbound:   test=   socks ( 127.0.0.1,   port: 10800, user: "my-username",pass:"my-password" )`,
		`outbound:   test2=   socks ( test,   port: 10800, user: "my-username",pass:"my-password" )`,
	}
	for _, test := range tests {
		S, _ := generateSyntaxTree(test)
		o := newOutbound(S.children[0].children[0].children[2])
		t.Log(o.Name, o.Value)
	}
}

func TestNewDefine(t *testing.T) {
	tests := []string{
		`outbound: socksout = socks(address: 127.0.0.1, port: 10800, user: "my-username", pass: "my-password")`,
		`default  : httpout`,
	}
	for _, test := range tests {
		S, _ := generateSyntaxTree(test)
		o := newDefine(S.children[0].children[0])
		t.Log(o.Name, o.Value)
	}
}

func TestNewRouting(t *testing.T) {
	tests := []string{
		`domain(domain: v2raya.mzz.pub) -> socksout`,
		`domain(full: dns.google) -> proxy`,
		`domain(contains: .google.) -> proxy`,
		`ip(127.0.0.1) -> direct`,
		`ip(192.168.0.0/16) -> direct`,
		`extern(ip, geoip, private) -> direct`,
		`extern(domain, geosite, category-ads) -> block`,
		`ip(8.8.8.8) && network(tcp, udp) && port(1-1023, 8443) -> proxy`,
		`ip(1.1.1.1, 1.2.3.4, 9.9.9.9) && protocol(http) -> direct`,
	}
	for _, test := range tests {
		S, _ := generateSyntaxTree(test)
		r := newRouting(S.children[0].children[0])
		t.Log(r.And, r.Out)
	}
}
