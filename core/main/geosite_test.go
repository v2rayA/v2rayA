package main

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"github.com/xtls/xray-core/common/geodata"
	"google.golang.org/protobuf/proto"
)

// writeGeositeDat writes entries in the varint-length-delimited GeoSite
// stream format the real .dat files use.
func writeGeositeDat(t *testing.T, path string, sites ...*geodata.GeoSite) {
	t.Helper()
	var out []byte
	for _, site := range sites {
		body, err := proto.Marshal(site)
		if err != nil {
			t.Fatal(err)
		}
		out = append(out, 0x0a)
		var lenBuf [binary.MaxVarintLen64]byte
		n := binary.PutUvarint(lenBuf[:], uint64(len(body)))
		out = append(out, lenBuf[:n]...)
		out = append(out, body...)
	}
	if err := os.WriteFile(path, out, 0o644); err != nil {
		t.Fatal(err)
	}
}

// A "file.dat:tag" reference (from an ext:"file.dat:tag" DNS rule) is read
// from that file in the asset directory; before this the resolver only ever
// opened geosite.dat, so rules naming a custom geosite file expanded to
// nothing.
func TestGeositeResolverReadsAnExternalFile(t *testing.T) {
	assetDir := t.TempDir()
	t.Setenv("XRAY_LOCATION_ASSET", assetDir)
	writeGeositeDat(t, filepath.Join(assetDir, "custom.dat"), &geodata.GeoSite{
		Code: "MYTAG",
		Domain: []*geodata.Domain{
			{Type: geodata.Domain_Domain, Value: "blocked.example"},
			{Type: geodata.Domain_Full, Value: "exact.blocked.example"},
		},
	})

	domains, suffixes := geositeResolver("custom.dat:MYTAG")
	if len(domains) != 1 || domains[0] != "exact.blocked.example" {
		t.Fatalf("domains = %v", domains)
	}
	if len(suffixes) != 1 || suffixes[0] != "blocked.example" {
		t.Fatalf("suffixes = %v", suffixes)
	}
}

// A bare tag keeps coming from geosite.dat.
func TestGeositeResolverReadsGeositeDatForBareTags(t *testing.T) {
	assetDir := t.TempDir()
	t.Setenv("XRAY_LOCATION_ASSET", assetDir)
	writeGeositeDat(t, filepath.Join(assetDir, "geosite.dat"), &geodata.GeoSite{
		Code:   "CN",
		Domain: []*geodata.Domain{{Type: geodata.Domain_Domain, Value: "cn.example"}},
	})

	_, suffixes := geositeResolver("cn")
	if len(suffixes) != 1 || suffixes[0] != "cn.example" {
		t.Fatalf("suffixes = %v", suffixes)
	}
}

// The file part is a name in the asset directory, never a path.
func TestGeositeResolverRejectsPaths(t *testing.T) {
	assetDir := t.TempDir()
	t.Setenv("XRAY_LOCATION_ASSET", assetDir)
	writeGeositeDat(t, filepath.Join(assetDir, "custom.dat"), &geodata.GeoSite{
		Code:   "MYTAG",
		Domain: []*geodata.Domain{{Type: geodata.Domain_Domain, Value: "blocked.example"}},
	})

	domains, suffixes := geositeResolver("../custom.dat:MYTAG")
	if domains != nil || suffixes != nil {
		t.Fatalf("a path escaped the asset directory: %v %v", domains, suffixes)
	}
}
