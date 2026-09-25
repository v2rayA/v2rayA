# Subscription failover for the OpenWrt 2.2.7.3 package

This branch targets v2rayA 2.2.7.3 with Xray 25.1.30, the versions in the official OpenWrt 24.10.4 package feed. `luci-app-v2raya` opens and controls this service; the selection logic belongs to v2rayA itself.

This draft preserves the 2.2.7.3 baseline and its CI files. The current upstream architecture has changed substantially; this is a tested compatibility implementation that still needs adaptation and validation against current main. Validation below applies to the compatibility implementation, not to a merge with main.

## Behavior

In the subscription’s **Modify** dialog, choose **Connect to a working server** under **Server Selection**. **Auto-Connect** controls initial selection when nothing is selected. Set automatic subscription updates and their interval in **Setting**. Both the subscription's **Update** button and the scheduled update use the same selection code.

1. Download and parse the subscription without changing the current connection.
2. Check every candidate by sending an HTTP request through an isolated Xray process. Check two candidates concurrently to bound memory use on routers. Each request has a five-second timeout; core startup uses the configured core startup timeout.
3. Wait for every result, then choose the reachable server with the lowest measured latency. HTTP 2xx is success; redirects, errors and timeouts are failures. Equal latencies keep list order.
4. Save the updated subscription and all remapped connection references in a single database transaction. Restart the active core only when the chosen server changes.
5. If downloading fails, the new list is empty, or all candidates fail, keep the previous subscription and connection. Return an error to manual updates and log errors from scheduled updates. If applying a selection fails, restore the previous database state and restart the previous core; report restoration failures too.

The active subscription owns the `proxy` outbound. Other subscriptions cannot take it over. If nothing is selected, the first subscription with Auto-Connect enabled owns the initial selection. A manually selected standalone server is retained. Other outbounds retain their existing endpoints, including nodes removed from the refreshed subscription. The selection policy applies whenever the active subscription refreshes, independently of Auto-Connect and monitoring. The V2Ray multi-server selection path is unchanged.

Subscription downloads retain their existing direct retry after a transport failure, with a timeout on both attempts. A shared HTTP client is never modified. Core shutdown now waits for the child process to finish, so a service restart cannot leave the previous core holding the proxy ports.

The probe URL is the existing `proxy` outbound's `probeURL`, defaulting to `https://gstatic.com/generate_204`. A reachable TCP port alone does not qualify a server. The probe listener binds to loopback; the generated outbound retains v2rayA's transparent-proxy bypass mark when transparent proxying is enabled. The active core is left running during probes. Concurrent API mutations are rejected as busy while the scheduler holds the shared configuration lock.

Working-server selection implements failover **at subscription refresh**. Optional monitoring also detects failures between updates, as described below. External-plugin nodes are excluded from isolated probing. The interval still uses the existing whole-hour setting. Native protocols use the existing v2rayA configuration generator.

## Server selection policy

The **Server Selection** switch is on by default: **Connect to a working server** checks candidates and selects the fastest successful node. Turning it off selects **Always use the first server**. Saving a changed policy applies it to this subscription's selected connection, while preserving a manual service stop. It never takes over another subscription or a standalone server.

First-entry mode follows position in the refreshed list, not the identity of the former first server. It probes only the first entry and never falls back to later entries. An unavailable first entry may therefore leave traffic unavailable by design. Monitoring can refresh and retry that entry, and can adopt a different first entry published by the provider; it cannot override the policy. Empty or invalid lists preserve saved state. Manual attempts to select a later node are rejected until the policy is changed. A dead replacement first entry does not restart the one-minute outage window during an ongoing recovery.

The existing **Auto-Connect** option controls which subscription may initially select a server when nothing is selected. It does not disable the active subscription's selection policy. Monitoring remains a separate opt-in switch.

## Optional connection monitoring

In the subscription's **Modify** dialog, enable **Recover failed connections automatically**. This switch defaults to off and is saved per subscription. It works independently of Auto-Connect and the hourly update schedule. Monitoring only follows the currently connected subscription; it does not connect an idle subscription or undo a manual disconnect.

One worker checks the active tunnel every ten seconds using the configured HTTP probe URL and a five-second request timeout. It uses a private loopback HTTP inbound in the existing Xray process, even when public proxy ports are disabled. Healthy monitoring creates no additional core processes and does not download the subscription or scan every node. A successful check clears the outage timer. After at least one minute of continuously failed observations, recovery downloads the subscription and probes all candidates, with at most two temporary cores at a time. Sampling and request timeouts mean recovery begins after the one-minute threshold rather than exactly sixty seconds after the physical outage.

If no candidate works, recovery refreshes again immediately once, then waits 5, 10, 20 and at most 30 seconds between subsequent attempts. These waits follow completed attempts; network and probe time is additional. It never waits for the ordinary hourly schedule. This bounded retry prevents a tight loop against an unavailable provider. A manual or scheduled refresh that finds no working candidate also wakes recovery immediately. A failed or empty download leaves the saved list intact and recovery checks its cached candidates.

Background download and probing release the configuration lock. Changing settings, selecting a server, manually stopping, or disabling monitoring cancels the job. The final commit rechecks the selected connection, running core and subscription snapshot so stale results cannot overwrite a user change. A shared scan lock prevents overlapping candidate scans. Applying a recovered candidate restarts the core even if the endpoint is unchanged, allowing recovery of a failed local core. Shutdown cancels and joins the single monitoring worker.

## Root cause in the base release

`SelectServersFromSubscription` calls `Connect` on each node in list order, then immediately breaks for Xray. `Connect` configures the core but does not prove that the remote node can pass traffic. The scheduler also invokes selection before downloading the subscription and again afterwards. With Xray, even the pre-update "disconnect" call takes the selection path. Manual refresh previously did no health-based selection at all.

The OpenWrt VM exposed an additional lifecycle failure: stopping the service cancelled the core's context but did not wait for the core to terminate. During an OpenWrt service restart, the new v2rayA process could find the old Xray still occupying the ports. Subsequent selection could update the saved identity without changing traffic. The fix explicitly terminates and reaps the managed core before shutdown or replacement, uses `exec.Cmd.Wait` to complete command cleanup, and waits for managed plugin processes too. This finding is why validation checks both the selected identity and the traffic path after restart.

## Reproduce the validation

Use Go 1.23.12 and Yarn 1.22.22. Build the GUI from this branch so the monitoring switch is included (do not use the upstream prebuilt GUI):

```sh
cd gui
yarn install --frozen-lockfile --ignore-engines
OUTPUT_DIR=../service/server/router/web yarn build
cd ..
```

Xray is a separate executable. After the GUI build completes, from `service/`:

```sh
CGO_ENABLED=0 go build -trimpath -o /tmp/v2raya .
V2RAYA_V2RAY_BIN=/absolute/path/to/xray go test -race ./core/v2ray ./server/service .
go vet ./core/v2ray ./server/service ./server/router ./db/configure .
```

From the repository root, run the real application and loopback fixtures:

```sh
python3 tests/subscription_e2e.py --binary /tmp/v2raya \
  --xray /absolute/path/to/xray --output /tmp/v2raya-validation
```

The suite creates a disposable local account and subscription, starts the application and Xray, passes real HTTP traffic, exercises refreshes, blocks a candidate, changes server order, removes the active node, tests concurrent mutations, injects a core startup failure, and tests automatic refresh after application restart. `--keep-running` keeps a demo open after successful checks. The Go integration test separately fires the actual subscription ticker.

For real OpenWrt service, VLESS and nftables/TPROXY checks, follow [tests/OPENWRT_VM.md](tests/OPENWRT_VM.md). For installation and rollback on the router, follow [install/openwrt/README.md](install/openwrt/README.md).

## Build the router package

```sh
cd service
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath \
  -ldflags '-s -w -X github.com/v2rayA/v2rayA/conf.Version=2.2.7.3-failover.3' \
  -o /tmp/v2raya-linux-arm64 .
cd ..
python3 install/openwrt/repack.py --base /path/to/official-v2raya.ipk \
  --sha256 CHECKSUM_FROM_CURRENT_OFFICIAL_PACKAGES_INDEX \
  --binary /tmp/v2raya-linux-arm64 \
  --output /tmp/v2raya_2.2.7.3-r4.failover3_aarch64_cortex-a53.ipk
```

The repacker retains the official package's init script, configuration file, upgrade retention list, conffiles, dependencies and installation/removal scripts. It replaces the executable and identifies the result as a local fork build. This is an unsigned custom package, not an official OpenWrt release.

The Linux ARM64 executable is static. The OpenWrt VM uses the same release and executable but a generic ARM64 board, not Cudy hardware. Physical LAN forwarding, Wi-Fi, provider-specific credentials and long-running operation on the Cudy TR3000 remain outside the VM test's coverage.
