#!/bin/bash
cd $srcdir
cp -rf install/aur/v2raya /tmp/v2raya
cd /tmp/v2raya
git init
sed -i s/{{pkgver}}/${VERSION:1}/g PKGBUILD
sed -i s/{{pkgver}}/${VERSION:1}/g .SRCINFO
git add .
git commit -m "release $VERSION"
git remote add origin "ssh://aur@aur.archlinux.org/v2raya.git"
git push -u -f origin master
cd $srcdir #回项目目录
