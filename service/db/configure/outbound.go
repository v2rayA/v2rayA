package configure

import (
	"crypto/sha256"
	"encoding/hex"
)

type ObservatoryType string

func (t ObservatoryType) String() string {
	return string(t)
}

const (
	LeastPing   ObservatoryType = "leastping"
	KeepCurrent ObservatoryType = "keepcurrent"
	RoundRobin  ObservatoryType = "roundrobin"
	Random      ObservatoryType = "random"
)

type OutboundSetting struct {
	AutoAdd       bool            `json:"autoAdd"`
	ProbeURL      string          `json:"probeURL"`
	ProbeInterval string          `json:"probeInterval"`
	Type          ObservatoryType `json:"type"`
	// Selected is the share link of the member the group routes through
	// alone; empty means every member, balanced by the observatory. A link
	// that matches no member is ignored, so a removed or renamed node
	// falls back to balancing instead of breaking the group.
	Selected string `json:"selected,omitempty"`
	// StickyCurrent is worker-owned state for KeepCurrent. It stores a hash
	// instead of a share link so the internal choice does not duplicate node
	// credentials in API responses or logs.
	StickyCurrent string `json:"stickyCurrent,omitempty"`
}

func NodeFingerprint(link string) string {
	sum := sha256.Sum256([]byte(link))
	return hex.EncodeToString(sum[:])
}

func HasAutomaticGroup() bool {
	for _, name := range GetOutbounds() {
		if GetOutboundSetting(name).AutoAdd {
			return true
		}
	}
	return false
}

// HasFailClosedGroup reports whether an empty group still needs a core so its
// traffic can terminate at a blackhole instead of falling through elsewhere.
func HasFailClosedGroup() bool {
	for _, name := range GetOutbounds() {
		setting := GetOutboundSetting(name)
		if setting.AutoAdd || setting.Type == KeepCurrent {
			return true
		}
	}
	return false
}

// DefaultOutboundSetting returns an OutboundSetting with default values.
func DefaultOutboundSetting() OutboundSetting {
	return OutboundSetting{
		ProbeURL:      DefaultProbeURL,
		ProbeInterval: DefaultProbeInterval,
		Type:          ObservatoryType(DefaultOutboundType),
	}
}
