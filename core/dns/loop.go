package dns

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net"
	"sync"
	"time"

	"github.com/miekg/dns"
)

// A query that leaves this module carries a token in an EDNS0 local option. If
// a query arrives carrying the same token, it is one of ours that an upstream
// sent back — the classic configuration where a local resolver forwards to
// v2rayA while a v2rayA rule points at that resolver. Without this check the
// two forward the same name to each other until one of them runs out of
// memory. The firewall-level self-redirect is handled by the 0x80 mark; this
// is the resolver-level counterpart.
//
// 65042 is inside the local/experimental range of EDNS0 option codes
// (65001-65534) reserved by RFC 6891 for exactly this kind of private use.
const loopOptionCode = 65042

var (
	loopTokenOnce  sync.Once
	loopTokenValue []byte
)

// loopToken returns this process's token, generated on first use.
func loopToken() []byte {
	loopTokenOnce.Do(func() {
		b := make([]byte, 8)
		if _, err := rand.Read(b); err != nil {
			// A fixed token is still better than none: the only cost of a
			// collision between two v2rayA instances is a refused query on a
			// path that is already looping.
			copy(b, []byte("v2rayadn"))
		}
		loopTokenValue = b
	})
	return loopTokenValue
}

// markOutgoing attaches this process's token to a query that is about to be
// sent upstream. It keeps the single OPT record miekg/dns expects.
func markOutgoing(msg *dns.Msg) {
	if msg == nil {
		return
	}
	opt := msg.IsEdns0()
	if opt == nil {
		// No DNSSEC OK: the module does not validate, and a client that did
		// not ask for signatures must not get RRSIGs back.
		msg.SetEdns0(4096, false)
		opt = msg.IsEdns0()
		if opt == nil {
			return
		}
	}
	for _, o := range opt.Option {
		if local, ok := o.(*dns.EDNS0_LOCAL); ok && local.Code == loopOptionCode {
			local.Data = loopToken()
			return
		}
	}
	opt.Option = append(opt.Option, &dns.EDNS0_LOCAL{
		Code: loopOptionCode,
		Data: loopToken(),
	})
}

// carriesOwnToken reports whether msg is a query this process already sent
// upstream, which means it has come back around a resolver loop.
func carriesOwnToken(msg *dns.Msg) bool {
	if msg == nil {
		return false
	}
	opt := msg.IsEdns0()
	if opt == nil {
		return false
	}
	token := loopToken()
	for _, o := range opt.Option {
		local, ok := o.(*dns.EDNS0_LOCAL)
		if !ok || local.Code != loopOptionCode || len(local.Data) != len(token) {
			continue
		}
		same := true
		for i := range token {
			if local.Data[i] != token[i] {
				same = false
				break
			}
		}
		if same {
			return true
		}
	}
	return false
}

// loopTokenHex is used in log lines and tests.
func loopTokenHex() string {
	return hex.EncodeToString(loopToken())
}

// loopDetected logs the loop at most once every loopLogInterval: the two
// resolvers can exchange thousands of queries a second, and a log line per
// query is its own denial of service.
func loopDetected(q dns.Question, client net.Addr) {
	loopLogMu.Lock()
	defer loopLogMu.Unlock()
	loopSuppressed++
	if !loopLastLog.IsZero() && time.Since(loopLastLog) < loopLogInterval {
		return
	}
	log.Printf("[dns] loop detected: %s %s came back from %v carrying this instance's token %s; "+
		"check that the upstream for this name is not a resolver that forwards back to v2rayA "+
		"(%d such queries refused so far)",
		dns.Type(q.Qtype).String(), q.Name, client, loopTokenHex(), loopSuppressed)
	loopLastLog = time.Now()
}

const loopLogInterval = 10 * time.Second

var (
	loopLogMu      sync.Mutex
	loopLastLog    time.Time
	loopSuppressed int
)
