#!/bin/bash
eval $(ssh-agent)
chmod 600 ./install/aur/deploy_key
./install/tool/ssh-add_expect ./install/aur/deploy_key
ssh-keyscan -H aur.archlinux.org >>~/.ssh/known_hosts
git config --global user.name "$(git show -s --format='%an')"
git config --global user.email "$(git show -s --format='%ae')"
bash ./install/aur/deploy_v2raya.sh
bash ./install/aur/deploy_v2raya_bin.sh
echo "ok"
cd $srcdir #回项目目录
