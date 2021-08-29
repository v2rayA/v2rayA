//go:build !linux
// +build !linux

package specialMode

type ExtraInfo struct {
	DohIps       []string
	DohDomains   []string
	ServerIps    []string
	ServerDomain string
}

func CouldUseSupervisor() bool {
	// TODO
	return true
}

func ShouldUseSupervisor() bool {
	return false
}

func CheckAndSetupDNSSupervisor() {
	return
}
func StopDNSSupervisor() {
	return
}
