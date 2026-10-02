# Automatic proxy-group membership: upgrade notes

Subscription updates refresh the catalog. A group's **Automatically add available
servers** setting owns its membership, using supported reachable nodes from the
whole catalog. Disabling it preserves the last membership for manual editing.

The one-time legacy `autoSelect` migration enables automatic `PROXY` membership
only when at least one subscription used that option and all existing `PROXY`
members belong to those subscriptions, or the group is empty. A standalone node
or a node from any other subscription keeps `PROXY` manual. Its current members,
probe interval and other settings are preserved; the log explains how to opt in
explicitly. The retired per-subscription flags are cleared in either case.
Later starts do not override a user's choice.

Deleting a subscription or standalone server works with automatic groups enabled.
Catalog entries and group references are removed in one transaction, including
clearing a group that loses its last member and renumbering surviving references.
Only a running main core is reloaded, once after the transaction commits.

An empty automatic `PROXY` deliberately remains the default proxy outbound and
blocks global and rule-port traffic that could previously fall through to another
group. Explicit routes to other groups and `direct` remain usable. External
plugin-managed nodes cannot be checked in isolation and must use manual groups.

The worker uses at most two temporary probe cores and performs network checks
outside the configuration lock. Applying membership and reloading the main core
still holds the lock to avoid committing stale configuration. Editing requests
may return `REQUEST_IN_PROGRESS` after waiting five seconds during a slow reload;
retry after it completes. Read-only requests remain available. A failed group
pass schedules its retry from completion, with a minimum delay of 30 seconds.

Membership reloads retain transparent interception only when the network is not
paused and the transparent-proxy settings support retention. Startup failure
removes retained rules to avoid trapping the router behind a dead core; this is
distinct from a running core blocking an empty group.
