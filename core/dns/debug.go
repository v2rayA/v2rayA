package dns

import (
	"log"
	"os"
	"strconv"
)

// dnsDebugEnabled controls per-query verbose logging. It is disabled by
// default because at info level every DNS query logs 2-3 lines; under a
// high query load (e.g. cloud-agent monitoring traffic) that log storm
// consumes CPU/IO/memory and can OOM the process. Enable with
// V2RAYA_DNS_DEBUG=1 when troubleshooting.
var dnsDebugEnabled = func() bool {
	v, err := strconv.ParseBool(os.Getenv("V2RAYA_DNS_DEBUG"))
	if err != nil {
		return false
	}
	return v
}()

// dnsLogf logs a line only when verbose DNS debug logging is enabled.
// High-frequency per-query logs MUST use this instead of log.Printf.
func dnsLogf(format string, args ...interface{}) {
	if dnsDebugEnabled {
		log.Printf(format, args...)
	}
}
