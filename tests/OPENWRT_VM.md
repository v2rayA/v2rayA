# OpenWrt VM validation

Use a disposable ARM64 OpenWrt VM. The test creates a local account, changes v2rayA settings, installs a synthetic subscription and restarts the guest service. It is not intended for a production router.

## Images and versions

The validation uses official OpenWrt 24.10.4 `armsr/armv8` images from:

<https://downloads.openwrt.org/releases/24.10.4/targets/armsr/armv8/>

| Image | SHA-256 |
| --- | --- |
| `openwrt-24.10.4-armsr-armv8-generic-kernel.bin` | `5175c026ad2fb75ef7db7e6b61889e9304ad444081bb972831f451cc58f974b2` |
| `openwrt-24.10.4-armsr-armv8-generic-ext4-rootfs.img.gz` | `d448813d12bc9b6bb54bcc6fb599dd1aa73ac55fa446163d65100bed8b8bf8ab` |

The guest runs Linux 6.6.110 and the official `xray-core` 25.1.30-r1. Package feeds may rebuild packages without changing their upstream version; record the current package index and validate download checksums. The host fixtures also use Xray 25.1.30.

## Start the VM

Decompress the root filesystem image to `rootfs.img`. On an Apple Silicon Mac with QEMU installed:

```sh
qemu-system-aarch64 -machine virt -accel hvf -cpu host -smp 2 -m 256 \
  -kernel openwrt-24.10.4-armsr-armv8-generic-kernel.bin \
  -append 'root=/dev/vda rootwait console=ttyAMA0' \
  -drive if=none,file=rootfs.img,format=raw,id=root \
  -device virtio-blk-pci,drive=root \
  -netdev user,id=wan,hostfwd=tcp:127.0.0.1:22022-:22,hostfwd=tcp:127.0.0.1:22017-:2017,hostfwd=tcp:127.0.0.1:22171-:20171,hostfwd=tcp:127.0.0.1:22080-:80 \
  -device virtio-net-pci,netdev=wan -device virtio-rng-pci \
  -nographic
```

The forwarded ports bind only to host loopback. Use `-accel kvm -cpu host` on a suitable Linux ARM64 host, or software emulation for another host architecture. These alternative host configurations were not part of the recorded validation.

Press Enter at the serial console. Configure the guest's single interface for the QEMU DHCP network:

```sh
uci set network.lan.proto=dhcp
uci -q delete network.lan.ipaddr
uci -q delete network.lan.netmask
uci commit network
/etc/init.d/network restart
```

Generate a disposable SSH key on the host and put its public key into the guest's `/etc/dropbear/authorized_keys` with mode 600. Start Dropbear. Use an explicit key, `IdentitiesOnly=yes`, a separate known-hosts file and `-F /dev/null` for test connections.

Install official guest packages:

```sh
opkg update
opkg install v2raya luci-app-v2raya xray-core v2ray-geoip v2ray-geosite curl
uci set v2raya.config.enabled=1
uci commit v2raya
/etc/init.d/v2raya start
```

The original ext4 filesystem is small. Enlarge it offline before provisioning if necessary; merely extending the disk file does not enlarge ext4. Online resizing of this stock image failed with a reserved-block layout error during validation. A replacement of an existing executable can also trigger opkg's conservative free-space check. This is a test-image preparation issue, not a reason to bypass the space check on the physical router.

## Compare baseline and candidate

Run against a fresh database with the official package first:

```sh
python3 tests/openwrt_vm_e2e.py --disposable-vm \
  --xray /absolute/path/to/host/xray \
  --ssh-key /absolute/path/to/disposable/key \
  --output /tmp/openwrt-baseline --expect-baseline
```

The baseline check expects the original defect: automatic refresh selects the dead first node and traffic fails even though later nodes work.

Restore a clean guest snapshot or stop the service and archive its disposable test database. Wait for both the v2rayA process and its Xray process to exit before moving the database or restarting: the init script can return before termination is complete. Install the candidate package. The VM uses `aarch64_generic`; the physical-router package uses `aarch64_cortex-a53`. Both can run this static baseline-ARM64 executable. In this isolated VM only, a compatible opkg architecture list can contain:

```text
arch all 1
arch noarch 1
arch aarch64_generic 10
arch aarch64_cortex-a53 20
```

Explicitly listing one architecture replaces opkg's implicit defaults, so retain all existing architectures. Do not copy this compatibility override to the physical router. Verify both the installed package version and the SHA-256 of `/usr/bin/v2raya` against the candidate binary.

Run the full suite:

```sh
python3 tests/openwrt_vm_e2e.py --disposable-vm \
  --xray /absolute/path/to/host/xray \
  --ssh-key /absolute/path/to/disposable/key \
  --output /tmp/openwrt-candidate
```

Each healthy node is a separate host Xray VLESS server, backed by a deterministic HTTP fixture returning that node's identity. Guest requests traverse real Xray/VLESS connections. Fixture delays make the intended ordering repeatable. All listeners are on host loopback; the guest reaches them through QEMU's `10.0.2.2` host address. No real subscriptions or VPN credentials are used.

The subscription includes six entries: a closed first endpoint, an invalid VLESS credential, an authenticated blackhole, and three working endpoints. The test keeps all entries while killing the fastest node and then its replacement. It checks the saved server identity and actual traffic, full outage, restoration, reordering, failed and empty downloads, uninterrupted traffic during probes, rejection of concurrent mutations, nftables/TPROXY traffic originating on the guest, and automatic refresh after a service restart. It also records memory and verifies that temporary probe files and processes are gone.

The base release deliberately returns busy from its `touch` endpoint while a manual refresh is in progress. The suite checks traffic during that interval, then checks the saved identity after completion.

## Exercise the real ticker inside the guest

Cross-compile the existing integration test from `service/`:

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go test -c -o /tmp/subscription-timer.test .
```

Copy it into the guest and run:

```sh
V2RAYA_V2RAY_BIN=/usr/bin/xray /tmp/subscription-timer.test \
  -test.v -test.run '^TestSubscriptionTimerSelectsHealthyServer$'
```

This fires the actual scheduler's ticker with a short test-only duration and real Xray probes. It does not change the production interval or claim to have waited for a full one-hour production interval. The test uses an isolated database and Lite mode; the application suite above uses the actual OpenWrt service in normal root mode.

## Limits

The VM exercises the same OpenWrt release and ARM64 application binary, but does not emulate the Cudy SoC, Wi-Fi, flash layout or physical LAN. The transparent traffic check originates on the guest; it does not represent a separate LAN client's forwarding path. The test endpoints use plain VLESS over TCP on an isolated local network. Provider-specific TLS/REALITY, external plugins and real subscription credentials need separate validation.

## Exercise production monitoring intervals

Reset the disposable v2rayA database after the refresh suite and run:

```sh
python3 tests/subscription_monitor_e2e.py --disposable-vm \
  --xray /absolute/path/to/host/xray \
  --ssh-key /absolute/path/to/disposable/key \
  --output /tmp/openwrt-monitor
```

This test disables the ordinary update schedule and leaves Auto-Connect off. It checks the default-off monitoring setting, healthy traffic with one core and no subscription downloads, a transient outage, an actual sustained outage exceeding one minute, real VLESS and TPROXY recovery, immediate retries when all nodes fail, recovery after a node returns, cancellation during a blocked subscription download, switch-off behavior, manual stop and persistence across an OpenWrt service restart. The production timer is not shortened. Use a fresh test database for each invocation.

## Selection policy and additional failure cases

With a fresh disposable database, run:

```sh
python3 tests/subscription_policy_e2e.py --disposable-vm \
  --xray /absolute/path/to/host/xray \
  --ssh-key /absolute/path/to/disposable/key \
  --output /tmp/openwrt-policy
```

This exercises both GUI-equivalent selection modes, first-position changes, an unavailable replacement first entry during recovery, empty updates, saving policy during a full outage, refresh with monitoring/Auto-Connect disabled, a subscription HTTP 503 with cached candidates, disabled public proxy listeners, unexpected local Xray death, manual stop during policy changes, and restart persistence. Production failure windows are unchanged, so the suite takes several minutes. It validates actual VLESS and guest-originated TPROXY traffic.
