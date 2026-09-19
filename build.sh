#!/usr/bin/env bash

set -euo pipefail

root_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
version="${1:-$(git -C "$root_dir" describe --tags --always --dirty)}"
web_dir="$root_dir/service/server/router/web"

yarn --cwd "$root_dir/gui" --check-files
yarn --cwd "$root_dir/gui" build

rm -rf "$web_dir"
mkdir -p "$web_dir"
cp -a "$root_dir/web/." "$web_dir/"

(
  cd "$root_dir/core"
  CGO_ENABLED=0 go build -trimpath \
    -o "$root_dir/v2raya_core" \
    -ldflags="-X main.Version=$version -s -w" \
    ./main
)

(
  cd "$root_dir/service"
  CGO_ENABLED=0 go build -trimpath \
    -o "$root_dir/v2raya" \
    -ldflags="-X github.com/v2rayA/v2rayA/conf.Version=$version -s -w"
)
