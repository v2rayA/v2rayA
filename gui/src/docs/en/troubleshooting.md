# Troubleshooting

The service log is the first place to look: `/var/log/v2raya/v2raya.log` under the Linux service unit, `journalctl -u v2raya` for what the unit itself prints, the path given by `--log-file` otherwise. The **Logs** page shows the same stream; the **Log Level** setting raises the core's verbosity. The generated core configuration is `config.json` in the configuration directory.

## The core does not start

`CORE_START_FAILED` means `v2raya_core` could not be launched, exited right after starting, or did not open its API port within `--core-startup-timeout`; the lines before it in the log say why. Usual reasons:

- **Core version mismatch** — the banner at the top of the page. `v2raya_core` must be the same version as `v2raya`; an `xray` or `v2ray` binary in its place is not accepted. Install both from the same release.
- **`illegal domain rule: geosite:...`** — a rule names a category that is not in `geosite.dat`. See the routing section for which names exist.
- **address already in use** — an inbound port is taken. Change it under Address and Ports or stop the other program. Port 53 is different: when it is busy the DNS module skips its listener there and says so in the log, and on macOS the existing resolver keeps answering.
- **`v2raya_core executable not found`** — the core is not next to `v2raya` or in `PATH`; pass `--v2ray-bin`.

## Nothing connects

- The dashboard shows the group and its members' latency. `TIMEOUT` means the TCP connection to the server failed from this machine; try another node or check the subscription.
- Test one inbound directly: `curl -x socks5h://127.0.0.1:20170 https://example.com`. If that works and the transparent proxy does not, look at the transparent proxy setup, the DNS rules and the group the transparent proxy routes to before blaming the node.
- With `redirect` or `tproxy`, Docker or a firewall may have replaced the iptables rules; stop and start the core to reinstall them.
- On Windows and macOS, an application that asks a LAN resolver directly for DNS bypasses the TUN.

## Rule data

Once the core is found, a start without `geoip.dat` or `geosite.dat` downloads the missing file from GitHub and exits when the download fails. Copy the files into the rule-data directory by hand (see the parameters section), or install a package that ships them.

## After an upgrade

A BoltDB database from an older release is migrated on the first start; the old file stays as `bolt.db.bak` (`bolt.db.bak.1` and up when that name is taken), and the accounts have to be registered again. Stop the service and copy the configuration directory before upgrading, then the migration can be repeated from that copy.

## Reset the password

Stop the service, run `v2raya --reset-password` as the account the service uses with the same `--config` and `--lite` flags (on Windows the service's directory is under `%ProgramData%\SYSTEM`, not the elevated user's), start it again and register.
