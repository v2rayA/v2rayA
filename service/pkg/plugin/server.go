package plugin

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/v2rayA/v2rayA/pkg/util/log"
)

// Server interface
type Server interface {
	// ListenAndServe sets up a listener and serve on it
	ListenAndServe() error
	ListenAddr() string
	Close() error
}

// ServerCreator is a function to create proxy servers
type ServerCreator func(s string, proxy Proxy) (Server, error)

var (
	serverCreators = make(map[string]ServerCreator)
)

// RegisterServer is used to register a proxy server
func RegisterServer(name string, c ServerCreator) {
	serverCreators[name] = c
}

// ServerFromURL calls the registered creator to create proxy servers
// dialer is the default upstream dialer so cannot be nil, we can use Default when calling this function
func ServerFromURL(s string, p Proxy) (Server, error) {
	if p == nil {
		return nil, fmt.Errorf("ServerFromURL: dialer cannot be nil")
	}

	if !strings.Contains(s, "://") {
		s = "socks5://" + s
	}

	u, err := url.Parse(s)
	if err != nil {
		log.Warn("parse err: %s\n", err)
		return nil, err
	}

	c, ok := serverCreators[strings.ToLower(u.Scheme)]
	if ok {
		return c(s, p)
	}

	return nil, fmt.Errorf("unknown scheme '" + u.Scheme + "'")
}
