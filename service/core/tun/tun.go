package tun

type Stack string

const (
	StackGvisor = Stack("gvisor")
	StackSystem = Stack("system")
)

type Tun interface {
	Start(stack Stack) error
	Close() error
}

var Default = NewSingTun()
