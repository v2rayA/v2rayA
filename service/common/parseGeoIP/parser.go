package parseGeoIP

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"google.golang.org/protobuf/proto"
)

func Parser(filename string, countryCode string) ([]string, []string, error) {
	var ipv4List []string
	var ipv6List []string
	var geoIpProto GeoIPList

	realpath, err := asset.GetV2rayLocationAsset(filename)
	if err != nil {
		return ipv4List, ipv6List, err
	}
	data, err := os.ReadFile(realpath)
	if err != nil {
		return ipv4List, ipv6List, err
	}

	if err := proto.Unmarshal(data, &geoIpProto); err != nil {
		return ipv4List, ipv6List, err
	}

	for _, geo := range geoIpProto.Entry {
		if geo.CountryCode == countryCode {
			for _, c := range geo.Cidr {
				ip := net.IP(c.Ip)
				if strings.Contains(ip.String(), ":") {
					ipv6List = append(ipv6List, fmt.Sprintf("%s/%d", ip.String(), c.Prefix))
				} else {
					ipv4List = append(ipv4List, fmt.Sprintf("%s/%d", ip.String(), c.Prefix))
				}
			}
			break
		}
	}
	return ipv4List, ipv6List, nil
}
