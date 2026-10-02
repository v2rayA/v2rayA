# Automatic-group review validation

The review corrections were checked on macOS ARM64 with Go 1.26.0 and in a
disposable OpenWrt 24.10.5 ARM64 VM (Linux 6.6.119, QEMU `virt` with HVF,
two vCPUs and 256 MiB RAM). The service was built from this branch with the
`2.5.7-resilient.4` version string to match the installed core, which is based on
xray-core 26.7.28. This is a VM result, not a measurement on router hardware.

## What the 200-node measurement means

The earlier 2.90-second result on OpenWrt 24.10.4 used 200 supported VLESS
catalog entries with different names pointing to **one local VLESS fixture**.
All 200 entries were reachable. The fixture forwarded each probe to an HTTP
server on the host. It was not a test of 200 separate Internet endpoints.

The repeated healthy test on 24.10.5 took **3.286 seconds**, started 200 probe
cores, and delivered 200 requests to the fixture. Peak sampled combined service
and probe-core RSS was **104,804 KiB**. Names are part of the node key, so these
records do not collapse into a single cached probe.

For the failure case, ten entries use the working fixture and 190 point to a TCP
server that accepts connections but never answers. Each probe has a five-second
deadline, with two workers. The timeout contribution alone is therefore at least
`190 × 5 / 2 = 475` seconds. The untraced run took **488.291 seconds**, with
200 core starts, ten HTTP successes and 190 silent TCP accepts. All 200 candidates
completed; the resulting group contained exactly the ten working nodes. Peak
sampled combined RSS was **115,268 KiB**.

The RSS sampler reads separate `/proc` entries every 20 ms. This is neither an
exact memory peak nor an atomic process snapshot. One sample in the untraced
timeout run counted three core entries during turnover. Use the process-lifetime
audit below for concurrency, rather than claiming that this sampler proves the
worker limit. Tracing itself adds overhead.

The repeated timeout run with process tracing took **481.829 seconds**. It
recorded exactly **200 core lifetimes, at most two overlapping**, with no probe
core left alive. All 200 candidates completed, zero were cancelled, and exactly
ten became members. The fixture recorded ten successful HTTP requests and 190
silent TCP accepts. Peak sampled service-plus-core RSS was **135,912 KiB**;
this excludes the tracer's own RSS. Differences between traced and untraced
runs also include host scheduling and load, so they are not an estimate of
tracer overhead.

| Scenario | Supported / reachable / timed out | Wall time | Probe core starts | Peak sampled service + cores RSS |
| --- | --- | --- | --- | --- |
| Local healthy fixture | 200 / 200 / 0 | 3.286 s | 200 | 104,804 KiB |
| Local fixture + silent TCP, traced | 200 / 10 / 190 | 481.829 s | 200 | 135,912 KiB |

The main proxy core was stopped during these scale measurements. They cover
isolated candidate scanning and membership application, not main-core reload
time or simultaneous user traffic. They do not predict WAN latency, CPU use or
memory consumption on a physical router.

## Cancellation while scanning

With the same slow catalog, disabling automatic membership after at least twelve
probe starts returned `SUCCESS` in **0.002 seconds**. The worker stopped its
temporary cores and did not apply a partial member list. The network scan runs
outside `ConfigurationMu`. Applying the result and reloading the main core still
holds that lock; a separate controller regression test verifies that reads stay
available and an editing request has a bounded wait with `REQUEST_IN_PROGRESS`.

## Reproducing the scale test

Use a disposable VM with a matching service/core pair already installed, SSH
forwarded on 25922, the API on 25917, and QEMU user networking (host address
`10.0.2.2`). The script temporarily replaces the test database and wraps the core
to count starts; its `finally` block restores both. Do not use a production
router or a database containing real subscriptions.

```sh
python3 docs/testing/automatic-group-scale.py \
  --xray /path/to/host/xray \
  --ssh-key /path/to/disposable-vm-key \
  --known-hosts /path/to/disposable-known-hosts \
  --output /tmp/group-healthy

# Install strace in the disposable guest first; process tracing needs root.
python3 docs/testing/automatic-group-scale.py \
  --xray /path/to/host/xray \
  --ssh-key /path/to/disposable-vm-key \
  --known-hosts /path/to/disposable-known-hosts \
  --output /tmp/group-timeouts --timeouts 190 --trace

python3 docs/testing/automatic-group-scale.py \
  --xray /path/to/host/xray \
  --ssh-key /path/to/disposable-vm-key \
  --known-hosts /path/to/disposable-known-hosts \
  --output /tmp/group-cancel --timeouts 190 --cancel-after-starts 12
```

Each complete run records the member count, HTTP successes, silent accepts,
process starts, wall time and RSS samples. `--trace` also saves the guest
`strace -f -tt -s 256 -e trace=process` output and checks the overlap between each
probe core's `execve` and process-exit events.

## Build and regression checks

The running OpenWrt service passed these API/database checks with real VLESS
fixtures:

- Only reachable members from two subscriptions and the standalone catalog are
  selected. Disabling automatic membership freezes the list; failed and recovered
  nodes are removed and restored on subsequent passes.
- Manual connect and disconnect requests without an `outbound` cannot bypass
  automatic `PROXY` ownership.
- Deleting a subscription removes its automatic members and renumbers surviving
  references. Deleting a standalone member and then the final subscription leaves
  no stale references. The main core remains running and the empty group is a
  blackhole.
- An empty group's request returns 503 without reaching the origin. An explicit
  direct route still reaches the origin. An injected failed replacement core
  removes retained nftables interception; router-originated traffic works directly
  and the matching core restarts after restoration.
- A real legacy database enables automatic membership when safe. A database
  containing a manual standalone `PROXY` member keeps its references, URL,
  60-second interval and manual mode. Its log gives the opt-in instruction once;
  a second startup does not repeat it. Previously migrated user choices (disabled
  group, 90-second interval) are preserved on a second startup.

- Service `go build ./...` and `go vet ./...` passed.
- Service `go test -skip '^TestBackupAndRestoreResolvSymlink$' ./...` passed on
  macOS. The omitted test expects the Linux resolver-symlink implementation.
  An initial concurrent full run also hit the timing-sensitive
  `TestStartRejectsExitDuringPostStart`; its isolated rerun and the subsequent
  full run passed.
- Race tests passed for `db`, `db/configure`, `server/service` and
  `server/controller`.
- The compiled `db` and `server/service` test suites passed natively as root
  inside the ARM64 OpenWrt VM. The automatic-group, retained-interception,
  paused-reload and resolver-hijacker tests in `kernel/v2ray` also passed there,
  including `TestBackupAndRestoreResolvSymlink`.
- Core `go build ./...` and `go vet ./...` passed.
- GUI lint, type checking, six-locale consistency, 217 tests in 45 files and the
  production build passed. The updated English group documentation was opened
  in the VM GUI and checked at 390 px width in light and dark themes.

The regression tests cover deletion from automatic groups, clearing the last
reference, renumbering across groups, transaction rollback, omitted-outbound
ownership checks, conservative legacy migration and idempotence, paused-network
reloads, and retry deadlines after slow or cancelled work.
