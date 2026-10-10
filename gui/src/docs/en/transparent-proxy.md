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
| `redirect`     | Linux                                                        | iptables/nftables `REDIRECT`; TCP only, plus DNS on port 53 redirected to the core's DNS module (port 52353) in the *service and interception* mode             |
| `tproxy`       | Linux                                                        | iptables/nftables `TPROXY`; TCP and UDP                                                                      |
| `tun`          | Linux, Windows, macOS                                        | the core opens a TUN device; TCP and UDP; excludes v2rayA and the core by itself                             |
| System proxy   | Windows; Linux and macOS when not running as root (`--lite`) | sets the desktop's proxy settings (GNOME and KDE on Linux); only applications that honour them are covered   |

`redirect` and `tproxy` need root and `iptables` or `nftables`; `tun` needs `/dev/net/tun` and `ip` on Linux and administrator rights on Windows and macOS.

**Excluded Interface Prefixes** keeps traffic arriving on the named interfaces (Docker bridges, VPN tunnels; `docker*`, `veth*`, `wg*`, `ppp*` by default) out of `redirect` and `tproxy`; their DNS is still intercepted in the *service and interception* mode.

`--redirect-respect-bound-device` (Linux) lets TCP sockets bound to a device with `SO_BINDTODEVICE` bypass `redirect`: NetworkManager's connectivity checks are such sockets, and a redirected check reports a limited connection and keeps applications offline. Off by default. When it is on, the service marks those sockets with `0x80` through cgroup BPF; that needs kernel 5.14 or later and cgroup v2, and `redirect` fails to start when they are missing.

## TUN

The core creates the TUN device and assigns its address. With **Auto Route** on, v2rayA installs the routes and, in the *service and interception* mode, points the system resolver at the core; with it off, the setup and teardown scripts under **Configure Route Script** do.

What never enters the TUN: the core's own connections, the `direct` outbound and the DNS module's upstream queries (by socket mark on Linux, by binding to the physical interface on Windows and macOS). v2rayA and the core are always excluded; **TUN Excluded Processes** excludes more by executable name, one per line, for connections whose owning process can be identified; in the *service and interception* mode their DNS still goes to the core's DNS module. Routes more specific than the default (connected networks, static routes) bypass the TUN as well.

In the *service and interception* mode, plain DNS on port 53 that reaches the TUN is answered by the core's DNS module according to the DNS rules; encrypted DNS is not intercepted. On Windows the system resolver is pointed at the TUN gateway, on macOS at the core's listener on `127.0.0.1`, which needs port 53 free.

In the other two modes the TUN builds no such relay and neither platform's system resolver is changed; a query that still arrives is routed directly out of the device. With **Auto Route** off and the route script in charge, the same holds: v2rayA installs no resolver setting, so the script owns DNS as well.

Known limitation: on Windows and macOS an application that queries a LAN resolver directly still bypasses the TUN.

## DNS

**Settings → DNS Settings** holds the mode and the rules the core's DNS module follows.

| Mode                    | DNS module | System queries                                                                                  |
| ----------------------- | ---------- | ----------------------------------------------------------------------------------------------- |
| Off                     | not running | untouched                                                                                          |
| Service only            | running    | untouched: the resolvers, the firewall DNS rules and the TUN relay are left alone               |
| Service and interception | running    | redirected to it: the resolver is repointed and port 53 is diverted                              |

The mode decides two separate things — whether the module runs, and whether the system's queries are sent to it — which the *service only* mode answers differently. `Off` and *service only* both let plain TCP/UDP 53 pass through `redirect`, `tproxy` and the system proxy, and send it straight out of `tun`; there is no need to leave interception for encrypted DNS, which is never sent to the module. Choose *service only* when another program already points DNS at v2rayA, and `Off` when nothing should answer DNS for you at all.

The rules say which upstream answers which domains, and whether the query goes out directly. The defaults send private names to `127.0.0.1:53` (`localhost`, which needs a resolver listening there), `geosite:cn` to `223.5.5.5` directly, and everything else to `1.0.0.1` through the proxy. An outbound of `direct` queries directly; any other value sends the query through the local SOCKS inbound, so it is routed like SOCKS traffic. The rule with an empty domain list answers every domain no other rule names. An upstream is an address (`8.8.8.8`, `dns.google`), `tcp://host`, `tls://host` for DNS over TLS or `https://host/dns-query` for DNS over HTTPS; DNS over QUIC is not supported and is refused on save.

`Off` hides the rules and keeps them: nothing would read them, and dropping them would make switching the mode back on a different configuration.

Saving validates the rules, mode and node DNS choice together in one request. They are stored in one transaction; if applying the configuration fails, the previous values are restored. Failed saves keep the edits in the dialog for retry.

**Node resolution DNS**, at the bottom of DNS Settings below the rules, selects the global resolver for proxy node server names. It applies to new node connections and TCP/HTTP latency tests in Service only and Service and interception modes, regardless of interception. Off mode disables the module and uses the existing system resolution path for nodes. It hides the node DNS control and retains its saved choice. Selecting an enabled mode shows the control again. IP node addresses need no DNS query. Website names, subscription servers and DNS upstream bootstrap keep their existing resolvers.

In these enabled modes, node connections and TCP/HTTP latency tests wait for the built-in DNS module to serve queries. They query its local listener; the module uses the selected upstream directly or returns a cached result. Node lookups do not use the system resolver. A startup timeout, cancellation or core exit stops the lookup. An upstream failure is reported as a DNS query error and does not trigger a fallback. When the core is stopped, latency tests start a temporary core with the selected DNS configuration and stop it after testing.

The dropdown lists IP-based UDP, TCP, DoT and DoH endpoints from DNS rules, the original system configuration, and the built-in fallback list. On Linux, system DNS comes from `/etc/resolv.conf` or its original backup while DNS hijacking is active. Categories show the source: **direct group**, **localhost DNS**, **proxy group**, **fallback DNS**. Every node DNS query goes directly, even for a proxy-group source. DoT/DoH certificates are verified. The options indicate valid current sources, not tested network reachability.

**auto** chooses the first endpoint in the first nonempty group: direct → localhost → fallback, after duplicate sources are removed. Its label shows the full URL, including protocol and port, just like an explicit endpoint; for example, `udp://223.5.5.5:53 (auto)`. A failed query reports a DNS error without switching endpoints. Candidates follow the rules currently being edited, so you can add or replace a DNS source and select it before saving. Refresh reloads sources. If an explicit source disappears, the dropdown shows auto and a warning below it includes the original address. The service revalidates the source on save. Off mode saves neither hidden rule edits nor hidden node DNS edits. DNS cache and prefetch still apply; changed results affect later connections, while existing connections keep their current IP.

When starting the core, if a saved explicit node DNS address is no longer among the current sources, the core uses **auto** and logs a warning with the original address and the automatic endpoint. The saved choice is preserved. Temporary cores for latency tests follow the same rule. If the source returns, the original explicit choice is used again on the next core start or dropdown refresh. Saving while the dropdown shows auto saves `auto`; cancel leaves the original choice unchanged. Off mode retains the unused choice. An unreachable upstream or a failed query does not trigger fallback. If auto is also unavailable, startup or saving reports an error; the service still rejects invalid choices submitted in settings.

## Sharing with the LAN

**Port Sharing** makes the SOCKS, HTTP and custom inbounds listen on all interfaces instead of `127.0.0.1`, so other devices can use this machine as their proxy; the API port stays on loopback. To route their traffic through the transparent proxy as well, use `redirect`, `tproxy` or `tun`, turn on **IP Forward**, keep the interface they arrive on out of the excluded prefixes, and point their default gateway at this machine; the system proxy modes cannot do this.
