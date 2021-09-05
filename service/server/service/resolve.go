package service

import (
	"fmt"
	"github.com/v2rayA/v2rayA/core/serverObj"
	"net/url"
	"strings"
)

var EmptyAddressErr = fmt.Errorf("ResolveURL error: empty address")
var InvalidURLErr = fmt.Errorf("ResolveURL error: invalid URL")

func ResolveURL(u string) (n serverObj.ServerObj, err error) {
	u = strings.TrimSpace(u)
	if len(u) <= 0 {
		err = EmptyAddressErr
		return
	}
	U, err := url.Parse(strings.TrimSpace(u))
	if err != nil {
		return nil, err
	}
	if U.Scheme == "" {
		return nil, InvalidURLErr
	}
	return serverObj.NewFromLink(U.Scheme, u)
}
