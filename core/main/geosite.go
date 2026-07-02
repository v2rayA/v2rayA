package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/xtls/xray-core/common/geodata"
	"google.golang.org/protobuf/proto"
)

// geositeResolver implements dns.GeositeResolver using xray-core's geodata.
// It loads geosite.dat from the standard asset directory and expands tags
// (e.g. "cn") into domain/suffix patterns.
func geositeResolver(tag string) (domains, suffixes []string) {
	// Determine geosite.dat location from environment.
	assetDir := os.Getenv("XRAY_LOCATION_ASSET")
	if assetDir == "" {
		assetDir = os.Getenv("V2RAY_LOCATION_ASSET")
	}
	if assetDir == "" {
		log.Printf("[dns] geosite: XRAY_LOCATION_ASSET not set, skipping geosite expansion")
		return nil, nil
	}

	geositePath := filepath.Join(assetDir, "geosite.dat")
	data, err := os.ReadFile(geositePath)
	if err != nil {
		log.Printf("[dns] geosite: cannot read %s: %v", geositePath, err)
		return nil, nil
	}

	var list geodata.GeoSiteList
	if err := proto.Unmarshal(data, &list); err != nil {
		log.Printf("[dns] geosite: cannot parse %s: %v", geositePath, err)
		return nil, nil
	}

	for _, entry := range list.Entry {
		if entry.Code != tag {
			continue
		}
		for _, domain := range entry.Domain {
			switch domain.Type {
			case geodata.Domain_Domain:
				// Domain match: exact domain (but also matches subdomains in xray routing).
				// In our DNS module, we add as suffix match for broader matching.
				suffixes = append(suffixes, domain.Value)
			case geodata.Domain_Full:
				// Full match: exact domain only.
				domains = append(domains, domain.Value)
			case geodata.Domain_Regex:
				// Regex: include as-is; the router will create Regex matchers.
				// For DNS rules, we treat these as domains (the matcher will fail gracefully).
				domains = append(domains, domain.Value)
			case geodata.Domain_Substr:
				// Substring: cannot be directly used as domain matcher; skip.
				// Note: keyword: tag would be better for substring matching.
			}
		}
		if len(domains) > 0 || len(suffixes) > 0 {
			log.Printf("[dns] geosite: expanded %s → %d domains, %d suffixes", tag, len(domains), len(suffixes))
		}
		return domains, suffixes
	}

	log.Printf("[dns] geosite: tag %q not found in geosite.dat", tag)
	return nil, nil
}
