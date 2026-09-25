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

## Automatic subscription updates

Each subscription has its own update mode:

- **Disabled:** update only when requested manually.
- **On service start:** update once whenever v2rayA starts.
- **At an interval:** update on startup and then after the configured number of minutes.
- **At an interval with fail-safe recovery:** use the regular schedule and also check the saved servers at the failure interval. When none works, refresh at that interval until at least one becomes available.

Failed or empty downloads keep the saved server list. A failure retry does not postpone the regular schedule, and a slow pass never overlaps another pass. On Linux, recovery downloads use marked sockets so transparent proxying cannot route them back into a failed proxy. This bypass is not guaranteed for TUN mode on macOS or Windows. Candidate checks may start temporary core processes, but automatic updates do not start a manually stopped main core.

On upgrade, the previous global **update on start** mode is assigned to every existing subscription as **On service start**. The previous interval mode becomes **At an interval** with the same interval converted from hours to minutes. Fail-safe recovery is never enabled during migration; select it explicitly where needed.

## Automatic proxy-group membership

Enable **Automatically add available servers** in a proxy group's settings to manage that group's members from the entire **Proxies** catalog, including local servers and every subscription. The group is checked after catalog changes and at its probe interval; a newly enabled automatic group defaults to `300s`. Available servers are added and unavailable servers are removed. While enabled, the automatic worker owns the member list. Turning it off preserves the last result and restores manual editing.

If no server is available, only traffic assigned to that group is sent to a blocking outbound. Other groups and direct routes keep working, and transparent interception remains active, so the outage cannot silently bypass the proxy. On upgrade, the old per-subscription auto-select option enables automatic membership for `PROXY` when at least one subscription used it.
