# Respect device-bound TCP sockets in REDIRECT mode

This experimental, opt-in feature bypasses REDIRECT for TCP sockets that are
bound to a device **before connect()**. It matches socket state, not executable
names, service cgroups, destination addresses, or output-interface routing.

## Motivation

NetworkManager binds connectivity-probe sockets to the interface being tested.
Redirecting such a connection to a local proxy listener can prevent it from
completing, even when an ordinary unbound connection to the same endpoint works.
A resulting `limited` connectivity state can keep network-aware applications
such as Akonadi offline.

Matching `oifname` is not equivalent to detecting an explicit socket binding:
it also matches ordinary connections routed through that interface. An
application or destination whitelist only addresses particular instances.

## Enable

```sh
v2raya --redirect-respect-bound-device
```

For a systemd-managed installation, add a service override:

```ini
[Service]
Environment="V2RAYA_REDIRECT_RESPECT_BOUND_DEVICE=true"
```

The option defaults to false and is only activated in REDIRECT mode. It requires
Linux, a cgroup v2 mount at `/sys/fs/cgroup`, permission to load/attach BPF,
connect4/connect6 programs, the namespace-cookie and setsockopt helpers, and
real cgroup BPF links. Unsupported or denied attachment fails setup rather than
silently claiming the exemption is active. Both address-family hooks must load.

## Implementation

The connect hooks inspect `bpf_sock.bound_dev_if`. If it is nonzero, they preserve
existing `SO_MARK` bits and add `0x80`, which the existing v2rayA REDIRECT and DNS
rules already exempt. Unbound TCP sockets and sockets bound only to a source IP
are unchanged. No new per-packet BPF program is installed.

The hooks are attached to the root cgroup to cover local applications, not to
select particular services. A network-namespace cookie check prevents marking
sockets in other network namespaces below that cgroup, such as containers.

The program uses cilium/ebpf's Go assembler and stable Linux UAPI offsets rather
than kernel `struct sock` offsets. Building it does not require clang or kernel
headers. IPv4 and IPv6 links are installed before REDIRECT rules and released
on stop/rebuild or setup failure. Real BPF links are required, without legacy
PROG_ATTACH fallback or pinning, so process exit also releases the attachments.

If adding the mark fails for a device-bound TCP socket, the hook rejects connect
(the application may see EPERM) instead of silently allowing REDIRECT to violate
the requested exemption.

## Limits and interactions

- Only TCP connect hooks are covered. Normal ping uses ICMP and is already
  outside TCP REDIRECT. UDP, raw packet tools, and separate UDP DNS interception
  are not addressed by this feature.
- Binding a device after connect, or connections established before activation,
  are not covered. Existing NAT/conntrack mappings are not rewritten.
- Unbinding a previously marked socket does not automatically clear its mark.
  Close and recreate the socket to return to normal interception behavior.
- Other mark bits are preserved, but adding `0x80` can still affect custom policy
  routing or firewall rules that inspect that bit. It must retain v2rayA's
  existing bypass meaning on the host.
- Other BPF programs that subsequently alter binding or marks can affect the
  result. This is a routing policy, not a security-isolation boundary.
- There is added work at connect time, including a namespace-cookie helper and,
  for bound sockets, setsockopt. No performance benchmark has been completed.

## Tests

Unprivileged tests check instruction encoding and rejection of a non-cgroup
mount. They do **not** demonstrate kernel-verifier acceptance:

```sh
cd service
go test -race ./kernel/bounddevice
```

The privileged test is opt-in. It creates a disposable child cgroup, attaches
only there, and moves test subprocesses into it. Sockets target loopback; the
test does not change nftables or the production proxy and does not attach to the
root cgroup. Normal cleanup closes links and removes the directory. A forced
interruption may leave an empty test directory, but no pinned BPF program.

```sh
cd service
go test -c -o /tmp/v2raya-bound-device-tests ./kernel/bounddevice
sudo env V2RAYA_BPF_TEST=1 /tmp/v2raya-bound-device-tests -test.v
```

Native coverage includes IPv4/IPv6 ordinary sockets, source-IP-only binding,
device binding, mark preservation, existing bypass marks, UDP exclusion,
detachment, and a mismatched namespace cookie. The cookie mismatch test models
a different namespace; it does not create a separate network namespace.

Before enabling by default or deploying, complete the privileged test,
end-to-end REDIRECT tests, restart/crash lifecycle checks, and a connect-rate
benchmark. Disable any existing destination-IP probe exemption during the
end-to-end test so it cannot mask failure of this implementation.
