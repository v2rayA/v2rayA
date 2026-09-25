# Subscription recovery on OpenWrt

These tests change settings, create accounts and subscriptions, stop/start the
service and kill the running proxy core. Use a **disposable VM with a fresh
v2rayA database for each script**, never a production router.

Validated target: OpenWrt 24.10.4 `armsr/armv8`, Linux 6.6.110, 256 MiB RAM,
matching ARM64 service/core binaries and the current LuCI package. The intended
router is a Cudy TR3000 (`aarch64_cortex-a53`); generic ARM64 virtualization
cannot verify its wireless drivers, hardware offload or flash upgrade.

Forward localhost TCP ports 22022 → guest SSH 22, 22017 → API 2017,
22171 → HTTP proxy 20171 and 22080 → LuCI 80. Configure a test SSH key and
install `curl`. The guest must reach loopback fixtures on the host through
QEMU's `10.0.2.2` gateway. Each script creates temporary VLESS servers using a
host-compatible Xray executable and sends real requests through the guest.
No real subscription or VPN credentials are used.

```sh
python3 tests/openwrt/policy.py --disposable-vm \
  --xray /path/to/host/xray --ssh-key /path/to/test/key --output /tmp/policy-results
# Stop the service, preserve the first test database and start with a fresh one.
python3 tests/openwrt/groups.py --disposable-vm \
  --xray /path/to/host/xray --ssh-key /path/to/test/key --output /tmp/group-results
```

The policy suite covers dead first entries, subscription reorder, an empty list,
all nodes down, unavailable subscription downloads, immediate recovery retries,
private monitoring with public proxy listeners disabled, unexpected core death,
manual stop and persisted settings. It uses the real one-minute failure window.
The group suite includes a closed port, wrong VLESS UUID, a blackhole, slow and
fast working servers; it verifies actual traffic after the fastest disappears,
all-dead preservation, recovery, restart and nftables TPROXY.

`results.json` contains completed assertions only. An exception or nonzero exit
means the suite did not pass, even if some earlier assertions succeeded.

## Recovery ownership

Automatic recovery operates when the `proxy` outbound belongs entirely to one
subscription. Existing group balancing is retained; first-entry mode restricts
that subscription to its first node. A group mixing subscriptions or standalone
nodes is not automatically taken over. Manual stop, disabled monitoring and
changed selections cancel recovery. The two switches default to off and survive
both SQLite schema upgrades and migration from the earlier Bolt database.

One worker checks the active route every ten seconds. After a continuous minute
of failures it refreshes the subscription and probes candidates with at most two
extra core processes. If no candidate works, it retries once immediately, then
waits 5/10/20/30 seconds between unsuccessful rounds. A failed fetch can use saved
candidates. Failed selection keeps the prior configuration; applying an invalid
core configuration rolls back both the database and the running configuration.

## OpenWrt defaults and upgrades

New OpenWrt settings do not exclude `br-*` from TPROXY: excluding the LAN bridge
also bypasses router-originated traffic routed through it. Explicit saved
interface exclusions remain unchanged. When upgrading a previously configured
2.5.x installation, inspect its excluded interfaces if transparent traffic is
still bypassed. The usual OpenWrt 2.2.x database does not have this new field.

The merged core's observer adapter uses one sample per configured interval.
Otherwise Xray burst's default ten samples turn a configured ten-second interval
into a hundred-second cycle, leaving stale healthy status after a node dies.
This is an adapter change in this repository; it does not patch XTLS sources.
