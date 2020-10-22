package iptables

type dropSpoofing struct{ iptablesSetter }

var DropSpoofing dropSpoofing

func (r *dropSpoofing) GetSetupCommands() SetupCommands {
	commands := `
iptables -N DROP_SPOOFING
iptables -A DROP_SPOOFING -p udp --sport 53 -m string --algo bm --hex-string "|00047f000001|" --from 60 --to 180 -j DROP
iptables -A DROP_SPOOFING -p udp --sport 53 -m string --algo bm --hex-string "|000400000000|" --from 60 --to 180 -j DROP
iptables -A INPUT -j DROP_SPOOFING 
iptables -A FORWARD -j DROP_SPOOFING
`
	return SetupCommands(commands)
}

func (r *dropSpoofing) GetCleanCommands() CleanCommands {
	commands := `
iptables -D INPUT -j DROP_SPOOFING 
iptables -D FORWARD -j DROP_SPOOFING 
iptables -F DROP_SPOOFING
iptables -X DROP_SPOOFING
`
	return CleanCommands(commands)
}
