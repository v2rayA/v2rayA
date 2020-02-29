package trie

type cast map[rune]*node

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
				p.c[r] = newNode()
			}
			p = p.c[r]
			if i == len(d)-1 {
				p.end = true
			}
		}
	}
	return &t
}

func (t *Trie) Match(str string) (prefix string) {
	var arr []rune
	p := t.root
	for _, r := range str {
		tmp, ok := p.c[r]
		if !ok {
			return
		}
		arr = append(arr, r)
		if tmp.end {
			prefix = string(arr)
		}
		p = tmp
	}
	return
}
