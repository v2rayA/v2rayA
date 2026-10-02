# OpenWrt 24.10.4 strategy VM

`openwrt-24.10.4-vm.ps1` is the reproducible target-platform check for proxy
group selection. It downloads the official x86-64 ext4 image, verifies its
SHA-256 digest, cross-compiles the affected Go test packages, and boots a fresh
copy of the image with QEMU. The script verifies the release and architecture
inside the guest before it runs the tests.

The guest uses QEMU user networking. The script waits for `br-lan` through the
serial console, assigns the conventional `10.0.2.15` guest address, and installs
an ephemeral SSH key. It does not modify the downloaded base image. The run
image is removed after the check unless `-KeepVM` is specified.

Requirements are PowerShell 7, Go 1.26, QEMU `qemu-system-x86_64`, OpenSSH
`ssh`, `scp`, and `ssh-keygen`. From the repository root:

```powershell
./docs/testing/openwrt-24.10.4-vm.ps1 `
  -Qemu (Get-Command qemu-system-x86_64).Source `
  -Go (Get-Command go).Source
```

The matching GitHub Actions workflow is
`.github/workflows/openwrt-24.10.4.yml`. Serial and QEMU logs are uploaded even
when a CI run fails.
