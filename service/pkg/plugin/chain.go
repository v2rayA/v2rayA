package plugin

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/v2rayA/v2rayA/pkg/util/log"
)

// The first is a server plugin, and the others are client plugins. Split by ",".
func ServerFromChain(chain string, nodeName string) (s Server, err error) {
	log.Trace("[plugin] building server from chain: %s (node: %s)", chain, nodeName)
	plugins := strings.Split(chain, ",")
	if len(plugins) == 0 {
		return nil, fmt.Errorf("empty plugin chain")
	}

	server := plugins[0]
	log.Trace("[plugin] server plugin: %s (node: %s)", server, nodeName)

	var dialer Dialer = &Direct{}
	for i := 1; i < len(plugins); i++ {
		log.Trace("[plugin] adding client plugin [%d/%d]: %s (node: %s)", i, len(plugins)-1, plugins[i], nodeName)
		dialer, err = DialerFromURL(plugins[i], dialer)
		if err != nil {
			log.Warn("[plugin] failed to build client plugin chain at step %d (node: %s): %v", i, nodeName, err)
			return nil, fmt.Errorf("build client plugin %d (%s): %w", i, plugins[i], err)
		}
	}

	log.Trace("[plugin] creating server from URL: %s (node: %s)", server, nodeName)

	// Extract protocol from the first client plugin (outbound protocol)
	// The server is the inbound protocol, client plugins are the outbound protocols
	protocol := "direct"
	if len(plugins) > 1 {
		// Use the first client plugin's protocol as it's the actual outbound protocol
		if u, err := url.Parse(plugins[1]); err == nil && u.Scheme != "" {
			protocol = u.Scheme
		}
	}

	s, err = ServerFromURL(server, nodeName, Dialer2Proxy(dialer, nodeName, protocol))
	if err != nil {
		log.Warn("[plugin] failed to create server (node: %s): %v", nodeName, err)
		return nil, fmt.Errorf("create server: %w", err)
	}

	log.Trace("[plugin] successfully built server chain (node: %s)", nodeName)
	return s, nil
}
