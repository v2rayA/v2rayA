#!/usr/bin/env bash

set -ex
CurrentDir="$(pwd)"

if [ -d "$CurrentDir/.git" ]; then
  date=$(git -C "$CurrentDir" log -1 --format="%cd" --date=short | sed s/-//g)
  count=$(git -C "$CurrentDir" rev-list --count HEAD)
  commit=$(git -C "$CurrentDir" rev-parse --short HEAD)
  version="unstable-$date.r${count}.$commit"
else
  version="unstable"
fi

# Build ngui frontend with pnpm + nuxt generate
cd "$CurrentDir"/ngui && pnpm install --frozen-lockfile && OUTPUT_DIR="$CurrentDir"/service/server/router/web pnpm run generate

# Build v2raya-core (merged xray-core + custom protocols)
cd "$CurrentDir"/core && CGO_ENABLED=0 go build -ldflags "-X main.Version=$version -s -w" -o "$CurrentDir"/v2raya_core ./main

cd "$CurrentDir"/service && CGO_ENABLED=0 go build -tags "with_gvisor" -ldflags "-X github.com/v2rayA/v2rayA/conf.Version=$version -s -w" -o "$CurrentDir"/v2raya
