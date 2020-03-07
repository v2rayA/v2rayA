package routingA

import "strings"

func mergeMap(m1 *map[string]string, m2 map[string]string) {
	for k, v := range m2 {
		(*m1)[k] = v
	}
}

func symMatch(a []symbol, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].sym != b[i] {
			return false
		}
	}
	return true
}

func parseN(s symbol) (m map[string]string) {
	if s.sym != 'N' {
		return nil
	}
	m = make(map[string]string)
	if !symMatch(s.children, []rune(",H:HN")) {
		return
	}
	return parseHHN(s.Slice(1, s.Len()))
}

//H:HN
func parseHHN(s symbol) (m map[string]string) {
	if !symMatch(s.children, []rune("H:HN")) {
		return
	}
	m = make(map[string]string)
	m[s.children[0].val] = parseH(s.children[2])
	N := s.children[3]
	if len(N.children) > 0 {
		mm := parseN(N)
		mergeMap(&m, mm)
	}
	return
}

func parseH(s symbol) string {
	if s.sym != 'H' {
		return ""
	}
	switch {
	case (strings.HasPrefix(s.val, `'`) && strings.HasSuffix(s.val, `'`)) ||
		(strings.HasPrefix(s.val, `"`) && strings.HasSuffix(s.val, `"`)):
		return s.val[1 : len(s.val)-1]
	default:
		return s.val
	}
}

func parseM(s symbol) (params []string) {
	if s.sym != 'M' {
		return
	}
	if len(s.children) == 0 {
		return
	}
	params = append(params, parseH(s.children[1]))
	params = append(params, parseM(s.children[2])...)
	return
}

func parseG(s symbol) (params []string, namedParams map[string]string) {
	if s.sym != 'G' {
		return
	}
	params = make([]string, 0)
	namedParams = make(map[string]string)
	switch {
	case symMatch(s.children, []rune("HMN")):
		params = append(params, parseH(s.children[0]))
		params = append(params, parseM(s.children[1])...)
		namedParams = parseN(s.children[2])
	case symMatch(s.children, []rune("H:HN")):
		namedParams = parseHHN(s)
	}
	return
}

func parseQ(s symbol) (and []Function) {
	if s.sym != 'Q' {
		return
	}
	and = make([]Function, 0)
	if symMatch(s.children, []rune("&&FQ")) {
		and = append(and, *newFunction(s.children[2]))
		and = append(and, parseQ(s.children[3])...)
	}
	return
}

func parseR(s symbol) (As symbols) {
	if s.sym != 'R' {
		return
	}
	As = make(symbols, 0)
	switch {
	case symMatch(s.children, []rune("rAR")):
		As = append(As, s.children[1])
		As = append(As, parseR(s.children[2])...)
	}
	return
}

func parseS(s symbol) (As symbols) {
	if s.sym != 'S' {
		return
	}
	As = make(symbols, 0)
	As = append(As, s.children[0])
	As = append(As, parseR(s.children[1])...)
	return
}
