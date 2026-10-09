package configure

import "fmt"

// The DNS mode answers two questions the opt-out could not tell apart: does
// the DNS module run at all, and does the system send its queries to it.
// A setting written before the mode existed carries only dnsHijack, so an
// absent mode is derived from it, and the opt-out is then rewritten to mirror
// the mode: the two must never name different decisions.

// ResolveDnsMode maps a stored setting onto the mode it describes. An
// explicit mode is used as it is. A setting with no mode falls back to the
// opt-out, where anything but "no" means the historical interception. A mode
// this build does not know resolves to off: an unrecognised value must never
// silently turn interception on, and PutSetting refuses to store one.
func ResolveDnsMode(setting *Setting) DnsMode {
	if setting == nil {
		return DnsModeHijack
	}
	switch setting.DnsMode {
	case DnsModeOff:
		return DnsModeOff
	case DnsModeService:
		return DnsModeService
	case DnsModeHijack:
		return DnsModeHijack
	case "":
		if setting.DnsHijack == No {
			return DnsModeOff
		}
		return DnsModeHijack
	default:
		return DnsModeOff
	}
}

// DnsServiceEnabled reports whether the DNS module runs. Every mode but off
// starts it, independently of the transparent proxy.
func (s *Setting) DnsServiceEnabled() bool {
	return ResolveDnsMode(s) != DnsModeOff
}

// DnsInterceptionEnabled reports whether the system's own DNS is redirected
// to the module: the resolver hijack, the REDIRECT and TPROXY rules, and the
// TUN relay. Only hijack mode does that.
func (s *Setting) DnsInterceptionEnabled() bool {
	return ResolveDnsMode(s) == DnsModeHijack
}

// ValidateDnsMode rejects a mode this build cannot honour. An empty mode is
// accepted and resolved from dnsHijack; an unknown one is refused rather
// than quietly resolved, so a typo is reported instead of leaving the mode
// to decide whether the host's DNS is redirected.
func ValidateDnsMode(mode DnsMode) error {
	switch mode {
	case "", DnsModeOff, DnsModeService, DnsModeHijack:
		return nil
	}
	return fmt.Errorf("unknown dnsMode %q; use off, service or hijack", string(mode))
}
