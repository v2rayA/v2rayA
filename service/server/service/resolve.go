package service

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
)

var EmptyAddressErr = common.Coded("LINK_EMPTY", fmt.Errorf("link is empty"), nil)
var InvalidURLErr = common.Coded("LINK_NO_SCHEME", fmt.Errorf("link has no scheme; expected vmess://..., ss://..., trojan://..., or a subscription URL starting with http(s)://"), nil)

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
