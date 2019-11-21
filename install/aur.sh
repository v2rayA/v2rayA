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
cp -r PKGBUILD .SRCINFO v2raya
cd v2raya && git add . && git commit -m "AUR" && git push origin master
