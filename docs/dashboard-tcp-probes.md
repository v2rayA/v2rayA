# Dashboard TCP probes and transparent proxying

The dashboard TCP latency check opens a connection to each candidate server.
It keeps transparent interception installed throughout the check. A successful
TCP handshake measures reachability of the server port, not whether its proxy
protocol, credentials or internet access work; use the HTTP latency check for
an end-to-end proxy request.

On Linux, only the service-owned TCP and DNS sockets receive `SO_MARK=0x80`,
the bypass mark already used by the core's outbounds. TPROXY, Redirect and TUN
continue intercepting unrelated traffic.

On macOS and Windows with TUN enabled, the probe uses the same egress-interface
selection as the core. Its IPv4 and IPv6 TCP and DNS sockets are bound to that
interface (`IP_BOUND_IF`/`IPV6_BOUND_IF` on macOS,
`IP_UNICAST_IF`/`IPV6_UNICAST_IF` on Windows). Failure to find or bind the
interface fails the probe rather than retrying through TUN. The probe never
temporarily removes TUN routes or restores them afterward.

DNS uses the shared resolver policy: try the system resolver, then retry with
the configured direct upstreams (or the built-in public resolvers if none are
configured) on an error, an empty answer, or a loopback/unspecified answer.
Every DNS socket in either attempt follows the same mark or interface binding
as the TCP probe. Other platforms, and macOS/Windows without TUN, use a plain
TCP dialer.
