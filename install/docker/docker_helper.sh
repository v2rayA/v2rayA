#!/bin/sh
# The image installs the release's own Alpine package, so every release
# installs its apk at least once before anyone else does. The release
# workflow substitutes the version.
set -ex
case "$(arch)" in
    x86_64)
        v2raya_arch="x64"
        ;;
    armv7l)
        v2raya_arch="armv7"
        ;;
    aarch64)
        v2raya_arch="arm64"
        ;;
    riscv64)
        v2raya_arch="riscv64"
        ;;
    *)
        echo "no Alpine package for $(arch)" >&2
        exit 1
        ;;
esac
package="installer_alpine_${v2raya_arch}_Realv2rayAVersion.apk"
wget -O "/tmp/$package" "https://github.com/v2rayA/v2rayA/releases/download/vRealv2rayAVersion/$package"
apk add --no-cache --allow-untrusted "/tmp/$package" iproute2 iptables iptables-legacy nftables tzdata
rm "/tmp/$package"
# GFWList mode reads LoyalsoldierSite.dat; the image has always shipped it as
# a copy of geosite.dat, which the first GFWList update replaces.
cp /usr/share/v2raya/geosite.dat /usr/share/v2raya/LoyalsoldierSite.dat
install ./iptables.sh /usr/local/bin/iptables
install ./ip6tables.sh /usr/local/bin/ip6tables
install ./iptables.sh /usr/local/bin/iptables-nft
install ./ip6tables.sh /usr/local/bin/ip6tables-nft
install ./iptables.sh /usr/local/bin/iptables-legacy
install ./ip6tables.sh /usr/local/bin/ip6tables-legacy
