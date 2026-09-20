package dns

import (
	"fmt"
	"strings"

	"github.com/miekg/dns"
)

// QTypeToString returns "A", "AAAA", … or "TYPE99".
func QTypeToString(qt QueryType) string {
	return dns.Type(qt).String()
}

// StringToQType accepts a type name, "TYPE<n>" or a number.
func StringToQType(s string) (QueryType, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("dns qtype: empty string")
	}
	upper := strings.ToUpper(s)
	if qt, ok := dns.StringToType[upper]; ok {
		return QueryType(qt), nil
	}
	var num uint16
	if _, err := fmt.Sscanf(upper, "TYPE%d", &num); err == nil {
		return QueryType(num), nil
	}
	if _, err := fmt.Sscanf(upper, "%d", &num); err == nil {
		return QueryType(num), nil
	}
	return 0, fmt.Errorf("dns qtype: unknown query type: %q", s)
}

// SupportsQType returns true if the given QueryType is one of the
// supported types (A, AAAA, CNAME, TXT, MX, SRV, NS, PTR, SOA).
func SupportsQType(qt QueryType) bool {
	switch qt {
	case TypeA, TypeAAAA, TypeCNAME, TypeTXT,
		TypeMX, TypeSRV, TypeNS, TypePTR, TypeSOA:
		return true
	default:
		return false
	}
}

// SupportedQTypes returns a list of all supported query types.
func SupportedQTypes() []QueryType {
	return []QueryType{
		TypeA,
		TypeAAAA,
		TypeCNAME,
		TypeTXT,
		TypeMX,
		TypeSRV,
		TypeNS,
		TypePTR,
		TypeSOA,
	}
}

// IsAddressQuery returns true if the query type is A or AAAA.
func IsAddressQuery(qt QueryType) bool {
	return qt == TypeA || qt == TypeAAAA
}

// NeedsDNS64 returns true if DNS64 synthesis should be considered
// for this query type (i.e., AAAA query).
func NeedsDNS64(qt QueryType) bool {
	return qt == TypeAAAA
}
