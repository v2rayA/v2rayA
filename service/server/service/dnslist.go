package service

import (
	"fmt"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/core/v2ray"
	"github.com/v2rayA/v2rayA/core/v2ray/service"
	"net"
	"net/url"
	"strconv"
	"strings"
)

var UnsupportedProtocol = fmt.Errorf("the version of installed core does not support the protocol")

func RefineDnsList(dnsList string) (string, error) {
	list := strings.Split(strings.TrimSpace(dnsList), "\n")
	if len(list) == 0 || strings.TrimSpace(dnsList) == "" {
		return "", nil
	}
	list = common.Deduplicate(list)
nextLine:
	for i, line := range list {
		dns := v2ray.ParseAdvancedDnsLine(line)
		if dns == nil {
			return "", fmt.Errorf("invalid format: %v: no outbound found", line)
		}
		if dns.Val == "localhost" {
			return "", fmt.Errorf("instead of localhost, use 127.0.0.1 or ::1 because it is a keyword of v2ray-core")
		}
		if dns.Val == "" {
			return "", fmt.Errorf("illegal server: %v", line)
		}
		if dns.Out == "block" {
			return "", fmt.Errorf("cannot use block as outobund")
		}
		if net.ParseIP(dns.Val) != nil {
			continue nextLine
		} else {
			_, port, err := net.SplitHostPort(dns.Val)
			if err == nil {
				if _, err := strconv.Atoi(port); err == nil {
					continue nextLine
				}
			}
			if u, err := url.Parse(dns.Val); err == nil {
				switch u.Scheme {
				case "https":
					if service.CheckDohSupported() != nil {
						return "", fmt.Errorf("%w: %v", UnsupportedProtocol, u.Scheme)
					}
				case "tcp":
					if strings.HasPrefix(u.Scheme, "tcp") && service.CheckTcpDnsSupported() != nil {
						return "", fmt.Errorf("%w: %v", UnsupportedProtocol, u.Scheme)
					}
				case "quic":
					// FIXME: after quic:// supported
					if service.CheckQuicLocalDnsSupported() != nil {
						return "", fmt.Errorf("%w: %v", UnsupportedProtocol, u.Scheme)
					}
				case "":
					goto invalid
				default:
					return "", fmt.Errorf("unsupported dns protocol: %v", u.Scheme)
				}
				port := u.Port()
				if port == "" {
					switch u.Scheme {
					case "https":
						port = "443"
					case "tcp":
						port = "53"
					case "quic":
						port = "784"
					}
					u.Host = net.JoinHostPort(u.Hostname(), port)
					list[i] = u.String() + "->" + dns.Out
				}
				continue nextLine
			}
		invalid:
			return "", fmt.Errorf("invalid server: %v", line)
		}
	}
	return strings.Join(list, "\n"), nil
}
