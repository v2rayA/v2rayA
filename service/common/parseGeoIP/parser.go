package parseGeoIP

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/v2rayA/v2rayA/kernel/v2ray/asset"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"
)

type parserResult struct {
	ipv4 []string
	ipv6 []string
}

type parserCacheKey struct {
	path        string
	size        int64
	modTime     int64
	countryCode string
}

var parserCache sync.Map

func Parser(filename string, countryCode string) ([]string, []string, error) {
	realpath, err := asset.GetV2rayLocationAsset(filename)
	if err != nil {
		return nil, nil, err
	}
	info, err := os.Stat(realpath)
	if err != nil {
		return nil, nil, err
	}

	cacheKey := parserCacheKey{
		path:        realpath,
		size:        info.Size(),
		modTime:     info.ModTime().UnixNano(),
		countryCode: countryCode,
	}
	if cached, ok := parserCache.Load(cacheKey); ok {
		result := cached.(parserResult)
		return result.ipv4, result.ipv6, nil
	}

	data, err := os.ReadFile(realpath)
	if err != nil {
		return nil, nil, err
	}

	ipv4List, ipv6List, err := parseGeoIP(data, countryCode)
	if err != nil {
		return ipv4List, ipv6List, err
	}

	parserCache.Store(cacheKey, parserResult{ipv4: ipv4List, ipv6: ipv6List})
	return ipv4List, ipv6List, nil
}

func parseGeoIP(data []byte, countryCode string) ([]string, []string, error) {
	var ipv4List []string
	var ipv6List []string

	// GeoIPList contains repeated GeoIP messages in field 1. Decoding the
	// complete list expands a roughly 22 MiB asset to hundreds of MiB in
	// memory. Decode one entry at a time and stop at the requested country.
	for len(data) > 0 {
		number, wireType, n := protowire.ConsumeTag(data)
		if n < 0 {
			return ipv4List, ipv6List, protowire.ParseError(n)
		}
		data = data[n:]

		if number == 1 && wireType == protowire.BytesType {
			entryData, consumed := protowire.ConsumeBytes(data)
			if consumed < 0 {
				return ipv4List, ipv6List, protowire.ParseError(consumed)
			}
			data = data[consumed:]

			var geo GeoIP
			if err := proto.Unmarshal(entryData, &geo); err != nil {
				return ipv4List, ipv6List, err
			}
			if geo.CountryCode != countryCode {
				continue
			}

			for _, c := range geo.Cidr {
				ip := net.IP(c.Ip)
				if strings.Contains(ip.String(), ":") {
					ipv6List = append(ipv6List, fmt.Sprintf("%s/%d", ip.String(), c.Prefix))
				} else {
					if strings.Contains(ip.String(), "198.18.0.0") {
						// 跳过fakeip
						continue
					}
					ipv4List = append(ipv4List, fmt.Sprintf("%s/%d", ip.String(), c.Prefix))
				}
			}
			return ipv4List, ipv6List, nil
		}

		consumed := protowire.ConsumeFieldValue(number, wireType, data)
		if consumed < 0 {
			return ipv4List, ipv6List, protowire.ParseError(consumed)
		}
		data = data[consumed:]
	}

	return ipv4List, ipv6List, nil
}
