#!/bin/bash
set -ex

# Git identity for the AUR commits. Override with AUR_GIT_NAME / AUR_GIT_EMAIL.
GIT_NAME="${AUR_GIT_NAME:-$(git show -s --format='%an')}"
GIT_EMAIL="${AUR_GIT_EMAIL:-$(git show -s --format='%ae')}"
git config --global user.name "$GIT_NAME"
git config --global user.email "$GIT_EMAIL"

mkdir -p ~/.ssh
ssh-keyscan -H aur.archlinux.org >>~/.ssh/known_hosts 2>/dev/null || true

# Fresh working copies, so the workflow can be re-run safely.
rm -rf /tmp/prepare /tmp/v2raya /tmp/v2raya-bin

bash ./install/aur/deploy_v2raya.sh
bash ./install/aur/deploy_v2raya_bin.sh
echo "ok"
