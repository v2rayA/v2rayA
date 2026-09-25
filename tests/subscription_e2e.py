#!/usr/bin/env python3
"""Exercise the running v2rayA API with a real Xray and loopback-only fixtures."""
import argparse
import base64
import concurrent.futures
import http.client
import http.server
import json
import os
from pathlib import Path
import shlex
import signal
import socket
import socketserver
import subprocess
import tempfile
import threading
import time
import urllib.request


class Proxy(socketserver.ThreadingTCPServer):
    allow_reuse_address = True
    daemon_threads = True

    def __init__(self, name, delay=0, status=200):
        self.name, self.delay, self.status = name, delay, status
        self.requests = 0
        super().__init__(("127.0.0.1", 0), Tunnel)
        threading.Thread(target=self.serve_forever, daemon=True).start()

    @property
    def link(self):
        return f"http-proxy://127.0.0.1:{self.server_address[1]}#{self.name}"


class Tunnel(socketserver.StreamRequestHandler):
    def handle(self):
        try:
            self.connection.settimeout(12)
            first = self.rfile.readline()
            while self.rfile.readline() not in (b"\r\n", b"\n", b""):
                pass
            if first.startswith(b"CONNECT "):
                self.wfile.write(b"HTTP/1.1 200 Connection established\r\n\r\n")
                self.wfile.flush()
                if not self.rfile.readline():
                    return
                while self.rfile.readline() not in (b"\r\n", b"\n", b""):
                    pass
            self.server.requests += 1
            time.sleep(self.server.delay)
            body = self.server.name.encode()
            self.wfile.write(f"HTTP/1.1 {self.server.status} Result\r\nContent-Length: {len(body)}\r\nConnection: close\r\n\r\n".encode() + body)
        except (OSError, ValueError):
            pass


class Subscription(http.server.BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(self.server.status)
        self.end_headers()
        self.wfile.write(base64.b64encode("\n".join(self.server.links).encode()))

    def log_message(self, *args):
        pass


def free_port():
    with socket.socket() as s:
        s.bind(("127.0.0.1", 0))
        return s.getsockname()[1]


def main():
    p = argparse.ArgumentParser()
    p.add_argument("--binary", type=Path, required=True)
    p.add_argument("--xray", type=Path, required=True)
    p.add_argument("--output", type=Path, required=True)
    p.add_argument("--keep-running", action="store_true")
    args = p.parse_args()
    args.output.mkdir(parents=True, exist_ok=True)
    root = Path(tempfile.mkdtemp(prefix="runtime-", dir=args.output)).resolve()
    config = root / "config"
    config.mkdir()
    slow, fast, blackhole = Proxy("slow", .25), Proxy("fast", .01), Proxy("blackhole", 8)
    closed_port = free_port()
    dead_link = f"http-proxy://127.0.0.1:{closed_port}#closed"
    sub = http.server.ThreadingHTTPServer(("127.0.0.1", 0), Subscription)
    sub.links, sub.status = [dead_link, slow.link, fast.link], 200
    threading.Thread(target=sub.serve_forever, daemon=True).start()
    sub_url = f"http://127.0.0.1:{sub.server_port}/subscription"
    api_port, http_port = free_port(), free_port()
    base = f"http://127.0.0.1:{api_port}"
    fail_marker = root / "fail-next-start"
    hook = root / "core-hook.sh"
    marker = shlex.quote(str(fail_marker))
    hook.write_text(f'#!/bin/sh\nif [ "$1" = "--stage=pre-start" ] && [ -f {marker} ]; then\n  rm {marker}\n  exit 1\nfi\n')
    hook.chmod(0o700)
    log = (root / "app.log").open("w")
    command = [str(args.binary.resolve()), "--lite", "--address", f"127.0.0.1:{api_port}", "--config", str(config), "--v2ray-bin", str(args.xray.resolve()), "--v2ray-assetsdir", str(args.xray.resolve().parent), "--core-hook", str(hook)]
    process = subprocess.Popen(command, stdout=log, stderr=subprocess.STDOUT, start_new_session=True)
    token = None
    results = []

    def api(path, data=None, method=None, fail=False):
        headers = {"Content-Type": "application/json"}
        if token:
            headers["Authorization"] = "Bearer " + token
        request = urllib.request.Request(base + "/api/" + path, data=None if data is None else json.dumps(data).encode(), headers=headers, method=method)
        with urllib.request.urlopen(request, timeout=60) as response:
            result = json.load(response)
        if not fail:
            assert result["code"] == "SUCCESS", result
        return result if fail else result["data"]

    def record(name):
        results.append(name)
        print("PASS:", name, flush=True)

    def selection():
        touch = api("touch")["touch"]
        current = next(w for w in touch["connectedServer"] if w["outbound"] == "proxy")
        return touch["subscriptions"][current["sub"]]["servers"][current["id"] - 1]["name"]

    def update(fail=False):
        return api("subscription", {"_type": "subscription", "id": 1}, "PUT", fail)

    def traffic():
        conn = http.client.HTTPConnection("127.0.0.1", http_port, timeout=2)
        try:
            conn.request("GET", "http://probe.invalid/traffic")
            response = conn.getresponse()
            assert response.status == 200
            return response.read().decode()
        finally:
            conn.close()

    try:
        for _ in range(200):
            if process.poll() is not None:
                raise RuntimeError("v2rayA exited; see " + str(root / "app.log"))
            try:
                api("version")
                break
            except OSError:
                time.sleep(.05)
        else:
            raise RuntimeError("API did not start")
        token = api("account", {"username": "localtest", "password": "local-test-only"})["token"]
        api("ports", {"socks5": free_port(), "http": http_port, "socks5WithPac": 0, "httpWithPac": 0, "vmess": 0}, "PUT")
        api("outbound", {"outbound": "proxy", "setting": {"probeURL": "http://probe.invalid/check", "probeInterval": "10s", "type": "leastping"}}, "PUT")
        api("import", {"url": sub_url})
        raw = api("touch")["touch"]["subscriptions"][0]
        raw["autoSelect"] = True
        api("subscription", {"subscription": raw}, "PATCH")
        api("connection", {"_type": "subscriptionServer", "sub": 0, "id": 2, "outbound": "proxy"})
        api("v2ray", {}, "POST")
        assert traffic() == "slow"
        record("application starts and routes real traffic through Xray")
        update()
        assert selection() == "fast" and traffic() == "fast"
        record("manual subscription refresh skips closed first node and selects fastest reachable node")

        sub.links = [blackhole.link, slow.link, fast.link]
        with concurrent.futures.ThreadPoolExecutor() as executor:
            future = executor.submit(update)
            time.sleep(.3)
            assert not future.done(), "did not wait for blackhole timeout"
            assert traffic() == "fast", "active connection interrupted by probes"
            competing = api("connection", {"_type": "subscriptionServer", "sub": 0, "id": 1, "outbound": "proxy"}, fail=True)
            assert competing["code"] == "FAIL", competing
            future.result()
        assert selection() == "fast"
        record("waits for every probe while old traffic works; simultaneous mutation is rejected")

        sub.links = [slow.link, fast.link, dead_link]
        update()
        assert selection() == "fast" and traffic() == "fast"
        record("subscription reorder preserves correct server references")
        before = api("touch")["touch"]
        sub.links = [dead_link]
        assert update(fail=True)["code"] == "FAIL"
        assert api("touch")["touch"] == before and traffic() == "fast"
        record("all nodes unavailable preserves subscription, selection and active traffic")

        sub.status = 503
        assert update(fail=True)["code"] == "FAIL"
        assert api("touch")["touch"] == before and traffic() == "fast"
        sub.status, sub.links = 200, []
        assert update(fail=True)["code"] == "FAIL"
        assert api("touch")["touch"] == before
        record("HTTP download errors and empty subscriptions preserve previous configuration")

        sub.links = [slow.link]
        fail_marker.touch()
        assert update(fail=True)["code"] == "FAIL"
        assert api("touch")["touch"] == before and traffic() == "fast"
        record("failed core restart rolls back database and restores active core")

        sub.links = [dead_link, slow.link]
        update()
        assert selection() == "slow" and traffic() == "slow"
        record("removal of current node switches to reachable replacement from same subscription")

        settings = api("setting")["setting"]
        settings["subscriptionAutoUpdateMode"] = "auto_update"
        api("setting", settings, "PUT")
        os.killpg(process.pid, signal.SIGTERM)
        process.wait(timeout=15)
        sub.links = [dead_link, fast.link, slow.link]
        process = subprocess.Popen(command, stdout=log, stderr=subprocess.STDOUT, start_new_session=True)
        for _ in range(300):
            try:
                # JWT signing keys are regenerated by this release on restart.
                if process.poll() is not None:
                    raise RuntimeError("v2rayA exited after restart")
                token = api("login", {"username": "localtest", "password": "local-test-only"})["token"]
                if selection() == "fast" and traffic() == "fast":
                    break
            except (OSError, AssertionError, KeyError):
                pass
            time.sleep(.05)
        else:
            raise AssertionError("startup automatic refresh did not select healthy server")
        record("automatic startup refresh selects healthy node after restoring previous core")
        (args.output / "e2e-results.json").write_text(json.dumps({"passed": results, "runtime": str(root), "url": base}, indent=2) + "\n")
        if args.keep_running:
            print("DEMO:", base, "login localtest / local-test-only", flush=True)
            (args.output / "demo.json").write_text(json.dumps({"url": base, "username": "localtest", "password": "local-test-only", "pid": process.pid, "runtime": str(root)}) + "\n")
            threading.Event().wait()
    finally:
        if process.poll() is None:
            os.killpg(process.pid, signal.SIGTERM)
            process.wait(timeout=15)
        log.close()
        sub.shutdown()
        for fixture in (slow, fast, blackhole):
            fixture.shutdown()
            fixture.server_close()


if __name__ == "__main__":
    main()
