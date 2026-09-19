//go:build !linux && !windows && !darwin

package v2ray

const (
	tunSupported = false
	// tunBindsEgress says the core binds its sockets to the physical
	// interface (no socket mark here).
	tunBindsEgress = false
	tunDeviceName  = "tun0"
)

func tunIPv6Enabled() bool { return false }

func tunRoutesUp(*Template, []string) error { return ErrTunUnsupported }

func tunDevicePresent() bool { return false }

func tunRoutesDown() {}

func tunInstalled() bool { return false }

func tunCleanupResidual() {}

// tunEgressInterface is the physical interface outbound sockets bind to on
// platforms without a socket mark; none here.
func tunEgressInterface() string { return "" }
