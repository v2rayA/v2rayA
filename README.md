# V2RayA

V2RayA是V2Ray的一个Web客户端。

尽管V2Ray的客户端很多，但在Linux上好用的却寥寥无几。[jiangxufeng/v2rayL](https://github.com/jiangxufeng/v2rayL)是目前Linux上较好的一个客户端，但暂时无法满足我对用户体验的较高要求，决定手撸一个。

虽然V2RayA是一个Web客户端，但它支持以PWA(Progressive Web App)的方式享受桌面端应用的体验。[食用方法](https://www.ithome.com/0/414/429.htm)

项目地址：https://github.com/mzz2017/V2RayA

## Status

### Build Status

| name   | docker image                                                 | travis-ci                                                    |
| ------ | ------------------------------------------------------------ | ------------------------------------------------------------ |
| V2RayA | [![Docker Cloud Build Status](https://img.shields.io/docker/cloud/build/mzz2017/v2raya?style=for-the-badge)](https://hub.docker.com/r/mzz2017/v2raya) | [![Travis (.org)](https://img.shields.io/travis/mzz2017/V2RayA?label=TRAVIS-CI%20BUILD&style=for-the-badge)](https://travis-ci.org/mzz2017/V2RayA) |

### Usage Statistics

| name   | docker pulls                                                 | github releases                                              |
| ------ | ------------------------------------------------------------ | ------------------------------------------------------------ |
| V2RayA | [![Docker Pulls](https://img.shields.io/docker/pulls/mzz2017/v2raya?style=for-the-badge)](https://hub.docker.com/r/mzz2017/v2raya) | [![GitHub All Releases](https://img.shields.io/github/downloads/mzz2017/V2RayA/total?label=releases%20downloads&style=for-the-badge)](https://github.com/mzz2017/V2RayA/releases) |

## 功能清单

**目前支持订阅、导入等必要功能，暂不支持手动录入节点，项目正在逐步开发中。**

已支持：

- [x] V2Ray服务控制
- [x] 导入vmess、ss、订阅地址
- [x] websocket、kcp、tcp、http、tls、shadowsocks全支持
- [x] 测试节点Ping时延
- [x] 二维码、地址分享
- [x] 自定义PAC模式（GFWList、大陆白名单、自定义路由规则）

待开发：

- [ ] 测试节点HTTP时延
- [ ] 自动更新订阅、PAC文件
- [ ] 多路复用、TCPFastOpen
- [ ] 手动添加/修改节点
- [ ] 登陆与安全

## 界面截图

<img src="http://mzzeast.shumsg.cn/FtwssiGjyR_IXalEiquQw--5ChYl" />
<p align="center">节点</p>
<img src="http://mzzeast.shumsg.cn/FlF9m8Ze5D24FlS0DfYykKCG0G3-" />
<p align="center">订阅源</p>
<details>
    <summary>点击展开更多截图</summary>
<img src="http://mzzeast.shumsg.cn/FgSf755x0pr6TF-Cgy21rF0AiKax" />
<p align="center">设置</p>
<img src="http://mzzeast.shumsg.cn/FiVwkK1H5PqTevGcVAp34GCOuERE" />
<p align="center">使用自定义PAC时可以配置路由规则</p>


</details>

## 使用

如下使用方法：

1. 拉取源码，在本地用docker-compose部署service，在[Web-GUI](https://v2raya.mzz.pub)使用（或手动部署Web-GUI）。
   
   ```bash
   git clone https://github.com/mzz2017/V2RayA.git
   cd V2RayA
   docker-compose up -d
   ```
   
2. 用docker拉取镜像部署service，在[Web-GUI](https://v2raya.mzz.pub)使用（或手动部署Web-GUI）。

   ```bash
   docker pull mzz2017/v2raya
   docker run -d -p 2017:2017 -p 10800-10802:10800-10802 --restart=always mzz2017/v2raya
   ```

3. 【不推荐】不使用docker

   ```bash
   git clone https://github.com/mzz2017/V2RayA.git
   cd V2RayA/service
   sudo go run -mod=vendor main.go
   ```

   或直接使用[Releases](https://github.com/mzz2017/V2RayA/releases)。注意，该方法不支持windows，其他系统请使用sudo运行。

默认使用的四个端口分别为：

2017: V2RayA后端端口

10800: SOCKS协议

10801: HTTP协议

10802: 带PAC的HTTP协议

用户可通过docker将上述端口映射到本地的任意端口。

## 注意

应用不会将任何用户数据保存在云端，所有用户数据存放在用户本地的docker容器中，当docker容器被清除时配置也将随之消失。

提供的[GUI demo](https://v2raya.mzz.pub)是由[Render](https://render.com/)在本Github项目自动部署完成的，如果担心安全性可以自行部署。

不要将本项目用于不合法用途，作者仅将该项目用于学习研究和内网穿透的用途。

# 感谢

[jiangxufeng/v2rayL](https://github.com/jiangxufeng/v2rayL)

[2dust/v2rayN](https://github.com/2dust/v2rayN)

[hq450/fancyss](https://github.com/hq450/fancyss)

[xlzd/quickdown](https://github.com/xlzd/quickdown)

[ToutyRater/v2ray-guide](https://github.com/ToutyRater/v2ray-guide/blob/master/routing/sitedata.md)

# 相似项目

[v2raywebui/V2RayWebUI](https://github.com/v2raywebui/V2RayWebUI)

[jiangxufeng/v2rayL](https://github.com/jiangxufeng/v2rayL)

[NoOne-hub/v2ray_client](https://github.com/NoOne-hub/v2ray_client)

# 协议

[![License: GPL v3](https://img.shields.io/badge/License-GPL%20v3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)