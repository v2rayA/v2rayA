package plugin

import (
	"strings"
)

// The first is a server plugin, and the others are client plugins. Split by ",".
func ServerFromChain(chain string) (s Server, err error) {
	plugins := strings.Split(chain, ",")
	server := plugins[0]
	var dialer Dialer = &Direct{}
	for i := 1; i < len(plugins); i++ {
		dialer, err = DialerFromURL(plugins[i], dialer)
		if err != nil {
			return nil, err
		}
	}
	s, err = ServerFromURL(server, Dialer2Proxy(dialer))
	if err != nil {
		return nil, err
	}
	return s, nil
}
