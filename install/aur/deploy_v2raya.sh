#!/bin/bash
set -ex

if [ -z "$VERSION" ]; then
  echo "::error::VERSION is not set" >&2
  exit 1
fi

mkdir -p /tmp/prepare/v2raya
cd /tmp/prepare/v2raya
cp -r "$P_DIR"/install/aur/v2raya/. ./

sed -i s/{{pkgver}}/"$VERSION"/g PKGBUILD .SRCINFO

# If this exact pkgver is already on the AUR (re-release), bump pkgrel.
CURRENT="$(curl -fsSL "https://aur.archlinux.org/rpc/?v=5&type=info&arg[]=v2raya" 2>/dev/null | jq -r '.results[0].Version // empty' || true)"
if [ -n "$CURRENT" ]; then
  CUR_VER="${CURRENT%-*}"
  CUR_REL="${CURRENT##*-}"
  if [ "$CUR_VER" = "$VERSION" ] && [[ "$CUR_REL" =~ ^[0-9]+$ ]]; then
    NEW_REL=$((CUR_REL + 1))
    sed -i "s/^pkgrel=.*/pkgrel=$NEW_REL/" PKGBUILD
    sed -i "s/^	pkgrel = .*/	pkgrel = $NEW_REL/" .SRCINFO
    echo "pkgver $VERSION is already on AUR; bumping pkgrel to $NEW_REL"
  fi
fi

# Never push a template with unsubstituted placeholders.
if grep -q '{{' PKGBUILD .SRCINFO; then
  echo "::error::Unsubstituted placeholders left in PKGBUILD/.SRCINFO" >&2
  exit 1
fi

rm -rf /tmp/v2raya
git clone ssh://aur@aur.archlinux.org/v2raya.git /tmp/v2raya
cd /tmp/v2raya
cp -rf /tmp/prepare/v2raya/. ./
git add .
git commit -m "release $VERSION"
git push origin master
