# 将已发布于github-release的最新版本使用fpm打包为deb, rpm, freebsd, pacman (ArchLinux) packages, Mac OS X .pkg files (osxpkg)
NAME=v2raya
VERSION=$(git describe --abbrev=0 --tags)

.PHONY: package
package:
	rm -f v2raya
	wget https://github.com/mzz2017/V2RayA/releases/download/$VERSION/V2RayA_linux_amd64 -o v2raya
	fpm -s dir -t deb -n $NAME -v ${VERSION:1} --prefix /usr/bin v2raya
	fpm -s dir -t rpm -n $NAME -v ${VERSION:1} --prefix /usr/bin v2raya
	fpm -s dir -t freebsd -n $NAME -v ${VERSION:1} --prefix /usr/bin v2raya
	wget https://github.com/mzz2017/V2RayA/releases/download/$VERSION/V2RayA_darwin_amd64 -o v2raya
	fpm -s dir -t osxpkg -n $NAME -v ${VERSION:1} --prefix /usr/bin v2raya