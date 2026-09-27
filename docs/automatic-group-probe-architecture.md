# Group membership and bounded probing

Membership copies the complete catalog, including unavailable nodes. It performs no network checks and never launches a probe core. Explicit membership refresh uses the saved policy and does not force a selection pass.

All isolated checks share one cancellable process slot. The slot remains held until the child process exits and is reaped. At most one isolated core runs alongside the traffic core, including manual HTTP latency tests and subscription recovery. Candidates are checked sequentially; no group gets its own parallel process pool.

Least latency first performs bounded direct TCP pings (up to eight sockets, with interception bypass), then tests candidates in ascending latency order. UDP-only transports with no TCP listener are checked through their real protocol after TCP-reachable candidates. A full URL check and 256 KiB speed sample must succeed at >=100 KiB/s. The first passing candidate wins; later candidates are not started. At each configured check interval, Keep current tests only its current candidate and its speed, then uses the same search on failure or low speed. Random shuffles TCP-reachable candidates before sequential speed checks. Round robin checks its members sequentially and uses only passing members. No successful speed measurement means no eligibility; if none qualify, the group fails closed without deleting its members.

Subscription fail-safe recovery tests one candidate at a time and stops on the first successful URL check. A speed-test failure alone does not imply every subscription endpoint is unavailable.

The firstavailable strategy is removed from the UI and new API settings. Existing persisted values normalize to leastping; explicit fixed selections remain fixed. Membership and selected-route updates are separate: catalog changes cannot discard unhealthy nodes, and switching the selected route does not rewrite membership.

Validation must cover cancellation, concurrent requests sharing the process slot, retained dead members, early stop, low/unknown speed, current-server failover, legacy settings, and one OpenWrt VM with process-count sampling during concurrent operations. The one-process bound reduces peak memory; it does not prove that unrelated core memory growth is resolved.
