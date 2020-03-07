package routingA

/*
outbound: httpout = http(address: 127.0.0.1, port: 8080, user: 'my-username', pass: 'my-password')
default: proxy
*/
type Define struct {
	Name  string
	Value interface{}
}

func newDefine(s symbol) (d *Define) {
	if s.sym != 'B' || !symMatch(s.children, []rune("D:E")) {
		return nil
	}
	E := s.children[2]
	d = new(Define)
	d.Name = s.children[0].val
	switch {
	case symMatch(E.children, []rune("D=F")):
		d.Value = *newOutbound(E)
	case symMatch(E.children, []rune("D")):
		d.Value = E.children[0].val
	default:
		return nil
	}
	return
}

/*
httpout = http(address: 127.0.0.1, port: 8080, user: 'my-username', pass: 'my-password')
*/
type Outbound struct {
	Name  string
	Value Function
}

func newOutbound(s symbol) (o *Outbound) {
	if s.sym != 'E' || !symMatch(s.children, []rune("D=F")) {
		return nil
	}
	o = new(Outbound)
	o.Name = s.children[0].val
	o.Value = *newFunction(s.children[2])
	return
}

/*
http(address: 127.0.0.1, port: 8080, user: 'my-username', pass: 'my-password')
*/
type Function struct {
	Name        string
	Params      []string
	NamedParams map[string][]string
}

func newFunction(s symbol) (f *Function) {
	if s.sym != 'F' {
		return nil
	}
	f = new(Function)
	f.Name = s.children[0].val
	G := s.children[2]
	f.Params, f.NamedParams = parseG(G)
	return
}

/*
domain(domain: v2raya.mzz.pub) -> socksout
*/
type Routing struct {
	And []Function
	Out string
}

func newRouting(s symbol) (r *Routing) {
	if s.sym != 'C' || !symMatch(s.children, []rune("FQ->D")) {
		return nil
	}
	r = new(Routing)
	r.Out = s.children[4].val
	r.And = append(r.And, *newFunction(s.children[0]))
	r.And = append(r.And, parseQ(s.children[1])...)
	return
}
