#!/bin/bash

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
# https://github.com/webpack/webpack/issues/14532#issuecomment-947012063
cd "$CurrentDir"/gui && yarn && OUTPUT_DIR="$CurrentDir"/service/server/router/web yarn build
find "$CurrentDir"/service/server/router/web \! -name \*.png -a \! -name \*.gz -a \! -name index.html -a ! -type d -exec gzip -9 {} +
cd "$CurrentDir"/service && CGO_ENABLED=0 go build -ldflags "-X github.com/v2rayA/v2rayA/conf.Version=$version -s -w" -o "$CurrentDir"/v2raya
