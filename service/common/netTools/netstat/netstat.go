package netstat

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// Socket states
type SkState uint8

const (
	pathNet  = "/proc/net"
	pathProc = "/proc"

	ipv4StrLen = 8
	ipv6StrLen = 32
)

const (
	Established SkState = 0x01
	SynSent             = 0x02
	SynRecv             = 0x03
	FinWait1            = 0x04
	FinWait2            = 0x05
	TimeWait            = 0x06
	Close               = 0x07
	CloseWait           = 0x08
	LastAck             = 0x09
	Listen              = 0x0a
	Closing             = 0x0b
)

var skStates = [...]string{
	"UNKNOWN",
	"ESTABLISHED",
	"SYN_SENT",
	"SYN_RECV",
	"FIN_WAIT1",
	"FIN_WAIT2",
	"TIME_WAIT",
	"", // CLOSE
	"CLOSE_WAIT",
	"LAST_ACK",
	"LISTEN",
	"CLOSING",
}

func (sk SkState) String() string {
	return skStates[sk]
}

type Address struct {
	IP   net.IP
	Port int
}

func parseIPv4(s string) (net.IP, error) {
	v, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return nil, err
	}
	ip := make(net.IP, net.IPv4len)
	binary.LittleEndian.PutUint32(ip, uint32(v))
	return ip, nil
}

func parseIPv6(s string) (net.IP, error) {
	ip := make(net.IP, net.IPv6len)
	const grpLen = 4
	i, j := 0, 4
	for len(s) != 0 {
		grp := s[0:8]
		u, err := strconv.ParseUint(grp, 16, 32)
		binary.LittleEndian.PutUint32(ip[i:j], uint32(u))
		if err != nil {
			return nil, err
		}
		i, j = i+grpLen, j+grpLen
		s = s[8:]
	}
	return ip, nil
}

func parseAddr(s string) (*Address, error) {
	fields := strings.Split(s, ":")
	if len(fields) < 2 {
		return nil, fmt.Errorf("netstat: not enough fields: %v", s)
	}
	var ip net.IP
	var err error
	switch len(fields[0]) {
	case ipv4StrLen:
		ip, err = parseIPv4(fields[0])
	case ipv6StrLen:
		ip, err = parseIPv6(fields[0])
	default:
		return nil, newError("Bad formatted string")
	}
	if err != nil {
		return nil, err
	}
	v, err := strconv.ParseUint(fields[1], 16, 16)
	if err != nil {
		return nil, err
	}
	return &Address{IP: ip, Port: int(v),}, nil
}

type Socket struct {
	LocalAddress  *Address
	RemoteAddress *Address
	State         SkState
	UID           string
	inode         string
	process       *Process
	processMutex  sync.Mutex
}

type Process struct {
	PID  string
	Name string
}

const (
	SocketFreed = "process not found, correspond socket was freed"
)

/*
较为消耗资源
*/
func (s *Socket) Process() (*Process, error) {
	s.processMutex.Lock()
	s.processMutex.Unlock()
	if s.process != nil {
		return s.process, nil
	}
	f, err := ioutil.ReadDir(pathProc)
	if err != nil {
		return nil, nil
	}
loop1:
	for _, fi := range f {
		fn := fi.Name()
		if !fi.IsDir() {
			continue
		}
		for _, t := range fn {
			if t > '9' || t < '0' {
				continue loop1
			}
		}
		if isProcessSocket(fn, []string{s.inode}) {
			s.process = &Process{
				PID:  fn,
				Name: getProcessName(fn),
			}
			return s.process, nil
		}
	}
	return nil, newError(SocketFreed)
}

/*
没有做缓存，每次调用都会扫描，消耗资源
*/
func findProcessID(pname string) (pid string, err error) {
	f, err := ioutil.ReadDir(pathProc)
	if err != nil {
		return
	}
loop1:
	for _, fi := range f {
		if !fi.IsDir() {
			continue
		}
		fn := fi.Name()
		for _, t := range fn {
			if t > '9' || t < '0' {
				continue loop1
			}
		}
		if getProcessName(fn) == pname {
			return fn, nil
		}
	}
	return "", newError("not found")
}

func getProcName(s string) string {
	i := strings.Index(s, "(")
	if i < 0 {
		return ""
	}
	s = s[i+1:]
	j := strings.LastIndex(s, ")")
	if i < 0 {
		return ""
	}
	return s[:j]
}

func getProcessName(pid string) (pn string) {
	p := filepath.Join(pathProc, pid, "stat")
	b, err := ioutil.ReadFile(p)
	if err != nil {
		return
	}
	sp := bytes.SplitN(b, []byte(" "), 3)
	pn = string(sp[1])
	return getProcName(pn)
}

func isProcessSocket(pid string, socketInode []string) bool {
	// link name is of the form socket:[5860846]
	p := filepath.Join(pathProc, pid, "fd")
	f, err := os.Open(p)
	fns, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return false
	}
	for _, fn := range fns {
		lk, err := os.Readlink(filepath.Join(p, fn))
		if err != nil {
			continue
		}
		for _, s := range socketInode {
			target := "socket:[" + s + "]"
			if lk == target {
				return true
			}
		}
	}
	return false
}

func getProcessSocketSet(pid string) (set []string) {
	// link name is of the form socket:[5860846]
	p := filepath.Join(pathProc, pid, "fd")
	f, err := os.Open(p)
	fns, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return
	}
	for _, fn := range fns {
		lk, err := os.Readlink(filepath.Join(p, fn))
		if err != nil {
			continue
		}
		if strings.HasPrefix(lk, "socket:[") {
			set = append(set, lk[8:len(lk)-1])
		}
	}
	return
}

func parseSocktab(r io.Reader) (map[int][]*Socket, error) {
	br := bufio.NewScanner(r)
	tab := make(map[int][]*Socket)

	// Discard title
	br.Scan()

	for br.Scan() {
		var s Socket
		line := br.Text()
		// Skip comments
		if i := strings.Index(line, "#"); i >= 0 {
			line = line[:i]
		}
		fields := strings.Fields(line)
		if len(fields) < 12 {
			return tab, fmt.Errorf("netstat: not enough fields: %v, %v", len(fields), fields)
		}
		addr, err := parseAddr(fields[1])
		if err != nil {
			return tab, err
		}
		s.LocalAddress = addr
		addr, err = parseAddr(fields[2])
		if err != nil {
			return tab, err
		}
		s.RemoteAddress = addr
		u, err := strconv.ParseUint(fields[3], 16, 8)
		if err != nil {
			return tab, err
		}
		s.State = SkState(u)
		s.UID = fields[7]
		s.inode = fields[9]
		tab[s.LocalAddress.Port] = append(tab[s.LocalAddress.Port], &s)
	}
	return tab, br.Err()
}
func ToPortMap(protocols []string) (map[string]map[int][]*Socket, error) {
	m := make(map[string]map[int][]*Socket)
	for _, proto := range protocols {
		switch proto {
		case "tcp", "tcp6", "udp", "udp6":
			b, err := os.Open(filepath.Join(pathNet, proto))
			if err != nil {
				continue
			}
			m[proto], err = parseSocktab(b)
		default:
		}
	}
	return m, nil
}

func IsProcessListenPort(pname string, port int) (is bool, err error) {
	protocols := []string{"tcp", "tcp6"}
	m, err := ToPortMap(protocols)
	if err != nil {
		return
	}
	iNodes := make([]string, 2)
	for _, proto := range protocols {
		for _, v := range m[proto][port] {
			if v.State == Listen || v.State == Established {
				iNodes = append(iNodes, v.inode)
			}
		}
	}
	if len(iNodes) == 0 {
		return false, nil
	}
	pid, err := findProcessID(pname)
	if err != nil {
		return
	}
	return isProcessSocket(pid, iNodes), nil
}

func FillAllProcess(sockets []*Socket) {
	mInodeSocket := make(map[string]*Socket)
	for _, v := range sockets {
		if v.process == nil {
			mInodeSocket[v.inode] = v
			v.processMutex.Lock()
			defer v.processMutex.Unlock()
		}
	}
	f, err := ioutil.ReadDir(pathProc)
	if err != nil {
		return
	}
loop1:
	for _, fi := range f {
		if !fi.IsDir() {
			continue
		}
		fn := fi.Name()
		for _, t := range fn {
			if t > '9' || t < '0' {
				continue loop1
			}
		}
		socketSet := getProcessSocketSet(fn)
		for _, s := range socketSet {
			if socket, ok := mInodeSocket[s]; ok {
				socket.process = &Process{
					PID:  fn,
					Name: getProcessName(fn),
				}
			}
			delete(mInodeSocket, s)
		}
	}
}

func Print(protocols []string) string {
	var buffer strings.Builder
	protos := make([]string, 0, 4)
	for _, proto := range protocols {
		switch proto {
		case "tcp", "tcp6", "udp", "udp6":
			protos = append(protos, proto)
		}
	}
	m, err := ToPortMap(protos)
	if err != nil {
		return ""
	}
	buffer.WriteString(fmt.Sprintf("%-6v%-25v%-25v%-15v%-6v%-9v%v\n", "Proto", "Local Address", "Foreign Address", "State", "User", "Inode", "PID/Program name"))
	var sockets []*Socket
	for _, proto := range protos {
		for _, v := range m[proto] {
			sockets = append(sockets, v...)
		}
	}
	FillAllProcess(sockets)
	for _, proto := range protos {
		for _, sockets := range m[proto] {
			for _, v := range sockets {
				process, err := v.Process()
				var pstr string
				if err != nil {
					pstr = ""
				} else {
					pstr = process.PID + "/" + process.Name
				}
				buffer.WriteString(fmt.Sprintf(
					"%-6v%-25v%-25v%-15v%-6v%-9v%v\n",
					proto,
					v.LocalAddress.IP.String()+"/"+strconv.Itoa(v.LocalAddress.Port),
					v.RemoteAddress.IP.String()+"/"+strconv.Itoa(v.RemoteAddress.Port),
					v.State.String(),
					v.UID,
					v.inode,
					pstr,
				))
			}
		}
	}
	return buffer.String()
}
