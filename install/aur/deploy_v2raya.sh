#!/bin/bash
set -ex

mkdir -p /tmp/prepare/v2raya
cd /tmp/prepare/v2raya
cp -r "$P_DIR"/install/aur/v2raya/. ./

sed -i s/{{pkgver}}/"$VERSION"/g PKGBUILD .SRCINFO

cd /tmp/
git clone ssh://aur@aur.archlinux.org/v2raya.git
cd v2raya
cp -rf /tmp/prepare/v2raya/. ./
git add .
git commit -m "release $VERSION"
git push -u -f origin master
cd $P_DIR
