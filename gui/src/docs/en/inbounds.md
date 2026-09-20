# Inbounds and sharing

An inbound is a port on which the core accepts connections from applications or other devices.

## Built-in inbounds

**Settings → Address and Ports** sets them; 0 closes one.

| Inbound | Default | Routing |
| --- | --- | --- |
| SOCKS5 | `20170` | as the transparent proxy setting, else everything through the proxy |
| HTTP | `20171` | same |
| SOCKS5 with rules | off | the rule port's mode |
| HTTP with rules | `20172` | the rule port's mode |
| VMess with rules | off | a VMess inbound for other devices; the page shows its share link |
| API | random | the core's own API, used by v2rayA for statistics and the balancer |

The inbounds listen on `127.0.0.1`. **Port Sharing** (Settings → Proxy) makes them listen on all interfaces so phones and other machines on the LAN can use them; give the SOCKS and HTTP inbounds a password through a custom inbound if the network is not yours.

## Custom inbounds

**Settings → Custom Inbound Ports** adds SOCKS or HTTP inbounds with their own port, an optional username and password, and their own routing: either a fixed proxy group, or a set of RoutingA rules written for that inbound. Two custom inbounds can therefore send traffic to two different groups, or one can bypass the proxy for a whole application that is pointed at it.

The tag is the inbound's name in the core configuration and in `inboundTag(...)` conditions.

## Docker and port mapping

In a container started with `--network=host` the inbounds are the host's. With bridge networking, publish port 2017 and the inbound ports you use, and turn on Port Sharing so they listen on the container's interfaces; the transparent proxy is not available there. v2rayA cannot tell whether a mapped port is free on the host.
