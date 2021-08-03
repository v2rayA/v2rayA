package ports

import (
	"github.com/v2rayA/v2rayA/common/netTools/netstat"
	"strconv"
	"strings"
)

func generatePortMap(syntax []string) (req map[int][]string, m map[string]map[int][]*netstat.Socket, err error) {
	rp := make([]string, 0, 4)
	udp := false
	tcp := false
	req = make(map[int][]string)
	for _, s := range syntax {
		a1 := strings.SplitN(s, ":", 2)
		p, err := strconv.Atoi(a1[0])
		if err != nil {
			continue
		}
		req[p] = make([]string, 0, 2)
		a2 := strings.Split(a1[1], ",")
		for _, proto := range a2 {
			switch strings.ToLower(proto) {
			case "tcp":
				req[p] = append(req[p], "tcp", "tcp6")
				if !tcp {
					tcp = true
				}
			case "udp":
				req[p] = append(req[p], "udp", "udp6")
				if !udp {
					udp = true
				}
			}
		}
	}
	if tcp {
		rp = append(rp, "tcp", "tcp6")
	}
	if udp {
		rp = append(rp, "udp", "udp6")
	}
	m, err = netstat.ToPortMap(rp)
	if err != nil {
		return
	}
	return
}

/*
example:

IsPortOccupied([]string{"80:tcp"})

IsPortOccupied([]string{"53:tcp,udp"})

IsPortOccupied([]string{"53:tcp,udp", "80:tcp"})
*/
func IsPortOccupied(syntax []string) (occupied bool, sockets []*netstat.Socket, err error) {
	req, m, err := generatePortMap(syntax)
	if err != nil {
		return
	}
	for p, protos := range req {
		for _, proto := range protos {
			for _, v := range m[proto][p] {
				if proto == "udp" || v.State != netstat.Close {
					occupied = true
					sockets = append(sockets, v)
				}
			}
		}
	}
	return occupied, sockets, nil
}

func IsOccupiedTCPPort(nsmap map[string]map[int][]*netstat.Socket, port int) bool {
	v := nsmap["tcp"][port]
	v6 := nsmap["tcp6"][port]
	v = append(v, v6...)
	for _, v := range v {
		if v.State != netstat.Close {
			return true
		}
	}
	return false
}
