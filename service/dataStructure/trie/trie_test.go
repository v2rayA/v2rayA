package trie

import (
	"testing"
)

func TestTrie_Match(t *testing.T) {
	trie := New([]string{
		"12345",
		"123456",
		"2222",
		"1",
	})
	test := [][2]string{
		{"1", "1"},
		{"123", "1"},
		{"12345", "12345"},
		{"123456", "123456"},
		{"1234567", "123456"},
		{"222", ""},
		{"2222", "2222"},
	}
	for _, tt := range test {
		if p := trie.Match(tt[0]); p == tt[1] {
			t.Log(tt[0], "match prefix", p)
		} else {
			t.Fatal(tt)
		}
	}
}
