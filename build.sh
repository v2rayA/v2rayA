#!/bin/bash
set -ex
SHELL_FOLDER="$(pwd)"/"$(dirname $0)"
if [ -d "$SHELL_FOLDER/.git" ]; then
  date=$(git log -1 --format="%cd" --date=short | sed s/-//g)
  count=$(git rev-list --count HEAD)
  commit=$(git rev-parse --short HEAD)
  version="unstable-$date.r${count}.$commit"
else
  version="unstable"
fi
cd "$SHELL_FOLDER"/gui && yarn && OUTPUT_DIR="$SHELL_FOLDER"/service/server/router/web yarn build
cd "$SHELL_FOLDER"/service && CGO_ENABLED=0 go build -ldflags "-X github.com/v2rayA/v2rayA/conf.Version=$version -s -w" -o "$SHELL_FOLDER"/v2raya
