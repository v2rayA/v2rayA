package serverObj

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestAnyTLSConfigurationMapsInsecure(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{value: "1", want: true},
		{value: "true", want: true},
		{value: "yes", want: true},
		{value: "on", want: true},
		{value: "0", want: false},
		{value: "false", want: false},
		{value: "", want: false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("value_%s", tt.value), func(t *testing.T) {
			server, err := ParseAnyTLSURL("anytls://password@example.com:443?insecure=" + tt.value + "#test")
			if err != nil {
				t.Fatalf("ParseAnyTLSURL() error = %v", err)
			}

			configuration, err := server.Configuration(PriorInfo{Tag: "test"})
			if err != nil {
				t.Fatalf("Configuration() error = %v", err)
			}

			var settings map[string]any
			if err := json.Unmarshal(configuration.CoreOutbound.Settings.Inlined, &settings); err != nil {
				t.Fatalf("unmarshal settings: %v", err)
			}

			got, present := settings["allow_insecure"].(bool)
			if tt.want {
				if !present || !got {
					t.Fatalf("allow_insecure = %v, present = %v; want true", got, present)
				}
			} else if present {
				t.Fatalf("allow_insecure unexpectedly present with value %v", got)
			}
		})
	}
}
