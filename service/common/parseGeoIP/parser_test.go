package parseGeoIP

import (
	"net"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
)

func TestParserCache(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XRAY_LOCATION_ASSET", dir)
	parserCache = make(map[parserCacheKey]parserResult)
	t.Cleanup(func() { parserCache = make(map[parserCacheKey]parserResult) })
	write := func(name string, generation int) {
		t.Helper()
		data, err := proto.Marshal(&GeoIPList{Entry: []*GeoIP{
			{CountryCode: "A", Cidr: []*CIDR{{Ip: []byte{10, byte(generation), 0, 0}, Prefix: 16}}},
			{CountryCode: "B", Cidr: []*CIDR{{Ip: []byte{192, 0, 2, 0}, Prefix: 24}}},
		}})
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
		stamp := time.Unix(int64(generation), 0)
		if err := os.Chtimes(path, stamp, stamp); err != nil {
			t.Fatal(err)
		}
	}
	write("other.dat", 9)
	if _, _, err := Parser("other.dat", "A"); err != nil {
		t.Fatal(err)
	}
	for generation := 1; generation <= 3; generation++ {
		write("test.dat", generation)
		var wg sync.WaitGroup
		results := make([][]string, 16)
		for i := range results {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				var err error
				results[i], _, err = Parser("test.dat", "A")
				if err != nil {
					t.Error(err)
				}
			}(i)
		}
		wg.Wait()
		want := net.IP{10, byte(generation), 0, 0}.String() + "/16"
		for _, result := range results {
			if len(result) != 1 || result[0] != want {
				t.Fatalf("got %v, want %s", result, want)
			}
			// Re-parsing allocates another slice; concurrent callers must share the cached result.
			if &result[0] != &results[0][0] {
				t.Fatal("concurrent miss parsed more than once")
			}
		}
		if got, _, err := Parser("test.dat", "B"); err != nil || !reflect.DeepEqual(got, []string{"192.0.2.0/24"}) {
			t.Fatalf("second country: %v, %v", got, err)
		}
		if len(parserCache) != 3 {
			t.Fatalf("cache retains %d entries, want 3", len(parserCache))
		}
	}
	if got, _, err := Parser("other.dat", "A"); err != nil || !reflect.DeepEqual(got, []string{"10.9.0.0/16"}) {
		t.Fatalf("other file: %v, %v", got, err)
	}

	if _, _, err := Parser("missing.dat", "A"); err == nil {
		t.Fatal("missing file accepted")
	}
	if err := os.WriteFile(filepath.Join(dir, "missing.dat"), []byte{0x0a, 0x02, 0xff}, 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Parser("missing.dat", "A"); err == nil {
		t.Fatal("malformed file accepted")
	}
	if len(parserCache) != 3 {
		t.Fatal("failed parse cached")
	}
	write("missing.dat", 4)
	if got, _, err := Parser("missing.dat", "A"); err != nil || !reflect.DeepEqual(got, []string{"10.4.0.0/16"}) {
		t.Fatalf("retry: %v, %v", got, err)
	}
}

func TestParseGeoIPStopsAtRequestedEntry(t *testing.T) {
	data, err := proto.Marshal(&GeoIPList{Entry: []*GeoIP{
		{
			CountryCode: "PRIVATE",
			Cidr: []*CIDR{
				{Ip: net.ParseIP("10.0.0.0").To4(), Prefix: 8},
				{Ip: net.ParseIP("198.18.0.0").To4(), Prefix: 15},
				{Ip: net.ParseIP("2001:db8::"), Prefix: 32},
			},
		},
		{
			CountryCode: "US",
			Cidr: []*CIDR{
				{Ip: net.ParseIP("192.0.2.0").To4(), Prefix: 24},
			},
		},
	}})
	if err != nil {
		t.Fatal(err)
	}

	// A malformed record after the requested entry must not be decoded.
	data = append(data, 0x0a, 0x02, 0xff)

	ipv4, ipv6, err := parseGeoIP(data, "PRIVATE")
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"10.0.0.0/8"}; !reflect.DeepEqual(ipv4, want) {
		t.Fatalf("IPv4 entries = %v, want %v", ipv4, want)
	}
	if want := []string{"2001:db8::/32"}; !reflect.DeepEqual(ipv6, want) {
		t.Fatalf("IPv6 entries = %v, want %v", ipv6, want)
	}
}

func TestParseGeoIPRejectsMalformedData(t *testing.T) {
	if _, _, err := parseGeoIP([]byte{0x0a, 0x02, 0xff}, "PRIVATE"); err == nil {
		t.Fatal("parseGeoIP() accepted malformed data")
	}
}
