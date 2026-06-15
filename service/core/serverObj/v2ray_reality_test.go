package serverObj

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRealityNoAllowInsecure(t *testing.T) {
	link := "vless://81c4b73a-bd74-4a14-a24a-145427941b95@[2001:19f0:5:f94:5400:3ff:fe35:c126]:9443?alpn=h3&fp=chrome&type=xhttp&sni=www.microsoft.com&sid=77c2358dc476ae9e&mode=auto&path=%2Ffuckgfw&security=reality&encryption=mlkem768x25519plus.native.0rtt._KYghfqzHjhPtQi4XIwcZmZz4SIJs1QMXXxYepxDsnmllvhGwkuF_nDAAgjDvZxCeaE4DMi4maodOyx_tqtUuDSH-7jAjbWlFAZYQ2pX-AG_dJfJKOGd06CKOYBd22SMv_J1mgwLzpsXjNu_2ltuLuiyprJwObe4WNRaqNp2xyoctzoELCKPTyAWpSpix2uxq7N4mTAMp-ZwXQidwZTKG0sY_TClIWMKw2NzO3OPP3KOhMN5hMdGeSGmXEu2cfXPemCkgqIGCHmhffRvSZVSAJ0tVUNVTPds70KujCo-G0wH_7cgtAC8hlsqXkB0jLhJ1QJ8Owd6GXZ440GlUyJPW5A7w8FgSkt3TKZUcUOHL8eKXiSrgUA4PuYov0Zg-CF7qipFUNuFLQRuBJot3HxWomu7Kxd0YHjFAMILakYM63hgfKoRHCfChXXC8nhm7dZQUfgQ8zbC6xfJuQCSPmShyaWm2Dqb1laxAbs2lptG3JWZl2WRL0ZdPoDOM7pH9-OTlWVfvXlcOYUbb-qxDFEQUAgUnEu45UQNc8ZZnMt2Y3W3LYq4THaoeBELQSFr1oQk9NXIQDeYsrKP3IvO4VwH5lFEoINiBzBzh7EUksZ-HzgWcaC1YzMceeQquLmZJZWODpZHnzaNv0tdX5ODSBFjXkmJm1Zeq0hHeqkx0yNdeOHJ6YFyQ7FDurF2IBQPZ7Z8R2a_dbC7BDDP6SkgU4tq_9Oz34K_u6pGWPWy9HsHCRcnMCoglZVNXbKgqcda2GgFFEoJpyZHA-hQjloon_lR64KHKbg9WQZ1YpIb8_s5fBpDMPxRF7c2CGWUKzl6cmlwvHmga2CG-XaczFBXLvco_FhNjgeO0OlfwNNl0HurWGxJaZqzdGkR-LHLvjBTudC0RgTEJZIDUQuSUHmomTlTdJSR0xSfEbGKo1h0flyHIEBBQeZtEWJpJWlRAHjJj1zHz2ZdELCP5gp2klIq4hFVQqUu7YcydEJ15oA92OkTA3tb_UWdYOG7o6UeCGqrxxRdNGQGvuph0yGpDYymOMESIPe1cZq_AB1rkeMWmzZJ3YZaIws7jtp53ct8nEUH4EowWfoIHSmI8Dm8uBFK09EBL6dtKpQrfLs8bXN4JeYa7Aup8Mim__E2uSZ-uYO-2kml8HmqA1u0_iFqQMlaeiBByKXH51rCBVp3Q2qn20FxpMynf9R3dVYB7sNjAeXEEXNuBZFMGKUEyCRGq_g51oW15WaQAYR82SKwDOVBzQKS-XFe0fCVSIQ4pvAufGQDXgtbDINK_IExxBV_eoolYDMSS0wrpNQQURCHRvKLtPWw1fpOBQR2EvUS_iNSS_SvNgqzEamdx2V-0tmhM2UULlRrcZgneRHAH7iXi8mvJ8qdUkCuQomVK4uplMOsb-RAaht2A5hhHuc_qFk7XZQv0RxLljS-gVQ8GkmIR0mkq7QL9HkIKZKMwqR1sxuXeTIAkFqWVZo2zeoLCLAvfpoxVpAWsVF5AhkzNawLSTJQBzfLG7QUDBAZEMmATFsEepKgqHuxChp3IxkEMsk6Qn-mqFw-hrbJ1linIjCxS25WlkoxQnw&pbk=qsQtGEMbCumiUX2CPo7bO0DmrIU57WGNDuq3tTPJxUY&spx=%2F"

	data, err := ParseVlessURL(link)
	if err != nil {
		t.Fatalf("Failed to parse Vless URL: %v", err)
	}

	config, err := data.Configuration(PriorInfo{})
	if err != nil {
		t.Fatalf("Failed to generate configuration: %v", err)
	}

	configJSON, err := json.MarshalIndent(config.CoreOutbound, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	configStr := string(configJSON)
	t.Logf("Generated config:\n%s", configStr)

	if strings.Contains(configStr, "allowInsecure") || strings.Contains(configStr, "AllowInsecure") {
		t.Errorf("Config contains forbidden 'allowInsecure' field!")
	}

	// Verify it contains the required reality fields
	required := []string{"realitySettings", "publicKey", "shortId", "spiderX", "serverName", "fingerprint"}
	for _, req := range required {
		if !strings.Contains(configStr, req) {
			t.Errorf("Config is missing required field: %s", req)
		}
	}
}
