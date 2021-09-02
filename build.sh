#!/bin/bash
set -ex
SHELL_FOLDER="$(pwd)"/"$(dirname $0)"
cd "$SHELL_FOLDER"/gui && yarn && OUTPUT_DIR="$SHELL_FOLDER"/service/server/router/web yarn build
cd "$SHELL_FOLDER"/service && CGO_ENABLED=0 go build -ldflags "-s -w" -o "$SHELL_FOLDER"/v2raya
