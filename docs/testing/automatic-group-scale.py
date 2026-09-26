#!/usr/bin/env python3
"""Measure real automatic-group probes in a disposable, preinstalled OpenWrt VM.

The VM uses QEMU user networking (host = 10.0.2.2), a 256 MiB guest, and a
matching service/core pair. It must contain only test data. See the adjacent
automatic-group-review.md for commands, topology and measurement limits.
"""
import argparse
import base64
import http.server
import json
import re
from pathlib import Path
import shlex
import socket
import socketserver
import subprocess
import threading
import time
import urllib.request
import uuid


class HTTP(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == "/subscription":
            body = base64.b64encode("\n".join(self.server.links).encode())
        else:
            self.server.probes += 1
            body = b"healthy"
        self.send_response(200)
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, *_args):
        pass


class Silent(socketserver.ThreadingTCPServer):
    daemon_threads = True


class Hold(socketserver.BaseRequestHandler):
    def handle(self):
        self.server.accepts += 1
        self.request.settimeout(30)
        try:
            while self.request.recv(65536):
                pass
        except OSError:
            pass


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--xray", type=Path, required=True)
    p.add_argument("--ssh-key", type=Path, required=True)
    p.add_argument("--known-hosts", type=Path, required=True)
    p.add_argument("--ssh-port", type=int, default=25922)
    p.add_argument("--api-port", type=int, default=25917)
    p.add_argument("--output", type=Path, required=True)
    p.add_argument("--nodes", type=int, default=200)
    p.add_argument("--timeouts", type=int, default=0)
    p.add_argument("--trace", action="store_true", help="audit core lifetimes with guest strace")
    p.add_argument("--cancel-after-starts", type=int, default=0, help="exercise an editing request during a slow pass")
    args = p.parse_args()
    assert 0 <= args.timeouts < args.nodes
    args.output.mkdir(parents=True, exist_ok=True)
    ssh_base = ["ssh", "-F", "/dev/null", "-i", str(args.ssh_key.resolve()),
                "-o", "IdentitiesOnly=yes", "-o", "StrictHostKeyChecking=yes",
                "-o", "UserKnownHostsFile=" + str(args.known_hosts.resolve()),
                "-p", str(args.ssh_port), "root@127.0.0.1"]

    def ssh(command):
        return subprocess.check_output(ssh_base + [command], timeout=60).decode()

    token = None

    def api(path, data=None, method=None):
        headers = {"Content-Type": "application/json"}
        if token:
            headers["Authorization"] = "Bearer " + token
        req = urllib.request.Request(f"http://127.0.0.1:{args.api_port}/api/{path}",
                                     headers=headers, method=method,
                                     data=None if data is None else json.dumps(data).encode())
        with urllib.request.urlopen(req, timeout=60) as response:
            reply = json.load(response)
        assert reply["code"] == "SUCCESS", reply
        return reply["data"]

    def wait(fn, timeout):
        deadline = time.monotonic() + timeout
        while time.monotonic() < deadline:
            try:
                if fn():
                    return
            except (OSError, AssertionError):
                pass
            time.sleep(0.1)
        raise AssertionError("timed out waiting for VM result")

    fixture_http = http.server.ThreadingHTTPServer(("127.0.0.1", 0), HTTP)
    fixture_http.probes, fixture_http.links = 0, []
    silent = Silent(("127.0.0.1", 0), Hold)
    silent.accepts = 0
    for srv in (fixture_http, silent):
        threading.Thread(target=srv.serve_forever, daemon=True).start()
    with socket.socket() as s:
        s.bind(("127.0.0.1", 0))
        vless_port = s.getsockname()[1]
    identity = str(uuid.uuid4())
    fixture_config = args.output / "fixture.json"
    fixture_config.write_text(json.dumps({
        "log": {"loglevel": "warning"},
        "inbounds": [{"listen": "127.0.0.1", "port": vless_port, "protocol": "vless",
                      "settings": {"clients": [{"id": identity}], "decryption": "none"}}],
        "outbounds": [{"protocol": "freedom"}],
    }))
    fixture_log = (args.output / "fixture.log").open("w")
    fixture = subprocess.Popen([str(args.xray.resolve()), "run", "-c", str(fixture_config.resolve())],
                               stdout=fixture_log, stderr=subprocess.STDOUT)
    ready = args.nodes - args.timeouts
    fixture_http.links = [f"vless://{identity}@10.0.2.2:{vless_port if i < ready else silent.server_address[1]}"
                  f"?encryption=none&security=none&type=tcp#candidate-{i:03d}"
                  for i in range(args.nodes)]
    backup = "/etc/v2raya/v2raya.db.scale-review-backup"
    # The wrapper logs one line per probe process before exec preserves its PID.
    wrapper = '''#!/bin/sh
case "$*" in *v2raya-subscription-*.json*)
  printf '%s %s\\n' "$$" "$(cut -d ' ' -f 1 /proc/uptime)" >> /tmp/group-probe-starts;;
esac
exec /usr/bin/v2raya_core.scale-real "$@"
'''
    wrapped = False
    try:
        ssh("/etc/init.d/v2raya stop; "
            f"cp /etc/v2raya/v2raya.db {backup}; "
            "rm -f /etc/v2raya/v2raya.db /etc/v2raya/v2raya.db-wal /etc/v2raya/v2raya.db-shm; "
            "mv /usr/bin/v2raya_core /usr/bin/v2raya_core.scale-real; "
            "printf %s " + shlex.quote(wrapper) + " > /usr/bin/v2raya_core; "
            "chmod +x /usr/bin/v2raya_core; /etc/init.d/v2raya start")
        wrapped = True
        wait(lambda: api("version"), 30)
        token = api("account", {"username": "vmtest", "password": "disposable-vm-only"})["token"]
        api("import", {"kind": "subscription", "url": f"http://10.0.2.2:{fixture_http.server_port}/subscription"}, "POST")
        assert len(api("touch")["touch"]["subscriptions"][0]["servers"]) == args.nodes
        if args.trace:
            ssh("strace -f -tt -s 256 -e trace=process -p $(pidof v2raya) -o /tmp/group-process.trace >/tmp/group-strace.log 2>&1 & echo $! >/tmp/group-strace.pid")
        ssh("rm -f /tmp/group-probe-starts /tmp/group-rss; touch /tmp/group-sampling; "
            "(while [ -e /tmp/group-sampling ]; do rss=0; cores=0; "
            "for p in $(pidof v2raya v2raya_core.scale-real 2>/dev/null); do "
            "v=$(awk '/VmRSS:/{print $2}' /proc/$p/status 2>/dev/null); rss=$((rss+${v:-0})); "
            "cmd=$(tr '\\0' ' ' </proc/$p/cmdline 2>/dev/null); "
            "case \"$cmd\" in *v2raya-subscription-*.json*) cores=$((cores+1));; esac; done; "
            "echo \"$(cut -d ' ' -f 1 /proc/uptime) $rss $cores\" >> /tmp/group-rss; "
            "sleep 0.02; done) >/dev/null 2>&1 &")
        started = time.monotonic()
        api("outbound", {"outbound": "proxy", "setting": {
            "autoAdd": True, "probeURL": f"http://127.0.0.1:{fixture_http.server_port}/health",
            "probeInterval": "300s", "type": "leastping"}}, "PUT")
        if args.cancel_after_starts:
            wait(lambda: int(ssh("wc -l </tmp/group-probe-starts")) >= args.cancel_after_starts, 60)
            edit_started = time.monotonic()
            api("outbound", {"outbound": "proxy", "setting": {
                "autoAdd": False, "probeURL": f"http://127.0.0.1:{fixture_http.server_port}/health",
                "probeInterval": "300s", "type": "leastping"}}, "PUT")
            edit_seconds = time.monotonic() - edit_started
            wait(lambda: int(ssh("ps w | grep '[v]2raya-subscription-' | wc -l")) == 0, 15)
            result = {"test": "editing during a long probe pass", "editing_seconds": round(edit_seconds, 3),
                      "code": "SUCCESS", "partial_membership_applied": bool(api("touch")["touch"].get("connectedServer"))}
            assert not result["partial_membership_applied"]
            (args.output / "results.json").write_text(json.dumps(result, indent=2) + "\n")
            print(json.dumps(result, indent=2), flush=True)
            return
        wait(lambda: len(api("touch")["touch"].get("connectedServer") or []) == ready,
             args.nodes * 6 + 60)
        elapsed = time.monotonic() - started
        samples = ssh("rm -f /tmp/group-sampling; sleep 1; cat /tmp/group-rss")
        starts = ssh("cat /tmp/group-probe-starts")
        (args.output / "rss.tsv").write_text(samples)
        (args.output / "probe-starts.tsv").write_text(starts)
        rows = [row.split() for row in samples.splitlines()]
        result = {
            "nodes": args.nodes, "supported_vless_nodes": args.nodes,
            "reachable": ready, "timeouts": args.timeouts,
            "probe_timeout_seconds": 5, "completed": args.nodes, "cancelled": 0,
            "group_members_after_pass": len(api("touch")["touch"].get("connectedServer") or []),
            "probe_core_starts": len(starts.splitlines()),
            "origin_requests": fixture_http.probes, "silent_tcp_accepts": silent.accepts,
            "seconds": round(elapsed, 3),
            "peak_sampled_combined_rss_kib": max(int(r[1]) for r in rows),
            "peak_sampled_probe_cores": max(int(r[2]) for r in rows),
            "sampling_interval_seconds": 0.02, "main_core_running": False,
            "topology": f"{args.nodes} catalog records with unique names; one local VLESS endpoint and one silent TCP endpoint; QEMU HVF ARM64, 2 vCPU, 256 MiB",
        }
        assert result["probe_core_starts"] == args.nodes, result
        assert fixture_http.probes == ready and silent.accepts == args.timeouts, result
        if args.trace:
            ssh("kill $(cat /tmp/group-strace.pid); sleep 1")
            trace = ssh("cat /tmp/group-process.trace")
            (args.output / "process.trace").write_text(trace)
            active, traced_starts, peak = set(), 0, 0
            for line in trace.splitlines():
                match = re.match(r"(\d+)\s", line)
                if not match:
                    continue
                pid = int(match.group(1))
                if 'execve("/usr/bin/v2raya_core.scale-real"' in line and 'v2raya-subscription-' in line:
                    active.add(pid)
                    traced_starts += 1
                    peak = max(peak, len(active))
                elif '+++ exited with ' in line or '+++ killed by ' in line:
                    active.discard(pid)
            result.update(traced_core_starts=traced_starts, peak_traced_core_lifetimes=peak)
            assert traced_starts == args.nodes and not active and peak <= 2, result
        if args.timeouts:
            assert elapsed >= args.timeouts * 5 / 2, result
        (args.output / "results.json").write_text(json.dumps(result, indent=2) + "\n")
        print(json.dumps(result, indent=2), flush=True)
    finally:
        if wrapped:
            ssh("rm -f /tmp/group-sampling; /etc/init.d/v2raya stop; "
                "mv /usr/bin/v2raya_core.scale-real /usr/bin/v2raya_core; "
                f"cp {backup} /etc/v2raya/v2raya.db; "
                "rm -f /etc/v2raya/v2raya.db-wal /etc/v2raya/v2raya.db-shm; /etc/init.d/v2raya start")
        fixture.terminate()
        fixture.wait(timeout=10)
        fixture_log.close()
        for srv in (fixture_http, silent):
            srv.shutdown()
            srv.server_close()


if __name__ == "__main__":
    main()
