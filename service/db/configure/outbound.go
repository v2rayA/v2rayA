package configure

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/v2rayA/v2rayA/kernel/serverObj"
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
	Fixed       ObservatoryType = "fixed"
)

func UsesWorkerSelection(strategy ObservatoryType) bool {
	switch strategy {
	case LeastPing, KeepCurrent, Random:
		return true
	default:
		return false
	}
}

func UsesWorkerProbe(strategy ObservatoryType) bool {
	return UsesWorkerSelection(strategy) || strategy == RoundRobin
}

type OutboundSetting struct {
	AutoAdd              bool            `json:"autoAdd"`
	ProbeURL             string          `json:"probeURL"`
	ProbeInterval        string          `json:"probeInterval"`
	Type                 ObservatoryType `json:"type"`
	CatalogRevision      string          `json:"catalogRevision,omitempty"`
	SelectionInvalidated bool            `json:"selectionInvalidated,omitempty"`
	// Selected is an explicit manual pin, independent of the automatic policy.
	// A link that matches no member leaves that group empty and fail-closed.
	Selected string `json:"selected,omitempty"`
	// StickyCurrent is worker-owned state for strategies that select one
	// member. It stores a hash
	// instead of a share link so the internal choice does not duplicate node
	// credentials in API responses or logs.
	StickyCurrent string `json:"stickyCurrent,omitempty"`
	// EligibleMembers caches worker-measured healthy fingerprints for every
	// automatic strategy. RoundRobin uses them to balance multiple nodes.
	EligibleMembers string `json:"eligibleMembers,omitempty"`
}

// NodeFingerprint identifies connection parameters, independently of display names.
func NodeFingerprint(link string) string {
	scheme, _, _ := strings.Cut(link, "://")
	if node, err := serverObj.NewFromLink(scheme, link); err == nil {
		node.SetName("")
		link = node.ExportToURL()
	}
	sum := sha256.Sum256([]byte(link))
	return hex.EncodeToString(sum[:])
}

// Accept old persisted fingerprints until the next successful worker update.
func MatchesNodeFingerprint(fingerprint, link string) bool {
	if fingerprint == "" {
		return false
	}
	legacy := sha256.Sum256([]byte(link))
	return fingerprint == NodeFingerprint(link) || fingerprint == hex.EncodeToString(legacy[:])
}

// MigrateNodeFingerprints runs before workers or subscription updates can rename
// nodes. Unknown fingerprints stay unknown, so migration cannot select a node.
func MigrateNodeFingerprints() error {
	for _, name := range GetOutbounds() {
		setting := GetOutboundSetting(name)
		previous := setting
		eligible := strings.Fields(setting.EligibleMembers)
		for _, ref := range GetConnectedServersByOutbound(name).Get() {
			raw, err := ref.LocateServerRaw()
			if err != nil || raw.ServerObj == nil {
				continue
			}
			link := raw.ServerObj.ExportToURL()
			fingerprint := NodeFingerprint(link)
			if MatchesNodeFingerprint(setting.StickyCurrent, link) {
				setting.StickyCurrent = fingerprint
			}
			for i, old := range eligible {
				if MatchesNodeFingerprint(old, link) {
					eligible[i] = fingerprint
				}
			}
		}
		setting.EligibleMembers = strings.Join(eligible, " ")
		if setting != previous {
			if err := SetOutboundSetting(name, setting); err != nil {
				return err
			}
		}
	}
	return nil
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
		if setting.AutoAdd || UsesWorkerProbe(setting.Type) || setting.Type == Fixed {
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
