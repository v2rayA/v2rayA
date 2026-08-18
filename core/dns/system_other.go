//go:build !windows && !darwin

package dns

import "errors"

func GetValidNetworkInterfaces() ([]string, error) {
	return nil, errors.New("not implemented")
}

func GetDNSServer(ifi string) ([]string, error) {
	return nil, errors.New("not implemented")
}

func SetDNSServer(ifi string, server ...string) error {
	return errors.New("not implemented")
}

func ReplaceDNSServer(ifi string, server ...string) ([]string, error) {
	return nil, errors.New("not implemented")
}
