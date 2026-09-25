#!/usr/bin/env python3
"""Replace v2rayA in an authenticated OpenWrt IPK, retaining its service files."""
import argparse
import copy
import gzip
import hashlib
import io
from pathlib import Path
import re
import tarfile


def repack(data, replacements):
    stream = io.BytesIO()
    with tarfile.open(fileobj=io.BytesIO(data)) as source:
        with tarfile.open(fileobj=stream, mode="w", format=tarfile.GNU_FORMAT) as target:
            for item in source:
                member = copy.copy(item)
                content = source.extractfile(item).read() if item.isfile() else None
                if item.name in replacements:
                    content = replacements.pop(item.name)
                    member.size = len(content)
                target.addfile(member, None if content is None else io.BytesIO(content))
    if replacements:
        raise ValueError("missing package entries: " + repr(list(replacements)))
    return gzip.compress(stream.getvalue(), compresslevel=9, mtime=0)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--base", type=Path, required=True)
    parser.add_argument("--sha256", required=True)
    parser.add_argument("--binary", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    data = args.base.read_bytes()
    if hashlib.sha256(data).hexdigest() != args.sha256:
        raise ValueError("base IPK checksum mismatch")
    binary = args.binary.read_bytes()
    if binary[:6] != b"\x7fELF\x02\x01" or int.from_bytes(binary[18:20], "little") != 183:
        raise ValueError("expected a Linux ARM64 ELF binary")
    with tarfile.open(fileobj=io.BytesIO(data)) as outer:
        payload = outer.extractfile("./data.tar.gz").read()
        controls = outer.extractfile("./control.tar.gz").read()
    payload = repack(payload, {"./usr/bin/v2raya": binary})
    with tarfile.open(fileobj=io.BytesIO(payload)) as files:
        installed_size = sum(((f.size + 1023) // 1024) * 1024 for f in files if f.isfile())
    with tarfile.open(fileobj=io.BytesIO(controls)) as control:
        text = control.extractfile("./control").read().decode()
    if "Package: v2raya\n" not in text or "Architecture: aarch64_cortex-a53\n" not in text:
        raise ValueError("unexpected base package")
    updates = {
        "Version": "2.2.7.3-r4.failover3",
        "Source": "https://github.com/wywywywycloud/v2rayA/tree/fix/openwrt-subscription-failover",
        "Maintainer": "Mikhail Levin",
        "URL": "https://github.com/wywywywycloud/v2rayA",
        "Installed-Size": str(installed_size),
    }
    for key, value in updates.items():
        text = re.sub(r"^" + re.escape(key) + r": .*?$", key + ": " + value, text, flags=re.M)
    text += " Local build: probe subscription servers and recover monitored connections.\n"
    controls = repack(controls, {"./control": text.encode()})
    result = repack(data, {"./data.tar.gz": payload, "./control.tar.gz": controls})
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_bytes(result)
    print(hashlib.sha256(result).hexdigest(), args.output)


if __name__ == "__main__":
    main()
