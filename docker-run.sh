#!/bin/bash
docker stop xray 2>/dev/null
docker rm xray 2>/dev/null
docker run -d \
  --restart=always \
  --privileged \
  --network=host \
  --name xray \
  -e V2RAYA_ADDRESS=0.0.0.0:2017 \
  -v /lib/modules:/lib/modules:ro \
  -v /etc/resolv.conf:/etc/resolv.conf \
  -v /etc/v2raya:/etc/v2raya \
  v2raya
