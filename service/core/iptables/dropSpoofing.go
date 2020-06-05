package iptables

type dropSpoofing struct{ iptablesSetter }

var DropSpoofing dropSpoofing

func (r *dropSpoofing) GetSetupCommands() SetupCommands {
	commands := `
iptables -N DROP_SPOOFING
BADIP="127.0.0.1 0.0.0.0";for IP in $BADIP ;do hexip=$(printf '%02X ' ${IP//./ }; echo) ;iptables -A DROP_SPOOFING -p udp --sport 53 -m string --algo bm --hex-string "|0004$hexip|" --from 60 --to 180 -j DROP ;done
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
