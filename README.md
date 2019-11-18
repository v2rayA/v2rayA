# V2RayA

V2RayA是V2Ray的一个Web客户端。

尽管V2Ray的客户端很多，但在Linux上好用的却寥寥无几。jiangxufeng/v2rayL是目前Linux上较好的一个客户端，但暂时无法满足我对用户体验的较高要求，决定手撸一个。

虽然V2RayA是一个Web客户端，但它支持以PWA(Progressive Web App)的方式享受桌面端应用的体验。[食用方法](https://www.ithome.com/0/414/429.htm)

项目地址：https://github.com/mzz2017/V2RayA

前端demo: https://v2raya.mzz.pub

## Build Status

| name   | docker image                                                                                                                                          | travis-ci                                                                                                                                          |
| ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| V2RayA | [![Docker Cloud Build Status](https://img.shields.io/docker/cloud/build/mzz2017/v2raya?style=for-the-badge)](https://hub.docker.com/r/mzz2017/v2raya) | [![Travis (.org)](https://img.shields.io/travis/mzz2017/V2RayA?label=TRAVIS-CI%20BUILD&style=for-the-badge)](https://travis-ci.org/mzz2017/V2RayA) |

## 功能清单

**目前已完成基本功能，项目正在逐步开发中。**

已支持：

- [x] 全局透明代理（不支持docker模式）
- [x] V2Ray服务控制
- [x] 导入vmess、ss、订阅地址
- [x] websocket、kcp、tcp、http、tls、shadowsocks全支持
- [x] 测试节点Ping时延
- [x] 二维码、地址分享
- [x] 同时开放socks5、http、带PAC的http三个端口（PAC模式可选GFWList、大陆白名单）
- [x] 服务端启动自动检查PAC、订阅更新
- [x] 多路复用、TCPFastOpen支持

待开发：

- [ ] 手动添加/修改节点
- [ ] 测试节点HTTP时延
- [ ] 自定义PAC路由规则
- [ ] QUIC协议支持
- [ ] 端口号配置
- [ ] 登陆与安全

## 界面截图

<img src="http://mzzeast.shumsg.cn/FtwssiGjyR_IXalEiquQw--5ChYl" />
<p align="center">节点</p>
<img src="http://mzzeast.shumsg.cn/FlF9m8Ze5D24FlS0DfYykKCG0G3-" />
<p align="center">订阅源</p>
<details>
    <summary>点击展开更多截图</summary>
<img src="http://mzzeast.shumsg.cn/Ft6KlgZuuMNsL5oCHxfkBllEFvuf" />
<p align="center">设置</p>
<img src="http://mzzeast.shumsg.cn/FiVwkK1H5PqTevGcVAp34GCOuERE" />
<p align="center">使用自定义PAC时可以配置路由规则</p>



</details>

## 使用

如下使用方法：

1. 拉取源码，使用docker-compose部署，在[Web-GUI](https://v2raya.mzz.pub)使用（或手动部署Web-GUI）。
   
   ```bash
   git clone https://github.com/mzz2017/V2RayA.git
   cd V2RayA
   docker-compose up -d --build
   ```

2. 使用docker命令部署，在[Web-GUI](https://v2raya.mzz.pub)使用（或手动部署Web-GUI）。

   我们同步发行[Docker镜像](https://hub.docker.com/r/mzz2017/v2raya)，如果无法使用docker-compose，可以参考[docker-compose.yml](https://github.com/mzz2017/V2RayA/blob/master/docker-compose.yml)并使用docker命令自行搭建。
   
3. 不使用docker

   如不使用docker，请确保已正确安装V2Ray及[RedisJSON](https://github.com/RedisLabsModules/redisjson)

   准备完毕后，可下载[Releases](https://github.com/mzz2017/V2RayA/releases)启动V2RayA服务端
   
   或拉取源码，通过golang启动

   ```bash
   git clone https://github.com/mzz2017/V2RayA.git
   cd V2RayA/service
   sudo go run -mod=vendor main.go
   ```

   注意，尽管golang具有交叉编译的特性，但由于项目使用了linux commands，导致该方法不支持windows。若想在windows体验，可借助Docker或WSL。

默认使用的四个端口分别为：

2017: V2RayA后端端口

20170: SOCKS协议

20171: HTTP协议

20172: 带PAC的HTTP协议


### 在不同运行环境下程序表现将不同

由于docker容器对systemd的限制性，在docker中将采用pid共享进程命名空间，volumes共享存储空间，更新配置后通过结束进程触发v2ray容器的重启来更新配置，以无inbounds的配置代替断开连接，这是一种折中方案，会有如下影响：

1. 在更换配置时略有卡顿
2. 无法使用全局透明代理

在宿主环境下以sudo权限运行将不受此限制。

### 支持Windows、MacOS吗

目前仅在Linux进行过测试，并计划优先适配Linux。目前尚未验证在Windows及MacOS上存在的问题。

实际上Windows和MacOS上已经存在很多优秀的V2Ray客户端，若无特殊需求，建议选择这些客户端。

### 已知问题

+ 在windows环境下无法监听2017端口。据了解，Caddy的2015端口也是不能监听的，后期将考虑支持自定义端口以解决问题。
+ 在使用GoLand进行开发时，如果开启了全局透明代理，由于程序捕获不了GoLand的结束signal，在程序退出后将无法恢复正常网络，因此建议使用`killall ___go_build_V2R`来结束进程。如已中招重启系统即可。


## 注意

应用不会将任何用户数据保存在云端，所有用户数据存放在用户本地配置文件中。若服务端运行于docker，则当docker容器被清除时配置也将随之消失，请做好备份。

提供的[GUI demo](https://v2raya.mzz.pub)是由[Render](https://render.com/)在本Github项目自动部署完成的，如果担心安全性可以自行部署。

不要将本项目用于不合法用途，作者仅将该项目用于学习研究和内网穿透的用途。

# 在docker环境中开发

```bash
docker-compose -f docker-compose.dev.yml up
```

gin会监测文件改动并热重载，见[codegangsta/gin](https://github.com/codegangsta/gin)。

如果出现问题，可以尝试添加`--build`参数，或将docker离开swarm模式：`docker swarm leave --force`。

# 感谢

[hq450/fancyss](https://github.com/hq450/fancyss)

[xlzd/quickdown](https://github.com/xlzd/quickdown)

[ToutyRater/v2ray-guide](https://github.com/ToutyRater/v2ray-guide/blob/master/routing/sitedata.md)

[txthinking/brook](https://github.com/txthinking/brook)

# 相似项目

[v2raywebui/V2RayWebUI](https://github.com/v2raywebui/V2RayWebUI)

[jiangxufeng/v2rayL](https://github.com/jiangxufeng/v2rayL)

[NoOne-hub/v2ray_client](https://github.com/NoOne-hub/v2ray_client)

# 协议

[![License: GPL v3](https://img.shields.io/badge/License-GPL%20v3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)