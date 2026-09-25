# Validation report

Validated on September 25, 2026, using the OpenWrt 24.10.4 release and a local Apple Silicon host.

## Outcome

The official v2rayA 2.2.7.3-r1 package reproduced the reported defect: automatic subscription refresh selected an unreachable first server and traffic failed while later servers remained available.

The final fork passed the complete OpenWrt VM suite with **256 MiB RAM**, including multiple consecutive failures while all six subscription entries stayed present. Every successful switch was checked against both the saved server identity and real traffic through a VLESS server. The earlier refresh-only build also ran these functional scenarios with 512 MiB RAM. Native application tests, real scheduler tests and targeted race checks passed.

## Server selection policy and additional edge cases

The `2.2.7.3-r4.failover3` build adds a per-subscription **Server Selection** switch in **Subscription → Modify**. On selects a working server; off follows the first position even when it is unavailable. The active subscription's policy applies on refresh independently of monitoring and initial Auto-Connect.

All fifteen additional OpenWrt records passed using real Xray/VLESS traffic and production outage timers. The final cleanup snapshot had 112,636 KiB available memory; this is not a peak-memory benchmark.

| Case | Expected and verified behavior |
| --- | --- |
| First entry dead, later entries healthy | First remains selected; manual fallback is rejected |
| Monitoring enabled in first-entry mode | Refreshes continue without selecting a later entry |
| Different dead first entry arrives during recovery | Its position is adopted; recovery does not reset to another minute |
| Working first entry slower than a later entry | First wins; TPROXY traffic confirms the selected identity |
| Empty update in first-entry mode | Existing subscription and connection remain intact |
| Subscription reordered | Selection follows the new first position |
| All entries dead when changing to working-server mode | Policy still saves; recovery finds a node after it returns |
| Change from first to working-server mode | A reachable later node replaces the unavailable first entry |
| Monitoring and Auto-Connect both disabled | The active subscription's selection policy still applies on refresh |
| Subscription host returns HTTP 503 | Recovery uses saved candidates and restores real traffic |
| Public HTTP/SOCKS proxy listeners disabled | Monitoring continues through its private listener |
| Local Xray killed unexpectedly | Monitoring restarts it and restores transparent traffic |
| Policy changed after a manual stop | Main core stays stopped in both modes |
| OpenWrt service restarted | Selection policy and monitoring setting persist |
| Cleanup | One main Xray remains, with no temporary probe configurations |

Browser testing confirmed the switch's location, saved values in both directions, selection of the first entry, and empty latency fields for entries that were deliberately not checked. Such entries are no longer falsely labelled unavailable. Targeted race tests cover incomplete probe results, stale metadata snapshots and the first-entry policy. The native regression suite passed its nine scenarios.

These checks exposed three refinements: retain the recovery state when a dead first entry changes; allow policy changes during a complete outage; distinguish untested entries from failed checks. The original ten monitoring scenarios below describe the preceding `r3.failover2` build; their evidence remains separately labelled in the archive.

## Environment

| Component | Version or configuration |
| --- | --- |
| OpenWrt | 24.10.4, r28959-29397011cc |
| Guest kernel | Linux 6.6.110, ARM64 |
| VM board | QEMU `virt`, `armsr/armv8`, two vCPUs |
| Final full-suite memory | 256 MiB, no swap |
| Official baseline package | v2raya 2.2.7.3-r1 |
| Candidate package | v2raya 2.2.7.3-r4.failover3 |
| Guest Xray | Official OpenWrt xray-core 25.1.30-r1 |
| Host Xray fixtures | Xray 25.1.30, macOS ARM64 |
| LuCI package installed | luci-app-v2raya 26.259.57575~0aa55c7 |
| Compiler | Go 1.23.12, static Linux ARM64 build |
| Production service mode | Normal root service through the official OpenWrt init script; Lite disabled |

## Traffic tests

The six-entry subscription contained a dead first endpoint, wrong VLESS credentials, an authenticated blackhole, and three working VLESS endpoints named `slow`, `backup` and `fast`. Each working endpoint used a separate real Xray server and returned a distinct marker through a deterministic HTTP backend. Tests did not use a provider subscription or real user credentials.

| Scenario | Observed result |
| --- | --- |
| Initial manual connection | Real traffic passed through `slow` |
| Refresh with dead first node | Selected `fast`, the sixth entry; traffic returned `fast` |
| Kill `fast`, retain all entries | Selected `backup`; traffic returned `backup` |
| Kill `backup`, retain all entries | Selected `slow`; traffic returned `slow` |
| Kill all working nodes | Refresh failed, previous state retained, traffic failed as expected |
| Restore a later node | Selected restored `fast`, with failed entries still present |
| Reorder nodes while probes run | Active traffic kept working; final identity still matched `fast` |
| Concurrent mutation | Rejected while refresh was running |
| Subscription HTTP 503 / empty body | Previous subscription and working connection retained |
| nftables/TPROXY | Unproxied HTTP requests originating on the guest followed `fast` to `backup` after failure |
| OpenWrt service restart | Startup refresh selected recovered `fast`; HTTP proxy and TPROXY traffic both returned `fast` |
| Cleanup | No temporary probe configuration files; exactly one managed Xray remained |

The full 256 MiB run produced 12 passing records including environment and cleanup checks. During the sampled two-probe interval, `MemAvailable` was 105,556 KiB (about 103 MiB). After completion it was 114,652 KiB. These are snapshots, not a measured peak or a long-duration memory benchmark. Kernel logs contained no out-of-memory kills during the run. Xray's large virtual address-space value in `ps` is not its resident memory usage.

## Optional monitoring validation

The monitoring suite ran the production timers in the same 256 MiB OpenWrt VM with the ordinary update schedule disabled and Auto Select off. Ten records passed:

| Scenario | Observed result |
| --- | --- |
| Healthy monitoring | Real probe traffic; no subscription downloads; exactly one active Xray; private listener bound to loopback |
| Short outage | No refresh or switch; successful checks reset the failure window |
| Sustained outage | No refresh before sixty seconds; automatic recovery after 69.4 seconds; actual VLESS and TPROXY traffic used `backup` |
| All candidates dead | Failed manual refresh woke recovery; three subscription requests covered manual refresh, recovery and its immediate retry |
| Node restored | Recovery found `fast` without the ordinary scheduler |
| Monitoring disabled | No health traffic or refreshes during the observation window |
| Blocked subscription download | GUI-equivalent switch mutation completed within four seconds and cancelled recovery; disabled state remained intact |
| Manual stop | No autonomous restart or refresh; zero managed Xray processes |
| Service restart | Monitoring flag persisted and real traffic worked |
| Cleanup | One active core, no temporary probe files; 101,796 KiB available memory in the final snapshot |

The GUI was rebuilt from branch source. Browser testing verified the switch defaults to off and persists both enabled and disabled values after Save and Apply and reopening the dialog. Targeted race tests also cover continuous failure timing, bounded retry delays, persistence, manual-stop ownership and cancellation without holding the configuration lock during download.

## Findings used to refine the fork

1. **Selection was positional.** The baseline configured the first Xray node without checking whether it passed traffic. Selection now waits for every supported candidate's actual HTTP probe and uses the fastest successful result.
2. **Subscription download retry could lose its timeout.** The initial code copied an HTTP client without using the copy, and the shared helper retried with an unbounded default client. Both download attempts now have a timeout, and shared clients remain unchanged.
3. **Service restart could leave the old core on its ports.** The VM caught a case where the saved selection changed but traffic still used the previous core. Context cancellation alone did not synchronously terminate the child. Shutdown now terminates and reaps the managed core before returning and completes `exec.Cmd.Wait` cleanup. The repeated VM restart test passed after this change.

4. **Background recovery must not block user control.** Downloads and candidate probes run outside the configuration lock; user mutations cancel them, and the final commit validates the original subscription and connection snapshot. The blocked-download VM test confirmed prompt cancellation.

Fixture issues were corrected separately: the guest HTTP proxy needed port sharing for the host-forwarded test; `touch` is deliberately busy during a manual update in this release; and BusyBox process matching needed the full Xray command line. The monitoring regression setup also needed to wait for the stopped procd service to exit before moving its database and starting the next test. These fixture corrections were not counted as product fixes.

## Other validation

- Native application suite: nine scenarios passed with the real final application and Xray, including injected core-start failure, database rollback and restoration of the previous traffic path.
- `go test -race ./core/v2ray ./server/service .`: passed, including real Xray HTTP status/timeout checks, connection ownership, reference remapping, failed downloads and synchronous process shutdown.
- `go vet ./core/v2ray ./server/service ./server/router ./db/configure .`: passed.
- The actual subscription ticker was fired by the integration test inside OpenWrt with the real guest Xray. This uses a shortened test-only tick, not a claim that a full one-hour production interval elapsed.
- Final IPK contents were checked against the checksum-verified official package. The UCI file, init script, upgrade retention list, conffiles, installation/removal hooks, dependencies and executable mode were preserved. The installed guest executable's SHA-256 matched the built executable.
- `git diff --check`: passed.

The whole-repository test and vet commands are **not green on the pristine upstream tag on this Mac**. Baseline comparison reproduced the existing `common` test import cycle, platform-specific pcap code errors, legacy external-download/hardcoded-path tests, and existing vet findings. These are outside the changed paths; the targeted results above must not be read as a passing whole-repository suite.

## Scope and limits

Working-server mode checks candidates at subscription refresh. Optional monitoring checks the active tunnel between updates and starts recovery after one minute of failed observations. It is off by default. Recovery retries immediately once, then uses bounded delays; failed/empty downloads fall back to saved candidates. Preserved state is not a promise of working traffic. The additional policy suite verifies the cached fallback after an HTTP 503; prolonged provider outages and long-running operation remain outside the recorded run.

The VM validates the same OpenWrt release and application binary on a generic ARM64 board. It does not validate the Cudy TR3000's SoC, Wi-Fi, flash layout, physical LAN forwarding, real provider credentials, TLS/REALITY variants, external plugins or long-running production load. TPROXY was checked with traffic originating on the guest, not a separate LAN client. External-plugin nodes are excluded from isolated selection probes.

Reproduction instructions: [OPENWRT_VM.md](OPENWRT_VM.md). Installation and rollback: [OpenWrt README](../install/openwrt/README.md).
