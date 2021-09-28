package configure

type ObservatoryType string

func (t ObservatoryType) String() string {
	return string(t)
}

const (
	LeastPing ObservatoryType = "leastPing"
)

type OutboundSetting struct {
	ProbeURL      string
	ProbeInterval string
	Type          ObservatoryType
}
