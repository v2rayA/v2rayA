package dns

import (
	"errors"
	"strings"

	"golang.org/x/sys/windows/registry"
)

const basePath = `SYSTEM\CurrentControlSet\Services\Tcpip\Parameters\Interfaces\`

func GetValidNetworkInterfaces() ([]string, error) {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, basePath, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return nil, err
	}
	defer key.Close()
	names, err := key.ReadSubKeyNames(0)
	if err != nil {
		return nil, err
	}
	var interfaces []string
	for _, name := range names {
		key, err = registry.OpenKey(registry.LOCAL_MACHINE, basePath+name, registry.READ)
		if err != nil {
			continue
		}
		if ip, _, _ := key.GetStringsValue("IPAddress"); len(ip) != 0 {
			interfaces = append(interfaces, name)
		} else if ip, _, _ := key.GetStringValue("DhcpIPAddress"); ip != "" {
			interfaces = append(interfaces, name)
		}
		key.Close()
	}
	return interfaces, nil
}

func GetDNSServer(ifi string) ([]string, error) {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, basePath+ifi, registry.READ)
	if err != nil {
		return nil, err
	}
	defer key.Close()
	server, _, err := key.GetStringValue("NameServer")
	if err != nil && !errors.Is(err, registry.ErrNotExist) {
		return nil, err
	}
	return strings.Split(server, ","), nil
}

func SetDNSServer(ifi string, server ...string) error {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, basePath+ifi, registry.WRITE)
	if err != nil {
		return err
	}
	defer key.Close()
	return key.SetStringValue("NameServer", strings.Join(server, ","))
}

func ReplaceDNSServer(ifi string, server ...string) ([]string, error) {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, basePath+ifi, registry.READ|registry.WRITE)
	if err != nil {
		return nil, err
	}
	defer key.Close()
	value, _, err := key.GetStringValue("NameServer")
	if err != nil && !errors.Is(err, registry.ErrNotExist) {
		return nil, err
	}
	return strings.Split(value, ","), key.SetStringValue("NameServer", strings.Join(server, ","))
}
