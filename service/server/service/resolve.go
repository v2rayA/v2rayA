package service

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/v2rayA/v2rayA/kernel/serverObj"
)

var EmptyAddressErr = fmt.Errorf("link is empty")
var InvalidURLErr = fmt.Errorf("link has no scheme; expected vmess://..., ss://..., trojan://..., or a subscription URL starting with http(s)://")

func ResolveURL(u string) (n serverObj.ServerObj, err error) {
	u = strings.TrimSpace(u)
	if len(u) <= 0 {
		err = EmptyAddressErr
		return
	}
	U, err := url.Parse(strings.TrimSpace(u))
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err
		}
		return nil, fmt.Errorf("link is not a valid URL: %w", err)
	}
	if U.Scheme == "" {
		return nil, InvalidURLErr
	}
	return serverObj.NewFromLink(U.Scheme, u)
}
