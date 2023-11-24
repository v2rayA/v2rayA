package tun

import "io"

type Closer []io.Closer

func (c Closer) Close() error {
	for i := len(c) - 1; i >= 0; i-- {
		c[i].Close()
	}
	return nil
}
