"""Loopback-only HTTP/subscription fixtures and real VLESS server processes."""
import base64
import http.server
import json
import socket
import socketserver
import subprocess
import threading
import time
import uuid

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


