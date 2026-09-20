# Quick start

v2rayA is used from this page in a browser. The service runs as root (or as an administrator on Windows) and starts `v2raya_core` for you; the browser only edits the configuration and shows the state.

## Sign in

The first account registered is the administrator; there is no other account. Registration needs no password, so on a shared network restrict who can reach port 2017 before you register, or run the service with `--address 127.0.0.1:2017` for local use only.

A forgotten password is reset from the command line, with the service stopped:

```sh
v2raya --reset-password
```

Pass `--config <directory>` as well when the service runs with one; `--lite` when it runs in lite mode.

## Import nodes

On the **Proxies** page, **Import** takes a share link (`vmess://`, `vless://`, `ss://`, `trojan://`, `hysteria2://`, `tuic://`, `juicity://`, `anytls://`, `wireguard://`, `socks5://`, `http://`, `https://`), a subscription address, several links pasted one per line, or a QR code image. ShadowsocksR links are refused.

Subscriptions keep their nodes together and can be updated by hand or on a schedule (**Settings → Automatically Update Subscriptions**). The mode setting next to it decides whether the update goes through the proxy.

## Groups

Nodes are used through **proxy groups**. The group `proxy` always exists; more can be added on the Proxies page. Add a node to a group from its menu or by selecting several. A group with several connected members balances by measured latency; **Edit group** on the dashboard pins one member so the group uses it alone, and the group settings set the probe address and interval.

The routing rules name groups as outbounds: `proxy` by default, and any other group by its name once it has a connected member.

## Start the core

**Start** on the dashboard, or the status button at the top of the page, starts `v2raya_core` with the current configuration. Every later change on the Settings page or in a group is applied by restarting the core for you.

Once the core runs, applications reach it through the local inbounds:

| Inbound | Address | Routing |
| --- | --- | --- |
| SOCKS5 | `127.0.0.1:20170` | as the transparent proxy setting, else everything through the proxy |
| HTTP | `127.0.0.1:20171` | same |
| HTTP with rules | `127.0.0.1:20172` | the **Traffic Splitting Mode of Rule Port** setting |

The ports are changed under **Settings → Address and Ports**; 0 closes an inbound. For traffic from every application without configuring each, turn on the transparent proxy.

## Routing

**Settings → Traffic splitting** picks how the rule port and the transparent proxy split traffic: proxy everything except Chinese sites, proxy only the GFWList, or your own RoutingA rules. RoutingA is described in its own section.
