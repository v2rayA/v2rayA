# Quick start

This page in the browser configures v2rayA and shows its state; the service runs as root (as an administrator on Windows; unprivileged with `--lite`) and starts `v2raya_core` itself.

## Sign in

The first account registered is the administrator; there is no other account. Registration asks for a username and a password of 6 to 32 characters and needs no existing credentials, so on a shared network restrict who can reach port 2017 before you register, or run the service with `--address 127.0.0.1:2017` for local use only.

A forgotten password is reset from the command line, with the service stopped:

```sh
v2raya --reset-password
```

Run it with the account and the `--config` directory the service uses (`--lite` too in lite mode); on Windows the service runs as SYSTEM and its directory differs from an elevated user's. Start the service again and register a new account.

## Import nodes

On the **Proxies** page, **Import** takes share links or a subscription: choose **Server link** for `vmess://`, `vless://`, `ss://`, `trojan://`, `hysteria2://`, `tuic://`, `juicity://`, `anytls://`, `wireguard://`, `socks5://`, `http://` and `https://` links, one per line, or a QR code image; choose **Subscription address** for a subscription. ShadowsocksR links are refused, and so are Shadowsocks links with a stream cipher (`rc4-md5`, `aes-*-cfb`, `chacha20-ietf`) or `none`: the core accepts the AEAD ciphers and the 2022-blake3 methods only. `allow_insecure` in a link is ignored: v2rayA never skips certificate verification; for a self-signed server, pin its certificate SHA-256 in the node's form.

Subscriptions have their own page, **Subscriptions** (on a phone the docs move to the app bar's menu to make room for it). They keep their nodes together and can be updated by hand or on a schedule (**Settings → Automatically Update Subscriptions**). The mode setting next to it decides whether the update goes through the proxy.

## Groups

Nodes are used through proxy groups. The group `proxy` always exists; more are created, renamed, configured and deleted from the group button in the app bar, the only place for that. A node joins a group from its menu, or select several on the Proxies page and pick the group under **Add to proxy group**. A group with several connected members uses the one with the lowest measured latency (**Auto (lowest latency)**); the node menu on the dashboard pins one member so the group uses it alone, **Add or remove nodes** changes the members, and the group settings set the probe address and interval.

The routing rules name groups as outbounds: `proxy` by default, and any other group by its name once it has a connected member.

## Start the core

**Start** on the dashboard, or the status button at the top of the page, starts `v2raya_core` with the current configuration. Changes on the Settings page take effect with **Save and Apply**: a running core is restarted with them, a stopped core picks them up on the next start.

Once the core runs, applications reach it through the local inbounds:

| Inbound         | Address           | Routing                                             |
| --------------- | ----------------- | --------------------------------------------------- |
| SOCKS5          | `127.0.0.1:20170` | everything through `proxy`                          |
| HTTP            | `127.0.0.1:20171` | everything through `proxy`                          |
| HTTP with rules | `127.0.0.1:20172` | the **Traffic Splitting Mode of Rule Port** setting |

The ports are changed under **Settings → Address and Ports**; 0 closes an inbound. To cover applications without configuring each, turn on the transparent proxy.

## Routing

**Settings → Traffic splitting** picks how the rule port and the transparent proxy split traffic: proxy everything except Chinese sites, proxy only the GFWList, or your own RoutingA rules. RoutingA is described in its own section.

## Keyboard

- Proxies, list view: `Ctrl`+`A` selects every listed node, `Esc` clears the selection; `/` goes to the search on the Proxies and Logs pages.
- Settings: `Ctrl`+`S` saves. RoutingA editor: `Ctrl`+`S` saves, `Tab` indents.
- Dialogs with a text area (import, subscription, route scripts, lists): `Ctrl`+`Enter` submits; `Esc` closes any dialog.
- Logs: `Home`, `End`, `PgUp`, `PgDn` move through the log once it has focus.

On macOS `Cmd` stands for `Ctrl`.

## Automatic groups and subscription updates

Enable **Automatically add available servers** in a proxy group's settings to
check every server in Proxies, including standalone nodes and all subscriptions.
The group keeps only servers that complete an HTTP(S) request through their VPN
connection. A listening TCP port or an ICMP response is not enough.

The first scan runs when enabled, then after catalog changes and at the group's
probe interval. The default for new groups is **300s**; saved intervals are
preserved. Turning the switch off keeps the current members and returns control
to manual editing. The group's existing Auto/least-ping mode chooses the server.
Subscription updates do not select a server or populate manual groups.

Enable **Automatically update subscription** for each subscription that should
refresh itself. Both minute fields are required:

- **Regular update interval:** 0 disables regular downloads; a positive whole
  number schedules unconditional updates.
- **Retry interval when all servers are unavailable:** at least 1 minute. If no
  saved candidate is reachable, refresh at this interval until one works.
  Failed or empty downloads retain the saved list. A failed subscription retries
  independently of healthy subscriptions in the same group.

Candidate health is checked after subscription/catalog changes and every 300
seconds for enabled subscription updates. Each automatic group uses its own
probe URL and interval. Timers start when enabled or the service starts.
Retries do not postpone unconditional updates. No background action starts a
manually stopped main core; temporary candidate probes can still run.

An empty automatic group has a blocking outbound. Traffic assigned to that group
is rejected instead of falling back to another group or direct access. Explicit
direct routes and other groups retain their own policy. On OpenWrt TPROXY and
Redirect, membership reloads retain existing interception when network settings
are unchanged and no custom hooks are installed. This is not a system-wide kill
switch for power loss, service termination, custom hooks, or TUN teardown.

Upgrading from the earlier Resilient recovery policy preserves accounts,
subscriptions, nodes and saved group members. Old first-server, auto-connect,
recovery and global subscription schedule flags no longer run. The new switches
default off: enable the desired group and subscription policies explicitly.
