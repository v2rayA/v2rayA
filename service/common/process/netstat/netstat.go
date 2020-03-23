package netstat

import (
	"bufio"
	"encoding/binary"
	"errors"
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
		return nil, errors.New("Bad formatted string")
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

func IsSocketFreed(err error) bool {
	return err != nil && errors.Is(err, errors.New(SocketFreed))
}

/*
较为消耗资源
*/
func (s *Socket) Process() (*Process, error) {
	s.processMutex.Lock()
	s.processMutex.Unlock()
	if s.process != nil {
		return s.process, nil
	}
	f, err := os.Open(pathProc)
	if err != nil {
		return nil, nil
	}
	names, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
loop1:
	for _, fn := range names {
		p := filepath.Join(pathProc, fn)
		fi, err := os.Stat(p)
		if err != nil || !fi.IsDir() {
			continue
		}
		for _, t := range fn {
			if t > '9' || t < '0' {
				continue loop1
			}
		}
		if isProcessSocket(fn, s.inode) {
			s.process = &Process{
				PID:  fn,
				Name: getProcessName(fn),
			}
			return s.process, nil
		}
	}
	return nil, errors.New(SocketFreed)
}

/*
没有做缓存，每次调用都会扫描，消耗资源
*/
func findProcessID(pname string) (pid string, err error) {
	f, err := os.Open(pathProc)
	if err != nil {
		return
	}
	names, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return
	}
loop1:
	for _, fn := range names {
		p := filepath.Join(pathProc, fn)
		fi, err := os.Stat(p)
		if err != nil || !fi.IsDir() {
			continue
		}
		for _, t := range fn {
			if t > '9' || t < '0' {
				continue loop1
			}
		}
		if getProcessName(fn) == pname {
			return fn, nil
		}
	}
	return "", errors.New("not found")
}

func getProcessName(pid string) (pn string) {
	p := filepath.Join(pathProc, pid, "stat")
	b, err := ioutil.ReadFile(p)
	if err != nil {
		return
	}
	sp := strings.SplitN(string(b), " ", 3)
	pn = sp[1]
	return pn[1 : len(pn)-1]
}

func isProcessSocket(pid string, socketInode string) bool {
	// link name is of the form socket:[5860846]
	target := "socket:[" + socketInode + "]"
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
		if lk == target {
			return true
		}
	}
	return false
}
func parseSocktab(r io.Reader) (map[int]Socket, error) {
	br := bufio.NewScanner(r)
	tab := make(map[int]Socket)

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
		tab[s.LocalAddress.Port] = s
	}
	return tab, br.Err()
}
func ToPortMap(protocols []string) map[string]map[int]Socket {
	m := make(map[string]map[int]Socket)
	for _, proto := range protocols {
		switch proto {
		case "tcp", "tcp6", "udp", "udp6":
			b, err := os.Open(filepath.Join(pathNet, proto))
			if err != nil {
				continue
			}
			m[proto], _ = parseSocktab(b)
		default:
		}
	}
	return m
}

func IsProcessPort(pname string, port int, protocols []string) (is bool) {
	pid, err := findProcessID(pname)
	if err != nil {
		return
	}
	m := ToPortMap(protocols)
	for _, proto := range protocols {
		switch proto {
		case "tcp", "tcp6", "udp", "udp6":
			if v, ok := m[proto][port]; ok && isProcessSocket(pid, v.inode) {
				return true
			}
		default:
		}
	}
	return false
}
