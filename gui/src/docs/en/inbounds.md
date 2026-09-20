# Inbounds and sharing

An inbound is a port on which the core accepts connections from applications or other devices.

## Built-in inbounds

**Settings → Address and Ports** sets them; 0 closes a proxy inbound and lets the API port be chosen at random.

| Inbound           | Default | Routing                                                            |
| ----------------- | ------- | ------------------------------------------------------------------ |
| SOCKS5            | `20170` | everything through `proxy`                                         |
| HTTP              | `20171` | everything through `proxy`                                         |
| SOCKS5 with rules | off     | the rule port's mode                                               |
| HTTP with rules   | `20172` | the rule port's mode                                               |
| VMess with rules  | off     | a VMess inbound for other devices; the page shows its share link   |
| API               | random  | the core's own API, used by v2rayA for statistics and the balancer |

The proxy inbounds listen on `127.0.0.1`. **Port Sharing** (Settings → Proxy) makes them listen on all interfaces so phones and other machines on the LAN can use them; the API port stays on loopback. On a network you do not trust, add a custom inbound with a username and password for the other devices and close the built-in SOCKS and HTTP ports, which take no password, or block them in the firewall.

## Custom inbounds

**Settings → Custom Inbound Ports** adds SOCKS or HTTP inbounds with their own port, an optional username and password, and their own routing: either a fixed proxy group, or RoutingA rules written for that inbound, with the chosen group as the fallback (`default:` is ignored there). Two custom inbounds can therefore send traffic to two different groups, and an inbound whose rules end in `network(tcp, udp) -> direct` bypasses the proxy for every application pointed at it.

The tag is the inbound's name in the core configuration and in `inboundTag(...)` conditions.

## Docker and port mapping

In a container started with `--network=host` the inbounds and the transparent proxy act on the host. With bridge networking, publish port 2017 and the inbound ports you use, and turn on Port Sharing so they listen on the container's interfaces; the transparent proxy then only sees the container. v2rayA cannot tell whether a mapped port is free on the host.
