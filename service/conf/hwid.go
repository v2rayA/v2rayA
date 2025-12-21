package conf

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/v2rayA/v2rayA/pkg/util/log"
)

var (
	hwidOnce sync.Once
	hwid     string
)

// GetHWID returns a unique hardware ID for this device.
// The HWID is generated once and stored persistently.
func GetHWID() string {
	hwidOnce.Do(func() {
		config := GetEnvironmentConfig()
		hwidPath := filepath.Join(config.Config, "hwid")
		
		// Try to read existing HWID
		if data, err := os.ReadFile(hwidPath); err == nil {
			hwid = string(data)
			if len(hwid) > 0 {
				return
			}
		}
		
		// Generate new HWID
		bytes := make([]byte, 16)
		if _, err := rand.Read(bytes); err != nil {
			log.Warn("failed to generate HWID: %v", err)
			// Fallback: use a simple hash based on config path
			hwid = fmt.Sprintf("%x", []byte(config.Config))
		} else {
			hwid = hex.EncodeToString(bytes)
		}
		
		// Save HWID to file
		if err := os.WriteFile(hwidPath, []byte(hwid), 0600); err != nil {
			log.Warn("failed to save HWID: %v", err)
		}
	})
	return hwid
}

