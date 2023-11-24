package dns

import (
	"bufio"
	"bytes"
	"fmt"
	"net/netip"
	"os/exec"
	"strings"
)

const networksetup = "/usr/sbin/networksetup"

func GetValidNetworkInterfaces() ([]string, error) {
	cmd := exec.Command(networksetup, "-listallnetworkservices")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("networksetup get: %v", err)
	}
	var interfaces []string
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "*") {
			continue
		}
		cmd = exec.Command(networksetup, "-getinfo", line)
		output, err = cmd.CombinedOutput()
		if err != nil {
			continue
		}
		valid := false
		scanner2 := bufio.NewScanner(bytes.NewReader(output))
		for scanner2.Scan() {
			line2 := scanner2.Text()
			key, value, ok := strings.Cut(line2, ": ")
			if !ok {
				continue
			}
			if key == "IP address" || key == "IPv6 IP address" {
				if _, err = netip.ParseAddr(value); err == nil {
					valid = true
					break
				}
			}
		}
		if valid {
			interfaces = append(interfaces, line)
		}
	}
	return interfaces, nil
}

func GetDNSServer(ifi string) ([]string, error) {
	cmd := exec.Command(networksetup, "-getdnsservers", ifi)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("networksetup get: %v", err)
	}
	var server []string
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if _, err = netip.ParseAddr(line); err == nil {
			server = append(server, line)
		}
	}
	return server, nil
}

func SetDNSServer(ifi string, server ...string) error {
	cmd := exec.Command(networksetup, "-setdnsservers", ifi)
	if len(server) == 0 {
		cmd.Args = append(cmd.Args, "Empty")
	} else {
		cmd.Args = append(cmd.Args, server...)
	}
	_, err := cmd.CombinedOutput()
	return err
}

func ReplaceDNSServer(ifi string, server ...string) ([]string, error) {
	old, err := GetDNSServer(ifi)
	if err != nil {
		return nil, err
	}
	return old, SetDNSServer(ifi, server...)
}
