package iptables

import "testing"

func TestLegacyDNSRedirectTargetRules(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    string
	}{
		{
			name:    "IPv4",
			command: "iptables",
			want: "iptables -w 2 -t nat -A DNS_REDIRECT -p tcp -j REDIRECT --to-port 52353\n" +
				"iptables -w 2 -t nat -A DNS_REDIRECT -p udp -j REDIRECT --to-port 52353\n",
		},
		{
			name:    "IPv6",
			command: "ip6tables",
			want: "ip6tables -w 2 -t nat -A DNS_REDIRECT -p tcp -j REDIRECT --to-port 52353\n" +
				"ip6tables -w 2 -t nat -A DNS_REDIRECT -p udp -j REDIRECT --to-port 52353\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := legacyDNSRedirectTargetRules(tt.command); got != tt.want {
				t.Fatalf("DNS REDIRECT 规则不符合预期:\n%s", got)
			}
		})
	}
}
