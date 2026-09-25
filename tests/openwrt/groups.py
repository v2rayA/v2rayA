#!/usr/bin/env python3
"""Test a disposable OpenWrt VM through real VLESS servers on the host.

The guest must have a fresh v2rayA database, its API/HTTP ports forwarded,
and key-only SSH access. This changes the guest's v2rayA settings and restarts
its service. Never point this test at a production router.
"""
import argparse
import concurrent.futures
import http.client
import http.server
import json
import os
from pathlib import Path
import socket
import subprocess
import threading
import time
import urllib.request
import uuid

from fixtures import Node, Subscription, free_port


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--disposable-vm", action="store_true", required=True)
    p.add_argument("--xray", type=Path, required=True)
    p.add_argument("--output", type=Path, required=True)
    p.add_argument("--ssh-key", type=Path, required=True)
    p.add_argument("--ssh-port", type=int, default=22022)
    p.add_argument("--api-port", type=int, default=22017)
    p.add_argument("--http-port", type=int, default=22171)
    p.add_argument("--host-address", default="10.0.2.2")
    p.add_argument("--expect-baseline", action="store_true")
    args = p.parse_args()
    args.output.mkdir(parents=True, exist_ok=True)
    token, results = None, []
    base = f"http://127.0.0.1:{args.api_port}/api/"
    nodes = [Node("slow", .4, args), Node("backup", .15, args),
             Node("fast", .01, args), Node("blackhole", 0, args, True)]
    slow, backup, fast, blackhole = nodes
    dead = (f"vless://{uuid.uuid4()}@{args.host_address}:{free_port()}"
            "?encryption=none&security=none&type=tcp#dead-first")
    wrong_key = fast.link.replace(fast.identity, str(uuid.uuid4())).replace("#fast", "#wrong-key")
    links = [dead, wrong_key, blackhole.link, slow.link, backup.link, fast.link]
    sub = http.server.ThreadingHTTPServer(("127.0.0.1", 0), Subscription)
    sub.links, sub.status = links[:], 200
    threading.Thread(target=sub.serve_forever, daemon=True).start()

    def ssh(command, timeout=60):
        return subprocess.check_output([
            "ssh", "-F", "/dev/null", "-i", str(args.ssh_key.resolve()),
            "-o", "IdentitiesOnly=yes", "-o", "StrictHostKeyChecking=accept-new",
            "-o", "UserKnownHostsFile=" + str((args.output / "known_hosts").resolve()),
            "-p", str(args.ssh_port), "root@127.0.0.1", command],
            timeout=timeout, stderr=subprocess.STDOUT).decode()

    def api(path, data=None, method=None, fail=False):
        headers = {"Content-Type": "application/json"}
        if token:
            headers["Authorization"] = "Bearer " + token
        request = urllib.request.Request(base + path, headers=headers, method=method,
                                         data=None if data is None else json.dumps(data).encode())
        with urllib.request.urlopen(request, timeout=90) as response:
            result = json.load(response)
        if not fail:
            assert result["code"] == "SUCCESS", result
        return result if fail else result["data"]

    def selection():
        touch = api("touch")["touch"]
        current = next(w for w in touch["connectedServer"] if w["outbound"] == "proxy")
        return touch["subscriptions"][current["sub"]]["servers"][current["id"] - 1]["name"]

    def traffic():
        conn = http.client.HTTPConnection("127.0.0.1", args.http_port, timeout=3)
        try:
            conn.request("GET", "http://198.18.0.1/traffic")
            response = conn.getresponse()
            assert response.status == 200, response.status
            return response.read().decode()
        finally:
            conn.close()

    def update(fail=False):
        return api("subscription", {"_type": "subscription", "id": 1}, "PUT", fail)

    def record(name, **evidence):
        results.append({"test": name, "result": "PASS", **evidence})
        (args.output / "results.json").write_text(json.dumps(results, indent=2) + "\n")
        print("PASS:", name, evidence or "", flush=True)

    def check(name):
        assert selection() == name, (selection(), name)
        assert traffic() == name, (traffic(), name)

    def restart():
        nonlocal token
        ssh("/etc/init.d/v2raya restart")
        for _ in range(300):
            try:
                token = api("login", {"username": "vmtest", "password": "disposable-vm-only"})["token"]
                return
            except (OSError, AssertionError, KeyError):
                time.sleep(.1)
        raise RuntimeError("guest v2rayA did not restart")

    def await_traffic(name, timeout=45):
        deadline = time.monotonic()+timeout
        while time.monotonic()<deadline:
            try:
                if traffic()==name: return
            except (OSError, AssertionError, http.client.HTTPException): pass
            time.sleep(1)
        raise AssertionError("traffic did not switch to "+name)

    try:
        version=api("version")
        assert version["coreVersionValid"], version
        token = api("account", {"username": "vmtest", "password": "disposable-vm-only"})["token"]
        api("ports", {"socks5": 20170, "http": 20171, "socks5WithPac": 0, "httpWithPac": 0, "vmess": 0}, "PUT")
        api("outbound", {"outbound":"proxy", "setting":{"probeURL":"http://198.18.0.1/check", "probeInterval":"10s", "type":"leastping"}}, "PUT")
        api("import", {"url":f"http://{args.host_address}:{sub.server_port}/subscription"})
        raw=api("touch")["touch"]["subscriptions"][0]
        raw.update(autoSelect=True, monitor=True)
        api("subscription", {"subscription":raw}, "PATCH")
        settings=api("setting")["setting"]
        assert "br-*" not in settings["tproxyExcludedInterfaces"], settings
        settings.update(portSharing=True, transparent="proxy", transparentType="tproxy", subscriptionAutoUpdateMode="none")
        api("setting", settings, "PUT")
        api("connection", {"_type":"subscriptionServer", "sub":0,"id":4,"outbound":"proxy"})
        api("v2ray", {}, "POST")
        update()
        await_traffic("fast")
        connected=api("touch")["touch"]["connectedServer"]
        assert len(connected)==len(sub.links), connected
        record("group refresh retains all members and routes through fastest working VLESS server")
        fast.stop()
        await_traffic("backup")
        record("group falls back to later healthy member when fastest server disappears")
        slow.stop(); backup.stop()
        before=api("touch")["touch"]
        update(fail=True)
        after=api("touch")["touch"]
        assert before["connectedServer"]==after["connectedServer"]
        assert before["subscriptions"][0]["address"]==after["subscriptions"][0]["address"]
        record("all-dead group preserves its membership")
        fast.start()
        await_traffic("fast",100)
        record("group resumes actual traffic after a server returns")
        restart()
        await_traffic("fast")
        assert len(api("touch")["touch"]["connectedServer"])==len(sub.links)
        record("group membership and traffic survive service restart")
        assert ssh("curl -fsS --max-time 10 http://198.18.0.1/traffic").strip()=="fast"
        record("current defaults pass router-originated TPROXY traffic")
    finally:
        try:
            (args.output / "guest.log").write_text(ssh("cat /var/log/v2raya/v2raya.log"))
        except (OSError, subprocess.SubprocessError):
            pass
        sub.shutdown()
        sub.server_close()
        for node in nodes:
            node.close()


if __name__ == "__main__":
    main()
