package proxy

import (
	"errors"
	"net"
	"net/url"
	"strings"

	"log"
)

// Server interface
type Server interface {
	// ListenAndServe sets up a listener and serve on it
	ListenAndServe() error

	// Serve serves a connection
	Serve(c net.Conn)
}

// ServerCreator is a function to create proxy servers
type ServerCreator func(s string, proxy Proxy) (Server, error)

var (
	serverMap = make(map[string]ServerCreator)
)

// RegisterServer is used to register a proxy server
func RegisterServer(name string, c ServerCreator) {
	serverMap[name] = c
}

// ServerFromURL calls the registered creator to create proxy servers
// dialer is the default upstream dialer so cannot be nil, we can use Default when calling this function
func ServerFromURL(s string, p Proxy) (Server, error) {
	if p == nil {
		return nil, errors.New("ServerFromURL: dialer cannot be nil")
	}

	if !strings.Contains(s, "://") {
		s = "mixed://" + s
	}

	u, err := url.Parse(s)
	if err != nil {
		log.Printf("parse err: %s\n", err)
		return nil, err
	}

	c, ok := serverMap[strings.ToLower(u.Scheme)]
	if ok {
		return c(s, p)
	}

	return nil, errors.New("unknown scheme '" + u.Scheme + "'")
}
