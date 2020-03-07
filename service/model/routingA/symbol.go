package routingA

type symbol struct {
	sym      rune
	children []symbol
	val      string
}

func (s symbol) Slice(begin, end int) symbol {
	if begin < 0 || end > len(s.children) {
		panic("index of requested symbol slice exceeds range")
	}
	val := ""
	newChildren := s.children[begin:end]
	for _, s := range newChildren {
		val += s.val
	}
	return symbol{
		sym:      0,
		children: newChildren,
		val:      val,
	}
}
func (s symbol) Len() int {
	return len(s.children)
}

type symbols []symbol

func (ss symbols) Runes() (runes []rune) {
	runes = make([]rune, 0, len(ss))
	for _, s := range ss {
		runes = append(runes, s.sym)
	}
	return
}

func (ss symbols) String() (str string) {
	return string(ss.Runes())
}
