package trie

import (
	"testing"
)

func TestTrie_Match(t *testing.T) {
	trie := New([]string{
		"12",
		"12345",
		"1234567",
		"2222",
		"1",
	})
	test := [][2]string{
		{"1", "1"},
		{"123", "12"},
		{"1233", "12"},
		{"12345", "12345"},
		{"123456", "12345"},
		{"1234567", "1234567"},
		{"123456789", "1234567"},
		{"222", ""},
		{"2222", "2222"},
		{"22222", "2222"},
		{"122", "12"},
	}
	for _, tt := range test {
		if p := trie.Match(tt[0]); p == tt[1] {
			t.Log(tt[0], "match prefix", p)
		} else {
			t.Error(tt[0], "expect", tt[1], "wrong prefix", p)
		}
	}
}
