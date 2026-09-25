# Group and subscription automation validation

Run automation.py only against the documented disposable VM. It uses synthetic
subscriptions and loopback VLESS fixtures, never a real subscription token.

The independent DNAT trap is first verified with the proxy stopped. Continuous
router-originated requests then run during membership changes. Any direct
response or trap request fails the test, including brief reload windows.

The suite covers all catalog sources, authenticated reachability, all-dead
blackhole membership, recovery, disabling automatic membership, required minute
fields, zero regular interval, real one-minute retries, retry cancellation after
recovery, and manual stop. The earlier policy.py and groups.py in Git history
describe the superseded recovery design; use automation.py for this release.

An empty automatic group has a blackhole outbound. Other groups and explicit
direct routes remain usable. Membership-only reloads retain installed Linux
TPROXY/Redirect rules when network settings are unchanged and no custom hooks
are configured. Candidate probes and subscription downloads use marked sockets
so recovery does not depend on the blocked group. This is not a system-wide
kill switch: manual stop, service/core crashes, TUN, custom hooks and network
configuration changes are outside the retained-firewall guarantee. The VM
traffic assertion covers router-originated IPv4 TCP under nftables TPROXY;
it does not prove every IPv6, UDP or hardware-offload path.

Results and VM logs must accompany the release report. A completed build or a
healthy first node alone is not a successful integration test.
