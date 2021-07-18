// Static trie
package trie

import (
	"strings"
)

type cast map[rune]*next

type next struct {
	*node
	str *string
}

type node struct {
	c   cast
	end bool
}

type Trie struct {
	root *node
}

func newNode() *node {
	return &node{
		c:   cast{},
		end: false,
	}
}

func New(dict []string) (trie *Trie) {
	var t Trie
	var ok bool
	var p *node
	t.root = newNode()
	for _, d := range dict {
		p = t.root
		for i, r := range d {
			_, ok = p.c[r]
			if !ok {
				n := next{
					node: newNode(),
					str:  nil,
				}
				p.c[r] = &n
			}
			p = p.c[r].node
			if i == len(d)-1 {
				p.end = true
			}
		}
	}
	//make jump
	makeJump(t.root)
	return &t
}

func fastJump(from *next, to *next, str *string) {
	from.str = str
	from.node = to.node
}

func _makeJump(cur *next, from *next, builder *strings.Builder) {
	var fork bool
	if cur.node.end || len(cur.node.c) > 1 {
		if builder.Len() > 1 {
			s := builder.String()
			fastJump(from, cur, &s)
		}
		fork = true
	}
	for k := range cur.node.c {
		child := cur.node.c[k]
		if fork {
			from = child
			builder = new(strings.Builder)
		}
		builder.WriteRune(k)
		_makeJump(child, from, builder)
	}
}

func makeJump(root *node) {
	//DFS
	for k := range root.c {
		builder := new(strings.Builder)
		builder.WriteRune(k)
		_makeJump(root.c[k], root.c[k], builder)
	}
}

func (t *Trie) Match(str string) (prefix string) {
	var builder strings.Builder
	var runes = []rune(str)
	var length = len(runes)
	p := t.root
	for i := 0; i < length; i++ {
		r := runes[i]
		tmp, ok := p.c[r]
		if !ok {
			return
		}
		if tmp.str == nil {
			builder.WriteRune(r)
		} else {
			if lenTmp := len(*tmp.str); builder.Len()+lenTmp <= length {
				if string(runes[i:lenTmp+i]) == *tmp.str {
					builder.WriteString(*tmp.str)
					i += len(*tmp.str) - 1
				} else {
					break
				}
			}
		}
		if tmp.node.end {
			if builder.Len() <= length {
				prefix = builder.String()
				if len(prefix) == length {
					break
				}
			} else {
				break
			}
		}
		p = tmp.node
	}
	return
}
