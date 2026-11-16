package dns

import "testing"

// An ext:"file.dat:tag" pattern (stored as ext:file.dat:tag) must reach the
// resolver as file.dat:tag, the same way a geosite: pattern hands over its
// tag; before this only the geosite: prefix was expanded, so an external
// geosite file named in a DNS rule matched nothing.
func TestBuildMatchersExpandsExtGeositeReferences(t *testing.T) {
	var got []string
	resolver := func(spec string) (domains, suffixes []string) {
		got = append(got, spec)
		return []string{"exact.example"}, []string{"suffix.example"}
	}
	rule := &DnsRule{
		Domain:       []string{"ext:rf-geosite.dat:RU-BLOCKED"},
		DomainSuffix: []string{"geosite:category-ru"},
	}
	matchers, err := BuildMatchers(rule, resolver)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"rf-geosite.dat:RU-BLOCKED", "category-ru"}
	if len(got) != len(want) {
		t.Fatalf("resolver got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("resolver got %v, want %v", got, want)
		}
	}
	if len(matchers) != 1 {
		t.Fatalf("matchers = %v, want a single OR group", matchers)
	}
	q := &DnsQuery{Name: "www.suffix.example."}
	if !MatchAll(matchers, q) {
		t.Fatalf("%q did not match the expanded suffix", q.Name)
	}
	q = &DnsQuery{Name: "exact.example"}
	if !MatchAll(matchers, q) {
		t.Fatalf("%q did not match the expanded exact domain", q.Name)
	}
	q = &DnsQuery{Name: "other.example"}
	if MatchAll(matchers, q) {
		t.Fatalf("%q matched nothing in the expansion", q.Name)
	}
}

// Without a resolver a geosite reference creates no matcher, as before.
func TestBuildMatchersSkipsGeositeWithoutResolver(t *testing.T) {
	matchers, err := BuildMatchers(&DnsRule{Domain: []string{"ext:rf-geosite.dat:RU-BLOCKED"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(matchers) != 0 {
		t.Fatalf("matchers = %v without a resolver", matchers)
	}
}
