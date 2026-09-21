# Transparent proxy

With the transparent proxy on, traffic reaches the core without any application setting; what is covered depends on the implementation. **Settings → Transparent Proxy/System Proxy** turns it on and picks the splitting policy; the setting below it picks the implementation.

## Policies

| Policy                | Traffic through the proxy                                                                                                                                                                                        |
| --------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Do not split          | everything                                                                                                                                                                                                       |
| Proxy except CN sites | everything except Chinese domains and addresses (`geosite:cn`, `geoip:cn`) and private addresses; sites outside China, Google and Hong Kong and Macao addresses are proxied even when a Chinese rule would match |
| Proxy only GFWList    | the domains on the GFWList (`gfw` and `greatfire` from Loyalsoldier's file) and Telegram's address ranges; run **Update GFWList** first, the mode is refused without the file                                    |
| Same as the rule port | the rule port's mode, including RoutingA                                                                                                                                                                         |

## Implementations

| Implementation | Platforms                                                    | Notes                                                                                                        |
| -------------- | ------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------ |
| `redirect`     | Linux                                                        | iptables/nftables `REDIRECT`; TCP only, plus DNS on port 53 redirected to the core's DNS module (port 52353) |
| `tproxy`       | Linux                                                        | iptables/nftables `TPROXY`; TCP and UDP                                                                      |
| `tun`          | Linux, Windows, macOS                                        | the core opens a TUN device; TCP and UDP; excludes v2rayA and the core by itself                             |
| System proxy   | Windows; Linux and macOS when not running as root (`--lite`) | sets the desktop's proxy settings (GNOME and KDE on Linux); only applications that honour them are covered   |

`redirect` and `tproxy` need root and `iptables` or `nftables`; `tun` needs `/dev/net/tun` and `ip` on Linux and administrator rights on Windows and macOS.

**Excluded Interface Prefixes** keeps traffic arriving on the named interfaces (Docker bridges, VPN tunnels; `docker*`, `veth*`, `wg*`, `ppp*` by default) out of `redirect` and `tproxy`; their DNS is still intercepted.

`--redirect-respect-bound-device` (Linux) lets TCP sockets bound to a device with `SO_BINDTODEVICE` bypass `redirect`: NetworkManager's connectivity checks are such sockets, and a redirected check reports a limited connection and keeps applications offline. Off by default. When it is on, the service marks those sockets with `0x80` through cgroup BPF; that needs kernel 5.14 or later and cgroup v2, and `redirect` fails to start when they are missing.

## TUN

The core creates the TUN device and assigns its address. With **Auto Route** on, v2rayA installs the routes and points the system resolver at the core; with it off, the setup and teardown scripts under **Configure Route Script** do.

What never enters the TUN: the core's own connections, the `direct` outbound and the DNS module's upstream queries (by socket mark on Linux, by binding to the physical interface on Windows and macOS). v2rayA and the core are always excluded; **TUN Excluded Processes** excludes more by executable name, one per line, for connections whose owning process can be identified; their DNS still goes to the core's DNS module. Routes more specific than the default (connected networks, static routes) bypass the TUN as well.

Plain DNS on port 53 that reaches the TUN is answered by the core's DNS module according to the DNS rules; encrypted DNS is not intercepted. On Windows the system resolver is pointed at the TUN gateway, on macOS at the core's listener on `127.0.0.1`, which needs port 53 free.

Known limitation: on Windows and macOS an application that queries a LAN resolver directly still bypasses the TUN.

## DNS

**Settings → DNS Settings** holds the rules the core's DNS module follows: which upstream answers which domains, and whether the query goes out directly. The defaults send private names to `127.0.0.1:53` (`localhost`, which needs a resolver listening there), `geosite:cn` to `223.5.5.5` directly, and everything else to `1.0.0.1` through the proxy. An outbound of `direct` queries directly; any other value sends the query through the local SOCKS inbound, so it is routed like SOCKS traffic. The rule with an empty domain list answers every domain no other rule names. An upstream is an address (`8.8.8.8`, `dns.google`), `tcp://host`, `tls://host` for DNS over TLS or `https://host/dns-query` for DNS over HTTPS; DNS over QUIC is not supported and is refused on save.

## Sharing with the LAN

**Port Sharing** makes the SOCKS, HTTP and custom inbounds listen on all interfaces instead of `127.0.0.1`, so other devices can use this machine as their proxy; the API port stays on loopback. To route their traffic through the transparent proxy as well, use `redirect`, `tproxy` or `tun`, turn on **IP Forward**, keep the interface they arrive on out of the excluded prefixes, and point their default gateway at this machine; the system proxy modes cannot do this.
