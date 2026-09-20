# Troubleshooting

The service log is the first place to look: `/var/log/v2raya/v2raya.log` under the service units, `journalctl -u v2raya` for what the unit itself prints, the path given by `--log-file` otherwise. The **Logs** page shows the same stream. `--log-level debug` adds the core's configuration and every request.

## The core does not start

`CORE_START_FAILED` means `v2raya_core` exited right after starting; the line before it in the log says why. Usual reasons:

- **Core version mismatch** — the banner at the top of the page. `v2raya_core` must be the same version as `v2raya`; an `xray` or `v2ray` binary in its place is not accepted. Install both from the same release.
- **`illegal domain rule: geosite:...`** — a rule names a category that is not in `geosite.dat`. See the routing section for which names exist.
- **address already in use** — an inbound port or port 53 (for `redirect` and the macOS TUN resolver) is taken. Change the port under Address and Ports or stop the other program.
- **`v2raya_core executable not found`** — the core is not next to `v2raya` or in `PATH`; pass `--v2ray-bin`.

## Nothing connects

- The dashboard shows the group and its members' latency. A member at `timeout` is not reachable from this machine; try another node or check the subscription.
- Test one inbound directly: `curl -x socks5h://127.0.0.1:20170 https://example.com`. If that works and the transparent proxy does not, the problem is in the transparent proxy setup, not the node.
- With `redirect` or `tproxy`, Docker or a firewall may have replaced the iptables rules; stop and start the core to reinstall them.
- On Windows and macOS, an application that asks a LAN resolver directly for DNS bypasses the TUN.

## Rule data

On the first start without `geoip.dat` and `geosite.dat` the service downloads them from GitHub and exits when the download fails. Copy the two files into `/usr/share/v2raya` by hand, or install a package that ships them.

## After an upgrade

A database older than 2.4 (BoltDB) is migrated on the first start; the old file stays as `bolt.db.bak`, and the accounts have to be registered again. Stop the service and copy the configuration directory before upgrading, then the migration can be repeated from that copy.

## Reset the password

Stop the service, run `v2raya --reset-password` with the same `--config` and `--lite` flags the service uses, start it again and register.
