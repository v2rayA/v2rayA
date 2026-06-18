package serverObj

import (
	"encoding/json"
	"testing"
)

func TestJuicityConfig(t *testing.T) {
	// 1. Create a Juicity share link
	link := "juicity://00000000-0000-0000-0000-000000000000:my_password@example.com:30020?congestion_control=bbr&sni=www.example.com&pinned_certchain_sha256=abcdef"

	// 2. Parse the link
	s, err := ParseJuicityURL(link)
	if err != nil {
		t.Fatalf("Failed to parse link: %v", err)
	}

	// 3. Verify parsed struct fields
	if s.Address != "example.com" {
		t.Errorf("Expected Address to be example.com, got %s", s.Address)
	}
	if s.Port != 30020 {
		t.Errorf("Expected Port to be 30020, got %d", s.Port)
	}
	if s.UUID != "00000000-0000-0000-0000-000000000000" {
		t.Errorf("Expected UUID, got %s", s.UUID)
	}
	if s.Password != "my_password" {
		t.Errorf("Expected password to be my_password, got %s", s.Password)
	}
	if s.Sni != "www.example.com" {
		t.Errorf("Expected sni to be www.example.com, got %s", s.Sni)
	}
	if s.CC != "bbr" {
		t.Errorf("Expected CC to be bbr, got %s", s.CC)
	}
	if s.PinnedCertchainSha256 != "abcdef" {
		t.Errorf("Expected PinnedCertchainSha256, got %s", s.PinnedCertchainSha256)
	}

	// 4. Generate hybrid-core Configuration
	c, err := s.Configuration(PriorInfo{Tag: "test_tag"})
	if err != nil {
		t.Fatalf("Failed to generate configuration: %v", err)
	}

	// 5. Verify generated outbound settings
	if c.CoreOutbound.Protocol != "juicity" {
		t.Errorf("Expected protocol to be juicity, got %s", c.CoreOutbound.Protocol)
	}

	// Check if allow_insecure is present in settings JSON
	var rawSettings map[string]interface{}
	if err := json.Unmarshal(c.CoreOutbound.Settings.Inlined, &rawSettings); err != nil {
		t.Fatalf("Failed to unmarshal settings: %v", err)
	}

	if _, ok := rawSettings["allow_insecure"]; ok {
		t.Errorf("allow_insecure field should be completely removed from the configuration")
	}
	if _, ok := rawSettings["allowInsecure"]; ok {
		t.Errorf("allowInsecure field should be completely removed from the configuration")
	}

	// 6. Verify unmarshaling into a client config struct matching core JSON tags
	type TargetClientConfig struct {
		Address           string `json:"address"`
		UUID              string `json:"uuid"`
		Password          string `json:"password"`
		SNI               string `json:"sni"`
		AllowInsecure     bool   `json:"allow_insecure"`
		CongestionControl string `json:"congestion_control"`
		PinnedSHA256      string `json:"pinned_certchain_sha256"`
	}

	var target TargetClientConfig
	if err := json.Unmarshal(c.CoreOutbound.Settings.Inlined, &target); err != nil {
		t.Fatalf("Failed to unmarshal into TargetClientConfig: %v", err)
	}

	if target.Address != "example.com:30020" {
		t.Errorf("Expected address example.com:30020, got %s", target.Address)
	}
	if target.UUID != "00000000-0000-0000-0000-000000000000" {
		t.Errorf("Expected UUID, got %s", target.UUID)
	}
	if target.Password != "my_password" {
		t.Errorf("Expected password, got %s", target.Password)
	}
	if target.SNI != "www.example.com" {
		t.Errorf("Expected SNI, got %s", target.SNI)
	}
	if target.AllowInsecure != false {
		t.Errorf("Expected AllowInsecure to default to false, got %t", target.AllowInsecure)
	}
	if target.CongestionControl != "bbr" {
		t.Errorf("Expected CC, got %s", target.CongestionControl)
	}
	if target.PinnedSHA256 != "abcdef" {
		t.Errorf("Expected PinnedSHA256, got %s", target.PinnedSHA256)
	}
}
