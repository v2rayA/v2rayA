// Package dns provides a standalone DNS module for v2rayA.
//
// This file exists solely to ensure go.sum tracks the github.com/miekg/dns
// dependency, which is required by the DNS module. The actual imports are
// in types.go and listener.go.
//
// This file will be removed once the project's go.mod properly tracks
// all dependencies through normal build workflows.
package dns

import (
	_ "github.com/miekg/dns"
)
