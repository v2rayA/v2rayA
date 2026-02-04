package plugin

import (
	"fmt"
	"strings"

	"github.com/v2rayA/v2rayA/pkg/util/log"
)

// The first is a server plugin, and the others are client plugins. Split by ",".
func ServerFromChain(chain string) (s Server, err error) {
	log.Trace("[plugin] building server from chain: %s", chain)
	plugins := strings.Split(chain, ",")
	if len(plugins) == 0 {
		return nil, fmt.Errorf("empty plugin chain")
	}

	server := plugins[0]
	log.Trace("[plugin] server plugin: %s", server)

	var dialer Dialer = &Direct{}
	for i := 1; i < len(plugins); i++ {
		log.Trace("[plugin] adding client plugin [%d/%d]: %s", i, len(plugins)-1, plugins[i])
		dialer, err = DialerFromURL(plugins[i], dialer)
		if err != nil {
			log.Warn("[plugin] failed to build client plugin chain at step %d: %v", i, err)
			return nil, fmt.Errorf("build client plugin %d (%s): %w", i, plugins[i], err)
		}
	}

	log.Trace("[plugin] creating server from URL: %s", server)
	s, err = ServerFromURL(server, Dialer2Proxy(dialer))
	if err != nil {
		log.Warn("[plugin] failed to create server: %v", err)
		return nil, fmt.Errorf("create server: %w", err)
	}

	log.Trace("[plugin] successfully built server chain")
	return s, nil
}
