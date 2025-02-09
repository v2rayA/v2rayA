#!/bin/sh

if [ "$IPTABLES_MODE" = "nftables" ]; then
    /usr/sbin/ip6tables-nft "$@"
elif [ "$IPTABLES_MODE" = "legacy" ]; then
    /usr/sbin/ip6tables-legacy "$@"
else
    /usr/sbin/ip6tables "$@"
fi
