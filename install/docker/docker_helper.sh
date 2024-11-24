#!/bin/sh
set -x
current_dir=$(pwd)
case "$(arch)" in
    x86_64)
        v2ray_arch="64"
        v2raya_arch="x64"
        ;;
    armv7l)
        v2ray_arch="arm32-v7a"
        v2raya_arch="armv7"
        ;;
    aarch64)
        v2ray_arch="arm64-v8a"
        v2raya_arch="arm64"
        ;;
    riscv64)
        v2ray_arch="riscv64"
        v2raya_arch="riscv64"
        ;;
    *)
        ;;
esac
mkdir -p build && cd build || exit
wget https://github.com/v2fly/v2ray-core/releases/latest/download/v2ray-linux-$v2ray_arch.zip
wget https://github.com/XTLS/Xray-core/releases/latest/download/Xray-linux-$v2ray_arch.zip
wget https://github.com/v2rayA/v2rayA/releases/download/vRealv2rayAVersion/v2raya_linux_"$v2raya_arch"_Realv2rayAVersion
unzip v2ray-linux-"$v2ray_arch".zip -d v2ray
install ./v2ray/v2ray /usr/local/bin/v2ray
unzip Xray-linux-"$v2ray_arch".zip -d xray
install ./xray/xray /usr/local/bin/xray
install ./v2raya_linux_"$v2raya_arch"_Realv2rayAVersion /usr/bin/v2raya
mkdir /usr/local/share/v2raya
ln -s /usr/local/share/v2raya /usr/local/share/v2ray
ln -s /usr/local/share/v2raya /usr/local/share/xray
wget -O /usr/local/share/v2raya/LoyalsoldierSite.dat https://raw.githubusercontent.com/mzz2017/dist-v2ray-rules-dat/master/geosite.dat
wget -O /usr/local/share/v2raya/geosite.dat https://raw.githubusercontent.com/mzz2017/dist-v2ray-rules-dat/master/geosite.dat
wget -O /usr/local/share/v2raya/geoip.dat https://raw.githubusercontent.com/mzz2017/dist-v2ray-rules-dat/master/geoip.dat
cd "$current_dir" || exit
rm -rf build
apk add --no-cache iptables iptables-legacy nftables tzdata
install ./iptables.sh /usr/local/bin/iptables
install ./ip6tables.sh /usr/local/bin/ip6tables
install ./iptables.sh /usr/local/bin/iptables-nft
install ./ip6tables.sh /usr/local/bin/ip6tables-nft
install ./iptables.sh /usr/local/bin/iptables-legacy
install ./ip6tables.sh /usr/local/bin/ip6tables-legacy