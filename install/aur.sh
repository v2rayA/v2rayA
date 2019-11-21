#!/bin/bash
cp -r install/PKGBUILD install/.SRCINFO ./
sed -i s/{{pkgver}}/${VERSION:1}/g PKGBUILD
sed -i s/{{pkgver}}/${VERSION:1}/g .SRCINFO
eval $(ssh-agent)
chmod 600 ./install/deploy_key
./install/ssh-add_expect ./install/deploy_key
ssh-keyscan -H aur.archlinux.org >>~/.ssh/known_hosts
git config user.name "$(git show -s --format='%an')"
git config user.email "$(git show -s --format='%ae')"
git clone ssh://aur@aur.archlinux.org/v2raya.git
cp -rf PKGBUILD .SRCINFO v2raya
mv v2raya /tmp/v2raya #换个地方，不能让git仓库有包含关系
cd /tmp/v2raya && git add . && git commit -m "AUR" || git push -f origin master || echo "ok"
