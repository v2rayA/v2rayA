package netstat

import (
	"fmt"
	"net"
	"sync"
)

var ErrorNotSupportOSErr = fmt.Errorf("netstat: operating system is not supported")

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

// Socket states
type SkState uint8

func (sk SkState) String() string {
	return skStates[sk]
}

type Address struct {
	IP   net.IP
	Port int
}

type Socket struct {
	LocalAddress  *Address
	RemoteAddress *Address
	State         SkState
	UID           string
	inode         string
	Proc          *Process
	processMutex  sync.Mutex
}

type Process struct {
	PID  string
	PPID string
	Name string
}