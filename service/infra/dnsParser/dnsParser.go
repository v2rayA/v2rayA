package dnsParser

import "strings"

type Dns struct {
	Val string
	Out string
}

func Parse(dnsLine string) *Dns {
	dnsLine = strings.TrimSpace(dnsLine)
	index := strings.LastIndex(dnsLine, "->")
	if index >= 0 {
		return &Dns{
			Val: strings.TrimSpace(dnsLine[:index]),
			Out: strings.TrimSpace(dnsLine[index+2:]),
		}
	}
	return nil
}
