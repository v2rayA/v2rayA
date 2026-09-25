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

Run the current group/timer suite:

    python3 tests/openwrt/automation.py --disposable-vm --xray /path/to/host/xray --ssh-key /path/to/test/key --output /tmp/automation-results

See [AUTOMATION.md](AUTOMATION.md) for the direct-escape trap and assertions.
The superseded first-entry/recovery tests are available in Git history.

results.json contains completed assertions only. A nonzero exit means the suite
did not pass, even if some earlier assertions succeeded.

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
