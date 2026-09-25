# OpenWrt installation

This fork build targets the official OpenWrt 24.10.4 package combination:

- v2rayA 2.2.7.3, based on upstream tag `v2.2.7.3`.
- Xray 25.1.30 from the OpenWrt package feed.
- Package architecture `aarch64_cortex-a53`, used by the Cudy TR3000 installation this change was prepared for.

The custom package is named `v2raya_2.2.7.3-r4.failover3_aarch64_cortex-a53.ipk`.
It updates the v2rayA service; keep the existing `luci-app-v2raya` package.
This is a fork build, not an official OpenWrt release.

## Install on the router

Replace `ROUTER` with the router's address. Keep a copy of the existing settings on your computer:

```sh
ssh root@ROUTER 'tar -czf /tmp/v2raya-before-failover.tar.gz /etc/v2raya /etc/config/v2raya'
scp root@ROUTER:/tmp/v2raya-before-failover.tar.gz .
scp v2raya_2.2.7.3-r4.failover3_aarch64_cortex-a53.ipk root@ROUTER:/tmp/
ssh root@ROUTER 'opkg install /tmp/v2raya_2.2.7.3-r4.failover3_aarch64_cortex-a53.ipk && /etc/init.d/v2raya restart'
```

Use `scp -O` if the router does not provide an SFTP server. Before installation, check the package checksum against the supplied `SHA256SUMS` and check `opkg print-architecture` on the router. Do not force an incompatible architecture or insufficient-space installation on the router.

The package retains the official init script, UCI configuration, upgrade retention file, dependencies and package hooks. An existing modified UCI configuration is preserved by opkg. The service's reported application version is `2.2.7.3-failover.3`.

## Enable selection

In the v2rayA web interface, open the subscription's **Modify** dialog and choose **Connect to a working server** under **Server Selection**. Enable automatic subscription updates in **Setting** and set an interval. The existing interval is measured in whole hours. The subscription's **Update** button uses the same selection path and can be used to check immediately.

To stay at the top of the subscription, turn **Server Selection** off: its label becomes **Always use the first server**. Saving applies the change to this subscription’s selected connection. Refreshes follow the first position even after reordering; monitoring never falls back to another entry. If the first entry fails, traffic stays unavailable until it recovers or the subscription supplies a working first entry. **Auto-Connect** separately permits initial selection when nothing is selected.

In working-server mode, each refresh waits for real HTTP checks through every supported candidate and selects the successful node with the lowest measured latency. The default destination is `https://gstatic.com/generate_204`; it can be changed through the existing `proxy` outbound probe URL setting. The chosen URL must be reachable through the VPN nodes. HTTP redirects and error responses do not count as success.

For recovery between scheduled updates, enable **Recover failed connections automatically** in the subscription’s **Modify** dialog. This per-subscription switch defaults to off and works with Auto-Connect disabled. It checks only the active tunnel, waits for at least one minute of continuously failed checks, then refreshes and searches all candidates. If none works, it retries immediately once, then after 5, 10, 20 and at most 30 seconds between completed attempts. Manual disconnects are respected. Disabling monitoring cancels background recovery. When monitoring is off, failover runs only when the subscription refreshes. Failed downloads, empty lists and subscriptions with no reachable candidates return an error while retaining the previous configuration. Retaining a dead connection does not make it work; it prevents replacing the saved configuration with another unverified choice.

External-plugin nodes are excluded from the isolated checks. The active subscription retains ownership of the main proxy connection; another subscription does not take over a manually selected connection.

## Roll back

Use the supplied, unmodified official package or download the same version from the official OpenWrt feed:

```sh
scp v2raya_2.2.7.3-r1_aarch64_cortex-a53.ipk root@ROUTER:/tmp/
ssh root@ROUTER 'opkg --force-downgrade install /tmp/v2raya_2.2.7.3-r1_aarch64_cortex-a53.ipk && /etc/init.d/v2raya restart'
```

The fork does not introduce a new database schema. The saved archive is available if you also need to restore the earlier settings. Stop the service before restoring it.

## Build from source

Follow [SUBSCRIPTION_FAILOVER.md](../../SUBSCRIPTION_FAILOVER.md). `repack.py` accepts only a checksum-verified original v2rayA package for `aarch64_cortex-a53` and a Linux ARM64 executable. The checksum must come from the current official package index; a cached web search can contain a checksum from an earlier feed rebuild.
