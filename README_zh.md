# v2rayA [![Docker Cloud Build Status](https://img.shields.io/docker/cloud/build/mzz2017/v2raya)](https://hub.docker.com/r/mzz2017/v2raya) [![Travis (.org)](https://img.shields.io/travis/mzz2017/v2rayA?label=travis-ci%20build)](https://travis-ci.org/mzz2017/v2rayA) [![Netlify](https://img.shields.io/netlify/f93dbfa7-d14b-4968-a7a4-5e503d8bf5e5?label=netlify%20build)](https://app.netlify.com/sites/xenodochial-jepsen-122e9b/deploys)

[**English**](https://github.com/mzz2017/v2rayA/blob/master/README.md)&nbsp;&nbsp;&nbsp;[**简体中文**](https://github.com/mzz2017/v2rayA/blob/master/README_zh.md)

v2rayA 是一个支持全局透明代理的 V2Ray Linux 客户端，同时兼容SS、SSR、[Trojan](https://github.com/trojan-gfw/trojan)、[PingTunnel](https://github.com/esrrhs/pingtunnel)协议。 [[SS/SSR支持清单]](https://github.com/mzz2017/shadowsocksR/blob/master/README.md#ss-encrypting-algorithm)

v2rayA 致力于提供最简单的操作，满足绝大部分需求。

得益于Web客户端的优势，你不仅可以将其用于本地计算机，还可以轻松地将它部署在路由器或NAS上。

项目地址：https://github.com/mzz2017/v2rayA

前端 demo: https://v2raya.mzz.pub


## 使用方法

v2rayA主要提供了下述使用方法：

1. 软件源安装
2. docker
3. 二进制文件、安装包

详见 [**v2rayA - Wiki**](https://github.com/mzz2017/v2rayA/wiki/使用方法)


## 界面截图

<img src="https://i.loli.net/2020/04/19/kp2oedPiSzVwgHJ.png" border="0">


## 注意

1. 程序不会将任何用户数据保存在云端，所有用户数据存放在用户本地配置文件中。若服务端运行于 docker，则当相应 docker volume 被清除时配置也将随之消失，请做好备份。

2. 提供的[GUI demo](https://v2raya.mzz.pub)是由[Netlify](https://app.netlify.com/)在本 Github 项目自动部署完成的，如果担心安全性可以[自行部署](https://github.com/mzz2017/v2rayA/wiki/%E9%83%A8%E7%BD%B2GUI)。

3. **不要将本项目用于不合法用途。**

## 感谢

[hq450/fancyss](https://github.com/hq450/fancyss)

[ToutyRater/v2ray-guide](https://github.com/ToutyRater/v2ray-guide/blob/master/routing/sitedata.md)

[nadoo/glider](https://github.com/nadoo/glider)

[Loyalsoldier/v2ray-rules-dat](https://github.com/Loyalsoldier/v2ray-rules-dat)

## 协议

[![License: GPL v3](https://img.shields.io/badge/License-GPL%20v3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
