#!/bin/bash

set -ex
P_DIR=$(pwd)
cd $P_DIR/gui && yarn && yarn build
rm -rf $P_DIR/service/server/router/web || true
cd $P_DIR && mv web $P_DIR/service/server/router/web
cd $P_DIR/service && go build -o ../v2raya