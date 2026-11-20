package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
)

func hwidOrigin(u *url.URL) string {
	port := u.Port()
	if port == "" {
		port = "443"
	}
	return net.JoinHostPort(strings.ToLower(u.Hostname()), port)
}

func newSubscriptionHWID() string {
	var id [16]byte
	rand.Read(id[:])
	return hex.EncodeToString(id[:])
}

func attachSubscriptionHWID(client *http.Client, req *http.Request, id string) {
	origin := hwidOrigin(req.URL)
	req.Header.Set("X-HWID", id)
	check := client.CheckRedirect
	client.CheckRedirect = func(next *http.Request, via []*http.Request) error {
		if next.URL.Scheme != "https" || hwidOrigin(next.URL) != origin {
			next.Header.Del("X-HWID")
		}
		if check != nil {
			return check(next, via)
		}
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		return nil
	}
}
