package resolv

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

type IPDNSEndpoint struct {
	URL    string
	IP     net.IP
	Port   string
	Scheme string
}

// ParseIPDNS is shared by discovery, validation and node queries. Encrypted
// endpoints keep their IP as the certificate identity; no bootstrap is needed.
func ParseIPDNS(address string) (*IPDNSEndpoint, error) {
	address = strings.TrimSpace(address)
	if !strings.Contains(address, "://") {
		if ip := net.ParseIP(address); ip != nil {
			address = net.JoinHostPort(ip.String(), "53")
		} else if strings.HasPrefix(address, "[") && strings.HasSuffix(address, "]") {
			address = net.JoinHostPort(strings.Trim(address, "[]"), "53")
		}
		address = "udp://" + address
	}
	u, err := url.Parse(address)
	if err != nil {
		return nil, fmt.Errorf("invalid IP DNS address: %w", err)
	}
	u.Scheme = strings.ToLower(u.Scheme)
	defaults := map[string]string{"udp": "53", "tcp": "53", "tls": "853", "https": "443"}
	port, supported := defaults[u.Scheme]
	if !supported || u.User != nil || u.Fragment != "" || u.Opaque != "" {
		return nil, fmt.Errorf("unsupported IP DNS URL %q", address)
	}
	ip := net.ParseIP(u.Hostname())
	if ip == nil || ip.IsUnspecified() || ip.IsMulticast() || ip.Equal(net.IPv4bcast) {
		return nil, fmt.Errorf("DNS host must be a unicast IP: %q", address)
	}
	if strings.HasSuffix(u.Host, ":") {
		return nil, fmt.Errorf("empty DNS port")
	}
	if p := u.Port(); p != "" {
		port = p
	}
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return nil, fmt.Errorf("invalid DNS port %q", port)
	}
	if u.Scheme != "https" && (u.Path != "" || u.RawQuery != "" || u.ForceQuery) {
		return nil, fmt.Errorf("only HTTPS DNS supports paths and queries")
	}
	port = strconv.Itoa(n)
	u.Host = net.JoinHostPort(ip.String(), port)
	if u.Scheme == "https" && u.Path == "" {
		u.Path = "/"
	}
	return &IPDNSEndpoint{URL: u.String(), IP: ip, Port: port, Scheme: u.Scheme}, nil
}

func (e *IPDNSEndpoint) Address() string { return net.JoinHostPort(e.IP.String(), e.Port) }
func (e *IPDNSEndpoint) Protocol() string {
	if e.Scheme == "tls" {
		return "tcp-tls"
	}
	return e.Scheme
}
func (e *IPDNSEndpoint) UpstreamAddress() string {
	if e.Scheme == "https" {
		return e.URL
	}
	return e.Address()
}

// FallbackDNSURLs returns a copy so discovery cannot change the resolver's list.
func FallbackDNSURLs() []string {
	out := make([]string, 0, len(dnsServers))
	for _, server := range dnsServers {
		out = append(out, server.network+"://"+server.addr)
	}
	return out
}

func IsLocalIP(ip net.IP) bool {
	if ip.IsLoopback() {
		return true
	}
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		local, _, _ := net.ParseCIDR(addr.String())
		if local.Equal(ip) {
			return true
		}
	}
	return false
}
