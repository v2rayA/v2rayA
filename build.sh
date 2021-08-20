#!/bin/bash

set -ex
P_DIR=$(pwd)
cd $P_DIR/gui && yarn && OUTPUT_DIR=$P_DIR/service/server/router/web yarn build
cd $P_DIR/service && go build -o ../v2raya