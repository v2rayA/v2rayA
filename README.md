<div align="center">

<img src="gui/public/static/v2raya-icon.svg" width="96" alt="v2rayA">

# v2rayA

**A web client for its own Xray-based core, with transparent proxy on Linux, Windows and macOS.**

English · [简体中文](README_zh.md) · [Русский](README_ru.md)

[Requirements](#requirements) • [Install](#install) • [First run](#first-run) • [Transparent proxy](#transparent-proxy) • [Data and upgrades](#data-and-upgrades) • [Support](#support)

</div>

v2rayA runs as a service and is used from a browser, on the machine itself or on a router or NAS. It imports subscriptions and share links for VMess, VLESS, Shadowsocks, Trojan, Hysteria2, TUIC, [Juicity](https://github.com/juicity), AnyTLS, WireGuard, SOCKS5 and HTTP(S), groups nodes and selects the member with the lowest measured latency or a pinned one, and splits traffic with rules written in RoutingA. ShadowsocksR is not supported.

## Requirements

| Component | Requirement |
| --- | --- |
| Core | `v2raya_core` of the **same version** as `v2raya`, next to it or in `PATH`. The packages and installers below ship both; a hand-installed pair must match. |
| Rule data | `geoip.dat` and `geosite.dat` in `/usr/share/v2raya` or `/usr/local/share/v2raya`. The packages ship them; a hand install downloads them from GitHub on the first start and exits if that fails. |
| Linux transparent proxy | root; `iptables` or `nftables` for `redirect` and `tproxy`; `/dev/net/tun` and `ip` from iproute2 for `tun` |
| Windows | Windows 10 build 14393 or later; administrator for `tun`; started without elevation it runs in lite mode and can still set the current user's system proxy |
| macOS | root for `tun`; started as a user it runs in lite mode with the system proxy instead |
| Browser | a current Chrome, Edge, Firefox or Safari |

## Install

The packages below install `v2raya` and `v2raya_core` together with a service definition. Enable the service with the command shown for the platform.

<details>
<summary><strong>Debian, Ubuntu and other APT distributions</strong></summary>

The packages come from the [Dae Universe repository](https://github.com/daeuniverse/repo-for-linux).

```sh
sudo apt update
sudo apt install curl
```

APT 3.0 or later:

```sh
sudo curl -fsSL -o /etc/apt/sources.list.d/daeuniverse.sources https://daeuniverse.pages.dev/daeuniverse.sources
```

APT older than 3.0:

```sh
sudo curl -fsSL -o /etc/apt/sources.list.d/daeuniverse.list https://daeuniverse.pages.dev/daeuniverse.list
```

Then the key and the package:

```sh
sudo curl -fsSL -o /usr/share/keyrings/daeuniverse-archive-goose.gpg https://daeuniverse.pages.dev/daeuniverse-archive-goose.gpg
sudo apt update
sudo apt install v2raya
sudo systemctl enable --now v2raya
```

Change the unit with `sudo systemctl edit --full v2raya.service`; an edit of the installed file is overwritten by the next upgrade.

</details>

<details>
<summary><strong>Fedora, RHEL, openSUSE and other RPM distributions</strong></summary>

Fedora, RHEL and derivatives:

```sh
sudo curl -fsSL -o /etc/yum.repos.d/daeuniverse.repo https://daeuniverse.pages.dev/daeuniverse.repo
sudo dnf install v2raya
```

openSUSE:

```sh
sudo curl -fsSL -o /etc/zypp/repos.d/daeuniverse.repo https://daeuniverse.pages.dev/daeuniverse.repo
sudo zypper install v2raya
```

Then:

```sh
sudo systemctl enable --now v2raya
```

</details>

<details>
<summary><strong>Arch Linux</strong></summary>

`v2raya` builds from source, `v2raya-bin` uses the release binaries; both are in the AUR. The release page also carries `installer_archlinux_<arch>_<version>.pkg.tar.zst` for `pacman -U`.

```sh
paru -S v2raya-bin
sudo systemctl enable --now v2raya
```

</details>

<details>
<summary><strong>Gentoo, Alpine and other OpenRC systems</strong></summary>

On Gentoo, install from the [gentoo-zh](https://github.com/gentoo-zh/overlay) overlay, then enable the service:

```sh
sudo emerge net-proxy/v2rayA
sudo rc-update add v2raya default
sudo rc-service v2raya start
```

Elsewhere, download the two binaries for the architecture from the [releases](https://github.com/v2rayA/v2rayA/releases) and use the OpenRC files from [`install/universal/`](install/universal/). Run this from a checkout of the repository, with `VERSION` set to the release number without the leading `v`; the example is for x64:

```sh
sudo install -m755 "v2raya_linux_x64_${VERSION}" /usr/bin/v2raya
sudo install -m755 "v2raya_core_linux_x64_${VERSION}" /usr/bin/v2raya_core
sudo install -m755 install/universal/v2raya.initd /etc/init.d/v2raya
sudo install -m644 install/universal/v2raya.confd /etc/conf.d/v2raya
sudo rc-update add v2raya default
sudo rc-service v2raya start
```

</details>

<details>
<summary><strong>Docker</strong></summary>

The image is `ghcr.io/v2raya/v2raya` (also `mzz2017/v2raya` on Docker Hub). This example is for a Linux host and gives the container the host network and privileges, which transparent proxy needs:

```sh
docker run -d --restart=always --privileged --network=host --name v2raya \
  -e V2RAYA_LOG_FILE=/tmp/v2raya.log \
  -v /lib/modules:/lib/modules:ro \
  -v /etc/resolv.conf:/etc/resolv.conf \
  -v /etc/v2raya:/etc/v2raya \
  ghcr.io/v2raya/v2raya
```

With bridge networking instead, publish port 2017 and the inbound ports you use, and turn on port sharing in the settings; the inbounds listen on the container's loopback otherwise. Transparent proxy is not available in that setup.

</details>

<details>
<summary><strong>Windows</strong></summary>

`installer_windows_inno_x64_<version>.exe` (or `arm64`) from the [releases](https://github.com/v2rayA/v2rayA/releases) installs both binaries and registers a service. The same installer is behind `winget install --id v2rayA.v2rayA` and the [scoop bucket](https://github.com/v2rayA/v2raya-scoop) (`scoop bucket add v2raya https://github.com/v2rayA/v2raya-scoop && scoop install v2raya-np`).

</details>

<details>
<summary><strong>macOS</strong></summary>

The [tap](https://github.com/v2rayA/homebrew-v2raya) installs both binaries. Started as root the service has the `tun` transparent proxy; started as you it runs in lite mode with the system proxy:

```sh
brew tap v2raya/v2raya
brew install v2raya/v2raya/v2raya
sudo brew services start v2raya   # root: tun
brew services start v2raya        # you: system proxy
```

A service started with `sudo` is stopped with `sudo brew services stop v2raya`, and Homebrew warns that the upgrade or uninstall of a root-started formula needs `sudo rm` of the paths it names.

</details>

<details>
<summary><strong>OpenWrt</strong></summary>

The [v2raya-openwrt](https://github.com/v2rayA/v2raya-openwrt) feed and the official packages feed currently package 2.2.7.x with a separate `xray-core`, not the release described here. Until they are updated, install the `mips32`, `mips32le`, `arm64` or `x64` release binaries by hand, as in the next section, with the init script of your choice.

</details>

<details>
<summary><strong>Other Linux, without a package</strong></summary>

Download `v2raya_linux_<arch>_<version>` and `v2raya_core_linux_<arch>_<version>` from the [releases](https://github.com/v2rayA/v2rayA/releases) (`x86`, `x64`, `arm64`, `armv7`, `riscv64`, `loongarch64`, `mips32`, `mips32le`, `mips64`, `mips64le`), check them against the `.sha256.txt` next to each, and install both into `/usr/bin`. With systemd:

```sh
sudo install -m644 install/universal/v2raya.service /etc/systemd/system/v2raya.service
sudo systemctl daemon-reload
sudo systemctl enable --now v2raya
```

With OpenRC, use the files from the section above. There is also a [snap](https://snapcraft.io/v2raya) maintained separately.

</details>

## First run

The service listens on `0.0.0.0:2017`, so on a router or NAS the interface is at `http://<device-address>:2017`. The first account registered becomes the administrator and registration needs no login, so restrict access before registering on a shared network, or start with `--address 127.0.0.1:2017` for local use only.

Open the interface and create the administrator. The tutorial then walks through importing a subscription or a share link, adding the nodes to a group and choosing the routing rules; start the core with Start on the dashboard or the status button at the top of the page.

Once the core runs, point an application at SOCKS5 `127.0.0.1:20170` or HTTP `127.0.0.1:20171` (`20172` applies the routing rules to HTTP), or turn on transparent proxy on the dashboard and pick a mode from the table below. A group with several connected members uses the one with the lowest measured latency; pinning a member routes the group through it alone. Subscriptions update from the subscriptions page or on the interval set in the settings.

<img src="docs/images/screenshot.png" alt="The dashboard, light and dark" width="100%">

## Transparent proxy

| Platform | Modes |
| --- | --- |
| Linux | `redirect`, `tproxy`, `tun`; with `--lite`, the system proxy on GNOME and KDE |
| macOS | `tun`; with `--lite`, the system proxy |
| Windows | `tun`, system proxy |

The system proxy modes set the desktop's proxy settings and cover applications that honour them. The other modes intercept traffic.

With `tun`, the core opens a TUN device and assigns its address; with automatic routing on (the default), v2rayA installs the routes and the DNS setting, otherwise your own script does. The core's own connections, the direct outbound and the DNS module's upstream queries never enter the TUN (by socket mark on Linux, by binding to the physical interface on Windows and macOS). Plain DNS on port 53 that reaches the TUN is answered by the core's DNS module; encrypted DNS is not intercepted. v2rayA and the core are always excluded, and other processes can be excluded by executable name in the settings. More specific connected and static routes bypass the TUN.

Known limitation: on Windows and macOS an application that queries a LAN resolver directly still bypasses the TUN. The system resolver is covered: Windows points it at the TUN gateway, macOS at the core's listener on `127.0.0.1`, which needs port 53 free.

## Routing rules

<img src="docs/images/routinga.png" alt="The RoutingA editor, light and dark" width="100%">

RoutingA has a list mode with a form per rule and a text mode with line numbers, colouring and per-line checks. The syntax reference sits beside the editor. Import or export rules as a text file.

## Data and upgrades

The SQLite database `v2raya.db` and the generated core config live in the configuration directory: `/etc/v2raya` on Linux and macOS, `%ProgramData%\SYSTEM\v2rayA` for the Windows service, the user's config directory with `--lite`, or the path given by `--config`. Logs go to `/var/log/v2raya/v2raya.log` under the systemd and OpenRC units, or to `--log-file` / `V2RAYA_LOG_FILE`. Network requests are made only for subscription updates, rule-data downloads, latency probes and DNS.

Stop the service and back up the configuration directory before upgrading. Upgrading from a release before 2.4 migrates the BoltDB database on the first start; the old file is kept as `bolt.db.bak` (`bolt.db.bak.1` and up when that name is taken), and accounts have to be registered again. Upgrade `v2raya` and `v2raya_core` together: the dashboard shows the core version, and a mismatch is reported in a banner.

## Support

Ask questions in the [discussions](https://github.com/v2rayA/v2rayA/discussions) and report bugs in the [issues](https://github.com/v2rayA/v2rayA/issues).

Do not use this project for anything illegal.

## Credits

Founded by [@mzz2017](https://github.com/mzz2017). The Material Design 3 interface, the in-core TUN and the 2.5 service rework are by [@Zakkaus](https://github.com/Zakkaus). The OpenRC files come from the [gentoo-zh community](https://gentoozh.org). Routing data from [v2fly/domain-list-community](https://github.com/v2fly/domain-list-community), [v2fly/geoip](https://github.com/v2fly/geoip) and, for the GFWList mode, [Loyalsoldier/v2ray-rules-dat](https://github.com/Loyalsoldier/v2ray-rules-dat); the transparent-proxy rules were learnt from [zfl9/ss-tproxy](https://github.com/zfl9/ss-tproxy) and [hq450/fancyss](https://github.com/hq450/fancyss).

## License

[AGPL-3.0-only](LICENSE). The core is a fork of [Xray-core](https://github.com/XTLS/Xray-core) (MPL-2.0).
