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

from subscription_e2e import Proxy, Subscription, free_port


class Node:
    def __init__(self, name, delay, args, blackhole=False):
        self.name, self.args = name, args
        self.port, self.identity = free_port(), str(uuid.uuid4())
        self.fixture = Proxy(name, delay)
        self.process = None
        self.log = (args.output / (name + ".log")).open("w")
        outbound = {"protocol": "blackhole", "settings": {}} if blackhole else {
            "protocol": "http", "settings": {"servers": [
                {"address": "127.0.0.1", "port": self.fixture.server_address[1]}]}}
        self.config = args.output / (name + ".json")
        self.config.write_text(json.dumps({
            "log": {"loglevel": "warning"},
            "inbounds": [{"listen": "127.0.0.1", "port": self.port,
                          "protocol": "vless", "settings": {
                              "clients": [{"id": self.identity}], "decryption": "none"}}],
            "outbounds": [outbound]}))
        self.start()

    @property
    def link(self):
        return (f"vless://{self.identity}@{self.args.host_address}:{self.port}"
                f"?encryption=none&security=none&type=tcp#{self.name}")

    def start(self):
        self.process = subprocess.Popen(
            [str(self.args.xray.resolve()), "run", "-c", str(self.config.resolve())],
            stdout=self.log, stderr=subprocess.STDOUT)
        for _ in range(100):
            if self.process.poll() is not None:
                raise RuntimeError("fixture Xray exited: " + self.name)
            try:
                with socket.create_connection(("127.0.0.1", self.port), .1):
                    return
            except OSError:
                time.sleep(.02)
        raise RuntimeError("fixture Xray did not start: " + self.name)

    def stop(self):
        if self.process and self.process.poll() is None:
            self.process.terminate()
            self.process.wait(timeout=10)

    def close(self):
        self.stop()
        self.fixture.shutdown()
        self.fixture.server_close()
        self.log.close()


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

    try:
        installed = ssh("opkg status v2raya")
        expected_version = "Version: 2.2.7.3-r1" if args.expect_baseline else "Version: 2.2.7.3-r4.failover3"
        assert expected_version in installed, installed
        record("guest environment", environment=ssh(
            "cat /etc/openwrt_release; uname -a; opkg list-installed | "
            "grep -E '^(v2raya|xray-core|luci-app-v2raya|kmod-nft-tproxy) '; free"))
        token = api("account", {"username": "vmtest", "password": "disposable-vm-only"})["token"]
        assert not api("touch")["touch"]["subscriptions"], "Use a fresh disposable guest database"
        api("ports", {"socks5": 20170, "http": 20171, "socks5WithPac": 0,
                      "httpWithPac": 0, "vmess": 0}, "PUT")
        api("outbound", {"outbound": "proxy", "setting": {
            "probeURL": "http://198.18.0.1/check", "probeInterval": "10s", "type": "leastping"}}, "PUT")
        api("import", {"url": f"http://{args.host_address}:{sub.server_port}/subscription"})
        raw = api("touch")["touch"]["subscriptions"][0]
        raw["autoSelect"] = True
        api("subscription", {"subscription": raw}, "PATCH")
        settings = api("setting")["setting"]
        settings["portSharing"] = True
        api("setting", settings, "PUT")
        api("connection", {"_type": "subscriptionServer", "sub": 0, "id": 4, "outbound": "proxy"})
        api("v2ray", {}, "POST")
        check("slow")
        record("production service passes traffic through a real VLESS server", selected=selection())
        settings = api("setting")["setting"]
        settings["subscriptionAutoUpdateMode"] = "auto_update"
        api("setting", settings, "PUT")
        if args.expect_baseline:
            restart()
            time.sleep(8)
            assert selection() == "dead-first", selection()
            try:
                traffic()
            except (OSError, AssertionError, http.client.HTTPException):
                record("baseline defect reproduced: startup selects dead first node and traffic fails")
            else:
                raise AssertionError("baseline unexpectedly passed traffic")
            return
        update()
        check("fast")
        record("dead first node, wrong VLESS key and blackhole skipped; fastest sixth node selected")

        # Keep all entries in the subscription. Kill the actual chosen server.
        for lost, expected in [(fast, "backup"), (backup, "slow")]:
            lost.stop()
            update()
            check(expected)
            assert len(api("touch")["touch"]["subscriptions"][0]["servers"]) == 6
            record("failover after killing " + lost.name, selected=expected, entries=6)

        before = api("touch")["touch"]
        slow.stop()
        assert update(True)["code"] == "FAIL"
        assert api("touch")["touch"] == before
        try:
            traffic()
        except (OSError, AssertionError, http.client.HTTPException):
            pass
        else:
            raise AssertionError("all nodes dead but traffic unexpectedly succeeded")
        record("all nodes dead: error, previous state retained, no false claim of working traffic")
        fast.start()
        update()
        check("fast")
        record("recovery when a later node returns, without deleting failed entries")
        slow.start()
        backup.start()
        sub.links = [dead, slow.link, wrong_key, fast.link, blackhole.link, backup.link]
        with concurrent.futures.ThreadPoolExecutor() as executor:
            future = executor.submit(update)
            time.sleep(.5)
            assert not future.done()
            # This release's touch controller deliberately reports busy during a
            # manual update. Check the data path now and the saved identity later.
            assert traffic() == "fast"
            snapshot = ssh("free; ps w | grep '[x]ray'; ls /tmp/v2raya-subscription-*.json")
            assert api("connection", {"_type": "subscriptionServer", "sub": 0,
                                      "id": 1, "outbound": "proxy"}, fail=True)["code"] == "FAIL"
            future.result()
        check("fast")
        record("reordering and slow probes preserve live traffic; concurrent changes rejected", during_probes=snapshot)
        before = api("touch")["touch"]
        sub.status = 503
        assert update(True)["code"] == "FAIL"
        assert api("touch")["touch"] == before
        check("fast")
        sub.status, sub.links = 200, []
        assert update(True)["code"] == "FAIL"
        assert api("touch")["touch"] == before
        record("failed download and empty subscription preserve working connection")
        sub.links = links[:]

        # Test nftables/TPROXY through a request from the router with no HTTP proxy.
        settings = api("setting")["setting"]
        settings.update(transparent="proxy", transparentType="tproxy")
        api("setting", settings, "PUT")
        assert ssh("curl -fsS --noproxy '*' --max-time 5 http://198.18.0.1/transparent") == "fast"
        fast.stop()
        update()
        check("backup")
        assert ssh("curl -fsS --noproxy '*' --max-time 5 http://198.18.0.1/transparent") == "backup"
        record("nftables TPROXY router traffic follows failover", selected="backup",
               rules=ssh("nft list tables; ip rule show"))
        fast.start()
        restart()
        for _ in range(300):
            try:
                if selection() == "fast" and traffic() == "fast":
                    break
            except (OSError, AssertionError, KeyError, http.client.HTTPException):
                pass
            time.sleep(.1)
        else:
            raise AssertionError("automatic startup refresh did not select recovered later node")
        assert ssh("curl -fsS --noproxy '*' --max-time 5 http://198.18.0.1/transparent") == "fast"
        record("automatic startup refresh and transparent traffic recover after service restart")
        leftovers = ssh("find /tmp -maxdepth 1 -name 'v2raya-subscription-*'")
        assert not leftovers.strip(), leftovers
        core_count = int(ssh("pgrep -f '^/usr/bin/xray run ' | wc -l").strip())
        assert core_count == 1, core_count
        record("probe cleanup and final memory", core_processes=core_count,
               details=ssh("ps w | grep '[x]ray'; free"))
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
