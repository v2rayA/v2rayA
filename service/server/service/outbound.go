package service

import (
	"fmt"
	"net/url"
	"time"

	"github.com/v2rayA/v2rayA/db/configure"
)

func ValidateOutboundSetting(setting configure.OutboundSetting) error {
	probeURL, err := url.Parse(setting.ProbeURL)
	if err != nil || probeURL.Host == "" || (probeURL.Scheme != "http" && probeURL.Scheme != "https") {
		return fmt.Errorf("probe URL must be HTTP or HTTPS")
	}
	interval, err := time.ParseDuration(setting.ProbeInterval)
	if err != nil || interval < time.Second || interval > 365*24*time.Hour {
		return fmt.Errorf("probe interval must be between 1 second and 365 days")
	}
	switch setting.Type {
	case configure.LeastPing, configure.LeastLoad, configure.RoundRobin, configure.Random:
	default:
		return fmt.Errorf("unsupported group type")
	}
	return nil
}
