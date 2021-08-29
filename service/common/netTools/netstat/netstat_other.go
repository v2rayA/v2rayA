// +build !linux,!plan9,!freebsd,!solaris

package netstat

func FillProcesses(sockets []*Socket) error {
	return ErrorNotSupportOSErr
}

func (s *Socket) Process() (*Process, error) {
	return nil, ErrorNotSupportOSErr
}


func ToPortMap(protocols []string) (map[string]map[int][]*Socket, error) {
	return nil, ErrorNotSupportOSErr
}

func IsProcessListenPort(pname string, port int) (is bool, err error) {
	return false, ErrorNotSupportOSErr
}

func Print(protocols []string) string {
	return ""
}
