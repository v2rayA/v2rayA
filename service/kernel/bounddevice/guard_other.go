//go:build !linux

package bounddevice

import "fmt"

type Guard struct{}

func Start(string) (*Guard, error) {
	return nil, fmt.Errorf("bound-device REDIRECT bypass requires Linux cgroup BPF")
}

func (*Guard) Close() error { return nil }
