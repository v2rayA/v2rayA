#!/bin/sh

if [ "$IPTABLES_MODE" = "nftables" ]; then
    /sbin/iptables-nft "$@"
elif [ "$IPTABLES_MODE" = "legacy" ]; then
    /sbin/iptables-legacy "$@"
else
    /sbin/iptables "$@"
fi
