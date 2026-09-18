// SPDX-License-Identifier: MPL-2.0
package tunmips

import (
	"context"
	"net/netip"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/xtls/xray-core/common/errors"
)

// owner is one process holding the socket a flow came from.
type owner struct {
	pid  uint32
	name string // executable basename, e.g. "curl" or "chrome.exe"
}

// ownerOf finds the processes holding the local socket behind a flow. proto
// is "tcp" or "udp"; src is the application's local address and port. A
// UDP socket bound to the wildcard address is matched by port alone when no
// exact match exists. Zero owners without an error means nothing on this
// host claims the socket right now: it may have closed already.
func ownerOf(proto string, src netip.AddrPort) ([]owner, error) {
	return platformOwnerOf(proto, src)
}

// exclusion decides which flows bypass the proxy: those owned by the core
// or v2rayA themselves (by PID, so a renamed binary still counts) and those
// owned by a process on the user's list (by executable basename).
type exclusion struct {
	names    map[string]struct{}
	selfPids map[uint32]struct{}
	// lookup is ownerOf unless a test replaces it.
	lookup func(proto string, src netip.AddrPort) ([]owner, error)
}

func newExclusion(names []string, selfPids []uint32) *exclusion {
	e := &exclusion{names: make(map[string]struct{}), selfPids: make(map[uint32]struct{}), lookup: ownerOf}
	for _, n := range names {
		if n = normalizeName(n); n != "" {
			e.names[n] = struct{}{}
		}
	}
	for _, p := range selfPids {
		e.selfPids[p] = struct{}{}
	}
	return e
}

func (e *exclusion) empty() bool {
	return len(e.names) == 0 && len(e.selfPids) == 0
}

// excluded reports whether the flow's owner is on the list. An unknown
// owner — lookup failed, or no process holds the socket any more — is
// treated as not excluded and the flow takes the normal route: proxying a
// packet that should have gone direct is the safer mistake.
func (e *exclusion) excluded(proto string, src netip.AddrPort) bool {
	if e == nil || e.empty() {
		return false
	}
	owners, err := e.lookup(proto, src)
	if err != nil {
		errors.LogDebug(context.Background(), "tun-mips: owner of ", proto, " ", src, " unknown: ", err)
		return false
	}
	if len(owners) == 0 {
		errors.LogDebug(context.Background(), "tun-mips: owner of ", proto, " ", src, " unknown: no socket")
		return false
	}
	for _, o := range owners {
		if _, ok := e.selfPids[o.pid]; ok {
			return true
		}
		if _, ok := e.names[normalizeName(o.name)]; ok {
			return true
		}
	}
	return false
}

// normalizeName reduces a process name to the form list entries are
// compared in: basename, and on Windows case-folded without ".exe" so that
// "Chrome" matches "chrome.exe".
func normalizeName(name string) string {
	name = strings.TrimSpace(name)
	name = filepath.Base(name)
	if name == "." || name == string(filepath.Separator) {
		return ""
	}
	if runtime.GOOS == "windows" {
		name = strings.ToLower(strings.TrimSuffix(strings.ToLower(name), ".exe"))
	}
	return name
}
