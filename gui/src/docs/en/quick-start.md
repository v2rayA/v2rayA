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

One probe core at a time. TCP ping orders candidates; URL and a complete 256 KiB sample verify at least 100 KiB/s. No suitable server means the group blocks traffic.

Subscriptions have their own page, **Subscriptions** (on a phone the docs move to the app bar's menu to make room for it). They keep their nodes together and can be updated by hand or on a schedule (**Settings → Automatically Update Subscriptions**). The mode setting next to it decides whether the update goes through the proxy.

## Groups

Nodes are used through proxy groups. The group `proxy` always exists; more are created and deleted from the group button in the app bar. The selected group's settings open from the cog there or from the **Proxy group** dashboard card. A node joins a group from its menu, or select several on the Proxies page and pick the group under **Add to proxy group**. **Add or remove nodes** changes the members, and the group settings set the probe address, interval and connection strategy.

Pings all servers, then checks speed in ascending TCP latency order. Stops at the first suitable server. Checks only the current server. On URL failure or speed below 100 KiB/s, selects a replacement by TCP latency and speed. Checks members sequentially and rotates new connections among servers passing URL and speed checks. Pings servers, shuffles reachable candidates and checks them one at a time until one passes the speed test.

One probe core at a time. TCP ping orders candidates; URL and a complete 256 KiB sample verify at least 100 KiB/s. No suitable server means the group blocks traffic.

Every strategy fails closed: if no member is available, traffic assigned to that group is sent to a blackhole instead of going directly. **Automatically add available servers** controls only the member list; the connection strategy is independent and also applies to manually managed groups.

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

Adds every server from Proxies and subscriptions, including unavailable servers. Membership updates do not ping, test speed or start a probe core.

One probe core at a time. TCP ping orders candidates; URL and a complete 256 KiB sample verify at least 100 KiB/s. No suitable server means the group blocks traffic.

- Pings all servers, then checks speed in ascending TCP latency order. Stops at the first suitable server.

- Checks current reachability at each interval. Three consecutive failures trigger a replacement ranked by TCP latency and a complete speed sample of at least 100 KiB/s. Low or unknown speed alone does not evict a reachable current server.

- Checks members sequentially and rotates new connections among servers passing URL and speed checks.

- Pings servers, shuffles reachable candidates and checks them one at a time until one passes the speed test.

On upgrade, legacy subscription auto-select enables automatic membership for `PROXY` only if at least one subscription used it and every current `PROXY` member belongs to those subscriptions (or the group is empty). If there are standalone members or members from other subscriptions, `PROXY` stays manual and its selections are preserved. The old flags are retired in both cases; the log explains how to enable **Automatically add all servers** explicitly. The migration runs once and never overrides a later choice.

Reordering or renaming nodes keeps connections intact. A real membership or connection change re-evaluates the selected strategy without changing its mode. During TPROXY/REDIRECT switches, interception stays installed, including when restarting fails.

A manual pin pauses automatic selection without changing the strategy or membership. Clear the pin on the Proxy group card to resume the same strategy.
