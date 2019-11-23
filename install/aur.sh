#!/bin/bash
eval $(ssh-agent)
chmod 600 ./install/deploy_key
./install/ssh-add_expect ./install/deploy_key
ssh-keyscan -H aur.archlinux.org >>~/.ssh/known_hosts
git config user.name "$(git show -s --format='%an')"
git config user.email "$(git show -s --format='%ae')"
cp -f install/PKGBUILD install/.SRCINFO install/.INSTALL /tmp/
cd /tmp/
git clone ssh://aur@aur.archlinux.org/v2raya.git
sudo cp -f /tmp/PKGBUILD /tmp/.SRCINFO /tmp/.INSTALL /tmp/v2raya/
cd /tmp/v2raya
sed -i s/{{pkgver}}/${VERSION:1}/g PKGBUILD
sed -i s/{{pkgver}}/${VERSION:1}/g .SRCINFO
git add .
git commit -m "AUR"
git push -f origin master
cd $bp #回项目目录
echo "ok"
