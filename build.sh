#!/bin/bash
set -ex
SHELL_FOLDER=$(dirname "$0")
cd "$SHELL_FOLDER"/gui && yarn && OUTPUT_DIR="$SHELL_FOLDER"/service/server/router/web yarn build
cd "$SHELL_FOLDER"/service && go build -o "$SHELL_FOLDER"/v2raya