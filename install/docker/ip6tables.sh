#!/bin/sh

if [ "$V2RAYA_NFTABLES_SUPPORT" != on ]; then
    if [ "$IPTABLES_MODE" = "nftables" ]; then
        /sbin/ip6tables-nft "$@"
    elif [ "$IPTABLES_MODE" = "legacy" ]; then
        /sbin/ip6tables-legacy "$@"
    else
        /sbin/ip6tables "$@"
    fi
fi