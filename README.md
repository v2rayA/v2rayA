# V2RayA [![Docker Cloud Build Status](https://img.shields.io/docker/cloud/build/mzz2017/v2raya)](https://hub.docker.com/r/mzz2017/v2raya) [![Travis (.org)](https://img.shields.io/travis/mzz2017/V2RayA?label=travis-ci%20build)](https://travis-ci.org/mzz2017/V2RayA) [![Netlify](https://img.shields.io/netlify/f93dbfa7-d14b-4968-a7a4-5e503d8bf5e5?label=netlify%20build)](https://app.netlify.com/sites/xenodochial-jepsen-122e9b/deploys)

[**English**](https://github.com/mzz2017/V2RayA/blob/master/README.md)&nbsp;&nbsp;&nbsp;[**简体中文**](https://github.com/mzz2017/V2RayA/blob/master/README_zh.md)

V2RayA is a V2Ray Linux client supporting global transparent proxy, compatible with SS, SSR, [Trojan](https://github.com/trojan-gfw/trojan), [PingTunnel](https://github.com/esrrhs/pingtunnel) protocols. [[SS/SSR protocol list]](https://github.com/mzz2017/shadowsocksR/blob/master/README.md#ss-encrypting-algorithm)

directlyWe are committed to providing the simplest operation and meet most needs.

Thanks to the advantages of Web GUI, you can not only use it on your local computer, but also easily deploy it on a router or NAS.

Project：https://github.com/mzz2017/V2RayA

Frontend demo: https://v2raya.mzz.pub


## Usage

V2RayA mainly provides the following methods of installation:

1. Install from apt-source or AUR
2. Docker
3. Binary file and installation package

See [**V2RayA - Wiki**](https://github.com/mzz2017/V2RayA/wiki/Usage)


## Screenshot

<img src="https://i.loli.net/2020/04/19/gt3NqOMiafYbp7L.png" border="0">

## Statement

1. The program does not save any user data in the cloud, all user data is stored in local. If the v2raya service is running in docker, the configuration will disappear when the related docker volume is removed. Please make a backup if necessary.
2. The provided [GUI demo](https://v2raya.mzz.pub) is automatically deployed by [Netlify](https://app.netlify.com/). If you are worried about security, you can [deploy it yourself](https://github.com/mzz2017/V2RayA/wiki/Deploy-GUI).
3. **Do not use this project for illegal purposes.**

## Credits

[hq450/fancyss](https://github.com/hq450/fancyss)

[ToutyRater/v2ray-guide](https://github.com/ToutyRater/v2ray-guide/blob/master/routing/sitedata.md)

[nadoo/glider](https://github.com/nadoo/glider)

[Loyalsoldier/v2ray-rules-dat](https://github.com/Loyalsoldier/v2ray-rules-dat)

## License

[![License: GPL v3](https://img.shields.io/badge/License-GPL%20v3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
