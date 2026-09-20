# Transparent proxy

With the transparent proxy on, every application's traffic passes through the core without any application setting. **Settings → Transparent Proxy/System Proxy** turns it on and picks the splitting policy; the setting below it picks the implementation.

## Policies

| Policy | Traffic through the proxy |
| --- | --- |
| Do not split | everything |
| Proxy except CN sites | everything except Chinese domains and addresses (`geosite:cn`, `geoip:cn`) and private addresses |
| Proxy only GFWList | the domains on the GFWList; needs the list downloaded once (**Update GFWList**) |
| Same as the rule port | the rule port's mode, including RoutingA |

## Implementations

| Implementation | Platforms | Notes |
| --- | --- | --- |
| `redirect` | Linux | iptables/nftables `REDIRECT`; TCP only; needs local port 53 free for its DNS interception; works with Docker containers on the host |
| `tproxy` | Linux | iptables/nftables `TPROXY`; TCP and UDP; not for traffic from Docker containers |
| `tun` | Linux, Windows, macOS | the core opens a TUN device; TCP and UDP; excludes v2rayA and the core by itself |
| System proxy | Windows always; Linux and macOS in `--lite` mode | sets the desktop's proxy settings (GNOME and KDE on Linux); only applications that honour them are covered |

`redirect` and `tproxy` need root and `iptables` or `nftables`; `tun` needs `/dev/net/tun` on Linux and administrator rights on Windows and macOS.

**Excluded Interface Prefixes** keeps traffic of the named interfaces (Docker bridges, VPN tunnels) out of `redirect` and `tproxy`.

## TUN

The core creates the TUN device and assigns its address. With **Auto Route** on, v2rayA installs the routes and points the system resolver at the core; with it off, your own scripts do (see the hooks in the parameters section).

What never enters the TUN: the core's own connections, the `direct` outbound and the DNS module's upstream queries (by socket mark on Linux, by binding to the physical interface on Windows and macOS). v2rayA and the core are always excluded; **TUN Excluded Processes** excludes more by executable name, one per line. Routes more specific than the default (connected networks, static routes) bypass the TUN as well.

Plain DNS on port 53 that reaches the TUN is answered by the core's DNS module according to the DNS rules; encrypted DNS is not intercepted. On Windows the system resolver is pointed at the TUN gateway, on macOS at the core's listener on `127.0.0.1`, which needs port 53 free.

Known limitation: on Windows and macOS an application that queries a LAN resolver directly still bypasses the TUN.

## DNS

**Settings → DNS** holds the rules the core's DNS module follows: which upstream answers which domains, and through which outbound the query travels. The defaults send private names to the system resolver, `geosite:cn` to `223.5.5.5` directly, and everything else to `1.0.0.1` through the proxy. A rule with an empty domain list is the fallback.

## Sharing with the LAN

**Port Sharing** makes the SOCKS, HTTP and custom inbounds listen on all interfaces instead of `127.0.0.1`, so other devices can use this machine as their proxy. To route their traffic through the transparent proxy as well, turn on **IP Forward** and point their default gateway at this machine.
