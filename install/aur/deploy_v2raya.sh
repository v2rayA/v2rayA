#!/bin/bash
mkdir -p /tmp/prepare/v2raya
cd /tmp/prepare/v2raya
cp "$srcdir"/install/aur/v2raya/* ./
cp "$srcdir"/install/aur/v2raya/.* ./

sed -i s/{{pkgver}}/"${VERSION:1}"/g PKGBUILD .SRCINFO

cd /tmp/
git clone ssh://aur@aur.archlinux.org/v2raya.git
cd v2raya
cp -f /tmp/prepare/v2raya/* ./
cp -f /tmp/prepare/v2raya/.* ./
git add .
git commit -m "release $VERSION"
git push -u -f origin master
cd $srcdir #回项目目录
