#!/bin/sh

if [ "$IPTABLES_MODE" = "nftables" ]; then
    /sbin/ip6tables-nft "$@"
elif [ "$IPTABLES_MODE" = "legacy" ]; then
    /sbin/ip6tables-legacy "$@"
else
    /sbin/ip6tables "$@"
fi
