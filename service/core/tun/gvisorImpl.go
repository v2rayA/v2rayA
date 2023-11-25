//go:build with_gvisor

package tun

import (
	"github.com/modern-go/reflect2"
	"github.com/sagernet/gvisor/pkg/tcpip/stack"
	tun "github.com/sagernet/sing-tun"
)

type gvisorCloser struct {
	stack tun.Stack
}

func (gc gvisorCloser) Close() error {
	err := gc.stack.Close()
	if _, ok := gc.stack.(*tun.GVisor); ok {
		typ := reflect2.TypeOfPtr(gc.stack).Elem().(*reflect2.UnsafeStructType)
		if field, ok := typ.FieldByName("stack").(*reflect2.UnsafeStructField); ok {
			value := field.UnsafeGet(reflect2.PtrOf(gc.stack))
			stack := *(**stack.Stack)(value)
			stack.Wait()
		}
	}
	return err
}
