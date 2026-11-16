package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/xtls/xray-core/common/geodata"
	"google.golang.org/protobuf/proto"
)

// geositeResolver implements dns.GeositeResolver using xray-core's geodata.
// It expands a reference from the standard asset directory: a bare tag
// (e.g. "cn") is read from geosite.dat, and "file.dat:tag" (from an
// ext:"file.dat:tag" rule) is read from that file, so custom geosite
// databases work the same way.
//
// The .dat files use a custom format: varint-length-delimited GeoSite
// protobuf messages concatenated together (NOT a single GeoSiteList message).
// This function correctly parses that format.
func geositeResolver(spec string) (domains, suffixes []string) {
	// Determine the asset directory from the environment.
	assetDir := os.Getenv("XRAY_LOCATION_ASSET")
	if assetDir == "" {
		assetDir = os.Getenv("V2RAY_LOCATION_ASSET")
	}
	if assetDir == "" {
		log.Printf("[dns] geosite: XRAY_LOCATION_ASSET not set, skipping geosite expansion")
		return nil, nil
	}

	file, tag := "geosite.dat", spec
	if f, t, ok := strings.Cut(spec, ":"); ok {
		file, tag = f, t
	}
	if filepath.Base(file) != file || file == "" || tag == "" {
		log.Printf("[dns] geosite: invalid reference %q, skipping", spec)
		return nil, nil
	}

	geositePath := filepath.Join(assetDir, file)
	geosite, err := loadGeositeTag(geositePath, tag)
	if err != nil {
		log.Printf("[dns] geosite: cannot load tag %q from %s: %v", tag, geositePath, err)
		return nil, nil
	}

	for _, domain := range geosite {
		switch domain.Type {
		case geodata.Domain_Domain:
			// Domain match: treat as suffix match (covers subdomains).
			suffixes = append(suffixes, domain.Value)
		case geodata.Domain_Full:
			// Full match: exact domain only.
			domains = append(domains, domain.Value)
		case geodata.Domain_Regex:
			// Regex: include as-is.
			domains = append(domains, domain.Value)
		case geodata.Domain_Substr:
			// Substring: cannot be directly used as domain matcher; skip.
		}
	}

	if len(domains) > 0 || len(suffixes) > 0 {
		log.Printf("[dns] geosite: expanded %s → %d domains, %d suffixes", spec, len(domains), len(suffixes))
	}
	return domains, suffixes
}

// geositeCache caches parsed GeoSite entries so we don't re-read the file
// on every rule expansion. The cache is keyed by (filePath, tag).
var geositeCache struct {
	mu   sync.RWMutex
	data map[string][]*geodata.Domain // key: "filePath:tag"
}

func init() {
	geositeCache.data = make(map[string][]*geodata.Domain)
}

// loadGeositeTag reads a geosite .dat file and extracts the domains for a
// specific tag.
//
// The file uses a custom format where each entry is a varint-length-delimited
// GeoSite protobuf message. This is NOT a single GeoSiteList message —
// attempting proto.Unmarshal of the entire file as GeoSiteList would only
// parse the first entry.
//
// This function correctly reads the stream by:
//  1. Reading the outer varint: frame tag (0x0a = field 1, wire type 2) + length
//  2. Extracting the message body
//  3. Unmarshalling it as GeoSite
//  4. Checking if the code matches
func loadGeositeTag(filePath, tag string) ([]*geodata.Domain, error) {
	// Check cache first (try original tag and uppercase).
	cacheKey := filePath + ":" + tag
	geositeCache.mu.RLock()
	if cached, ok := geositeCache.data[cacheKey]; ok {
		geositeCache.mu.RUnlock()
		return cached, nil
	}
	geositeCache.mu.RUnlock()

	tagUpper := strings.ToUpper(tag)
	if tagUpper != tag {
		upperKey := filePath + ":" + tagUpper
		geositeCache.mu.RLock()
		if cached, ok := geositeCache.data[upperKey]; ok {
			geositeCache.mu.RUnlock()
			return cached, nil
		}
		geositeCache.mu.RUnlock()
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", filePath, err)
	}
	defer f.Close()

	br := bufio.NewReaderSize(f, 64*1024)

	for {
		// Read the outer frame tag byte.
		// Each entry starts with field 1 wire type 2 (length-delimited) = 0x0a.
		tagByte, err := br.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read frame tag: %w", err)
		}
		if tagByte != 0x0a {
			// Not the expected frame format — skip.
			continue
		}

		// Read the varint length of the message body.
		bodyLen, err := binary.ReadUvarint(br)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("read frame length: %w", err)
		}

		// Read the message body.
		body := make([]byte, bodyLen)
		if _, err := io.ReadFull(br, body); err != nil {
			return nil, fmt.Errorf("read frame body: %w", err)
		}

		// Unmarshal as GeoSite and check code.
		var site geodata.GeoSite
		if err := proto.Unmarshal(body, &site); err != nil {
			// Skip entries we can't parse.
			continue
		}

		if site.Code != tag && site.Code != tagUpper {
			continue
		}

		// Found matching entry — cache and return.
		domains := site.Domain
		geositeCache.mu.Lock()
		geositeCache.data[cacheKey] = domains
		geositeCache.mu.Unlock()
		return domains, nil
	}

	// Tag not found — cache empty result to avoid re-scanning.
	geositeCache.mu.Lock()
	geositeCache.data[cacheKey] = nil
	geositeCache.mu.Unlock()

	return nil, fmt.Errorf("tag %q not found in %s", tag, filePath)
}
