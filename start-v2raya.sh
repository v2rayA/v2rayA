#!/bin/bash

l_container_name=v2raya

l_count=$(docker ps -a --filter "name=$l_container_name" |grep -c "$l_container_name")

if [ $l_count -gt 0 ]; then
	echo 停止容器$l_container_name
	docker rm -f $l_container_name
fi

echo 启动容器$l_container_name

# 启动V2rayA

docker run -d \
  --restart=always \
  --privileged \
  --network=host \
  --name $l_container_name \
  -e V2RAYA_LOG_FILE=/tmp/logs/v2raya.log \
  -e V2RAYA_V2RAY_BIN=/usr/bin/v2ray \
  -e V2RAYA_NFTABLES_SUPPORT=on \
  -e IPTABLES_MODE=nftables \
  -v /lib/modules:/lib/modules:ro \
  -v /etc/resolv.conf:/etc/resolv.conf \
  -v /etc/v2raya:/etc/v2raya \
  -v /storage/v2rayA/logs:/tmp/logs \
  -v /storage/geodata:/root/.local/share/v2ray \
  -v /etc/localtime:/etc/localtime \
  registry.cn-hangzhou.aliyuncs.com/mosaicwang/v2raya:2.2.5.8