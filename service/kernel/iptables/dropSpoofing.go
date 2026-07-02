package iptables

type dropSpoofing struct{}

var DropSpoofing dropSpoofing

func (r *dropSpoofing) GetSetupCommands() Setter {
	commands := `
iptables -w 2 -N DROP_SPOOFING
iptables -w 2 -A DROP_SPOOFING -p udp --sport 53 -m string --algo bm --hex-string "|00047f|" --from 60 --to 180 -j DROP
iptables -w 2 -A DROP_SPOOFING -p udp --sport 53 -m string --algo bm --hex-string "|000400000000|" --from 60 --to 180 -j DROP

iptables -w 2 -A DROP_SPOOFING -p udp --sport 53 -m string --algo bm --hex-string "|001000000000000000000000000000000000|" --from 60 --to 180 -j DROP
iptables -w 2 -A DROP_SPOOFING -p udp --sport 53 -m string --algo bm --hex-string "|001000000000000000000000000000000001|" --from 60 --to 180 -j DROP
iptables -w 2 -A DROP_SPOOFING -p udp --sport 53 -m string --algo bm --hex-string "|0010fc|" --from 60 --to 180 -j DROP
iptables -w 2 -A DROP_SPOOFING -p udp --sport 53 -m string --algo bm --hex-string "|0010fd|" --from 60 --to 180 -j DROP
iptables -w 2 -A DROP_SPOOFING -p udp --sport 53 -m string --algo bm --hex-string "|0010ff|" --from 60 --to 180 -j DROP
iptables -w 2 -I INPUT -j DROP_SPOOFING
iptables -w 2 -I FORWARD -j DROP_SPOOFING
`
	if IsIPv6Supported() {
		commands += `
ip6tables -w 2 -N DROP_SPOOFING
ip6tables -w 2 -A DROP_SPOOFING -p udp --sport 53 -m string --algo bm --hex-string "|00047f|" --from 60 --to 180 -j DROP
ip6tables -w 2 -A DROP_SPOOFING -p udp --sport 53 -m string --algo bm --hex-string "|000400000000|" --from 60 --to 180 -j DROP

ip6tables -w 2 -A DROP_SPOOFING -p udp --sport 53 -m string --algo bm --hex-string "|001000000000000000000000000000000000|" --from 60 --to 180 -j DROP
ip6tables -w 2 -A DROP_SPOOFING -p udp --sport 53 -m string --algo bm --hex-string "|001000000000000000000000000000000001|" --from 60 --to 180 -j DROP
ip6tables -w 2 -A DROP_SPOOFING -p udp --sport 53 -m string --algo bm --hex-string "|0010fc|" --from 60 --to 180 -j DROP
ip6tables -w 2 -A DROP_SPOOFING -p udp --sport 53 -m string --algo bm --hex-string "|0010fd|" --from 60 --to 180 -j DROP
ip6tables -w 2 -A DROP_SPOOFING -p udp --sport 53 -m string --algo bm --hex-string "|0010ff|" --from 60 --to 180 -j DROP
ip6tables -w 2 -I INPUT -j DROP_SPOOFING
ip6tables -w 2 -I FORWARD -j DROP_SPOOFING
`
	}
	return Setter{
		Cmds: commands,
	}
}

func (r *dropSpoofing) GetCleanCommands() Setter {
	commands := `
iptables -w 2 -D INPUT -j DROP_SPOOFING
iptables -w 2 -D FORWARD -j DROP_SPOOFING
iptables -w 2 -F DROP_SPOOFING
iptables -w 2 -X DROP_SPOOFING
`
	if IsIPv6Supported() {
		commands += `
ip6tables -w 2 -D INPUT -j DROP_SPOOFING
ip6tables -w 2 -D FORWARD -j DROP_SPOOFING
ip6tables -w 2 -F DROP_SPOOFING
ip6tables -w 2 -X DROP_SPOOFING
`
	}
	return Setter{
		Cmds: commands,
	}
}
