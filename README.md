# V2RayA

V2RayA 是一个支持全局透明代理的 V2Ray Linux 客户端，同时兼容SS、SSR协议。

V2RayA 致力于提供最简单的操作，满足绝大部分需求。

同时兼容V2Ray、SS、SSR意味着用户不再需要在不同工具之间切换，你甚至可以使用一个混合协议的订阅。

虽然 V2RayA 是一个 Web 客户端，它也支持以 PWA(Progressive Web App) 的方式享受桌面端应用的体验。[食用方法](https://www.ithome.com/0/414/429.htm)

得益于Web客户端的优势，你不仅可以将其用于本地计算机，还可以轻松地将它部署在路由器或NAS上。

目前V2RayA仅在部分Linux发行版进行过充分测试，在使用过程中如果遇到问题，欢迎提出issue和PR。

项目地址：https://github.com/mzz2017/V2RayA

前端 demo: https://v2raya.mzz.pub

## Build Status

| name   | docker image                                                 | travis-ci                                                    | netlify                                                      |
| ------ | ------------------------------------------------------------ | ------------------------------------------------------------ | ------------------------------------------------------------ |
| V2RayA | [![Docker Cloud Build Status](https://img.shields.io/docker/cloud/build/mzz2017/v2raya?style=for-the-badge)](https://hub.docker.com/r/mzz2017/v2raya) | [![Travis (.org)](https://img.shields.io/travis/mzz2017/V2RayA?label=TRAVIS-CI%20BUILD&style=for-the-badge)](https://travis-ci.org/mzz2017/V2RayA) | [![Netlify](https://img.shields.io/netlify/f93dbfa7-d14b-4968-a7a4-5e503d8bf5e5?label=netlify%20build&style=for-the-badge)](https://app.netlify.com/sites/xenodochial-jepsen-122e9b/deploys) |

## 功能清单

已支持：

- [x] 全局透明代理
- [x] V2Ray 服务控制
- [x] 导入 vmess、ss、ssr、订阅地址
- [x] 手动添加/修改节点
- [x] websocket、kcp、tcp、http、tls、shadowsocks、shadowsocksR 全支持 [[SS/SSR支持清单]](https://github.com/mzz2017/shadowsocksR/blob/master/README.md#ss-encrypting-algorithm)
- [x] 测试节点 Ping、HTTP 时延
- [x] 二维码、地址分享
- [x] 支持PAC
- [x] 服务端启动自动检查 PAC、订阅更新
- [x] 多路复用、TCPFastOpen 支持
- [x] 自动检查版本更新
- [x] 自定义端口
- [x] 自定义路由规则

待开发：

- [ ] RoutingA
- [ ] 回国模式（当前版本可通过自定义路由规则实现）
- [ ] QUIC、auth_chain\*支持
- [ ] 日志
- [ ] 多语言

## 界面截图

<img src="https://s2.ax1x.com/2020/02/03/1wfTbt.png" alt="1wfTbt.png" border="0">

<p align="center">节点</p>
<details>
    <summary>点击展开更多截图</summary>


<img src="https://s2.ax1x.com/2020/02/03/1wf4vd.png" alt="1wf4vd.png" border="0">

<p align="center">订阅源</p>
<img src="https://s2.ax1x.com/2020/02/03/1wfoDI.png" alt="1wfoDI.png" border="0">

<p align="center">设置</p>
<img src="https://s2.ax1x.com/2020/02/03/1wfIKA.png" alt="1wfIKA.png" border="0">

<p align="center">自定义路由规则</p>


</details>

# 使用方法

V2RayA主要提供了下述使用方法：

1. 软件源安装
2. docker
3. 二进制文件、安装包

详见 [**V2RayA - Wiki**](https://github.com/mzz2017/V2RayA/wiki/使用方法)


# 注意

1. 程序不会将任何用户数据保存在云端，所有用户数据存放在用户本地配置文件中。若服务端运行于 docker，则当相应 docker volume 被清除时配置也将随之消失，请做好备份。

2. 提供的[GUI demo](https://v2raya.mzz.pub)是由[Netlify](https://app.netlify.com/)在本 Github 项目自动部署完成的，如果担心安全性可以自行部署。

3. **不要将本项目用于不合法用途。**

# 感谢

[hq450/fancyss](https://github.com/hq450/fancyss)

[ToutyRater/v2ray-guide](https://github.com/ToutyRater/v2ray-guide/blob/master/routing/sitedata.md)

[nadoo/glider](https://github.com/nadoo/glider)

# 协议

[![License: GPL v3](https://img.shields.io/badge/License-GPL%20v3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
