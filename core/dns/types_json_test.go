package dns

import (
	"encoding/json"
	"testing"
)

func TestRuleConfigQueryTypeUnmarshalJSON(t *testing.T) {
	var rule RuleConfig
	if err := json.Unmarshal([]byte(`{"query_type":["A","AAAA",15]}`), &rule); err != nil {
		t.Fatalf("unmarshal rule config: %v", err)
	}

	want := []QueryType{TypeA, TypeAAAA, TypeMX}
	if len(rule.QueryType) != len(want) {
		t.Fatalf("query type count = %d, want %d", len(rule.QueryType), len(want))
	}
	for i, qtype := range want {
		if rule.QueryType[i] != qtype {
			t.Errorf("query type[%d] = %d, want %d", i, rule.QueryType[i], qtype)
		}
	}
}
