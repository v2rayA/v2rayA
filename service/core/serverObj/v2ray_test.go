package serverObj

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	jsoniter "github.com/json-iterator/go"
)

func loadXHTTPFixture(t *testing.T, name string) map[string]interface{} {
	t.Helper()
	path := filepath.Join("..", "..", "..", "tmp", "test-configs", name)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", path, err)
	}
	var m map[string]interface{}
	if err = jsoniter.Unmarshal(b, &m); err != nil {
		t.Fatalf("failed to unmarshal fixture %s: %v", path, err)
	}
	return m
}

func marshalJSON(t *testing.T, v interface{}) string {
	t.Helper()
	b, err := jsoniter.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal json: %v", err)
	}
	return string(b)
}

func assertJSONEqual(t *testing.T, got string, want interface{}) {
	t.Helper()
	var gotValue interface{}
	if err := jsoniter.Unmarshal([]byte(got), &gotValue); err != nil {
		t.Fatalf("failed to unmarshal got json: %v", err)
	}
	wantJSON := marshalJSON(t, want)
	var wantValue interface{}
	if err := jsoniter.Unmarshal([]byte(wantJSON), &wantValue); err != nil {
		t.Fatalf("failed to unmarshal want json: %v", err)
	}
	if !reflect.DeepEqual(gotValue, wantValue) {
		t.Fatalf("unexpected json\nwant: %s\ngot:  %s", wantJSON, got)
	}
}

func passthroughJSONFromXHTTPFixture(t *testing.T, name string) string {
	t.Helper()
	m := loadXHTTPFixture(t, name)
	delete(m, "path")
	delete(m, "host")
	delete(m, "mode")
	return marshalJSON(t, m)
}

func configuredXHTTPSettingsJSON(t *testing.T, v *V2Ray) string {
	t.Helper()
	cfg, err := v.Configuration(PriorInfo{Tag: "proxy"})
	if err != nil {
		t.Fatalf("failed to generate configuration: %v", err)
	}
	return marshalJSON(t, cfg.CoreOutbound.StreamSettings.XHTTPSettings)
}

func TestV2RayXHTTPConfigurationWithoutRawExtras(t *testing.T) {
	v := &V2Ray{
		Add:       "proxy.example.com",
		Port:      "443",
		ID:        "11111111-1111-1111-1111-111111111111",
		Net:       "xhttp",
		Path:      "/",
		Host:      "proxy.example.com",
		XHTTPMode: "auto",
		TLS:       "none",
		Protocol:  "vless",
	}

	got := configuredXHTTPSettingsJSON(t, v)
	want := map[string]interface{}{
		"path": "/",
		"host": "proxy.example.com",
		"mode": "auto",
	}
	assertJSONEqual(t, got, want)
}

func TestV2RayXHTTPConfigurationWithDownloadSettings(t *testing.T) {
	v := &V2Ray{
		Add:          "proxy.example.com",
		Port:         "443",
		ID:           "11111111-1111-1111-1111-111111111111",
		Net:          "xhttp",
		Path:         "/",
		Host:         "proxy.example.com",
		XHTTPMode:    "packet-up",
		XHTTPRawJson: passthroughJSONFromXHTTPFixture(t, "xhttp-settings-with-downloads.json"),
		TLS:          "reality",
		SNI:          "proxy.example.com",
		PublicKey:    "FAKE_PUBLIC_KEY",
		ShortId:      "0123456789ab",
		SpiderX:      "/",
		Fingerprint:  "chrome",
		Protocol:     "vless",
	}

	got := configuredXHTTPSettingsJSON(t, v)
	assertJSONEqual(t, got, loadXHTTPFixture(t, "xhttp-settings-with-downloads.json"))
}

func TestV2RayXHTTPRoundtripPreservesRawExtras(t *testing.T) {
	rawJSON := passthroughJSONFromXHTTPFixture(t, "xhttp-settings-with-downloads.json")
	original := &V2Ray{
		Ps:           "xhttp-node",
		Add:          "proxy.example.com",
		Port:         "443",
		ID:           "11111111-1111-1111-1111-111111111111",
		Net:          "xhttp",
		Path:         "/",
		Host:         "proxy.example.com",
		XHTTPMode:    "packet-up",
		XHTTPRawJson: rawJSON,
		TLS:          "reality",
		SNI:          "proxy.example.com",
		PublicKey:    "FAKE_PUBLIC_KEY",
		ShortId:      "0123456789ab",
		SpiderX:      "/",
		Fingerprint:  "chrome",
		Protocol:     "vless",
	}

	link := original.ExportToURL()
	parsedObj, err := ParseVlessURL(link)
	if err != nil {
		t.Fatalf("failed to parse exported url: %v", err)
	}
	if parsedObj.XHTTPRawJson != rawJSON {
		t.Fatalf("xhttpRawJson changed after roundtrip\nwant: %s\ngot:  %s", rawJSON, parsedObj.XHTTPRawJson)
	}

	got := configuredXHTTPSettingsJSON(t, parsedObj)
	assertJSONEqual(t, got, loadXHTTPFixture(t, "xhttp-settings-with-downloads.json"))
}
