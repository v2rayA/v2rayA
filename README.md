# V2RayA

V2RayA 是一个支持全局透明代理的 V2Ray Linux 客户端。

V2RayA 致力于提供最简单的操作，满足绝大部分需求。

虽然 V2RayA 是一个 Web 客户端，它也支持以 PWA(Progressive Web App)的方式享受桌面端应用的体验。[食用方法](https://www.ithome.com/0/414/429.htm)

得益于Web客户端的优势，你可以轻松地将它部署在软路由或NAS上，并通过http访问管理你的节点。

目前V2RayA仅在部分Linux发行版进行过充分测试，在使用过程中如果遇到问题，欢迎提出issue。

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
- [x] 导入 vmess、ss、订阅地址
- [x] 手动添加/修改节点
- [x] websocket、kcp、tcp、http、tls、shadowsocks 全支持
- [x] 测试节点 Ping 时延
- [x] 二维码、地址分享
- [x] 同时开放 socks5、http、带 PAC 的 http 三个端口（PAC 模式可选 GFWList、大陆白名单）
- [x] 服务端启动自动检查 PAC、订阅更新
- [x] 多路复用、TCPFastOpen 支持
- [x] 以service方式启动
- [x] 检查版本更新
- [x] 登陆与安全
- [x] 服务端端口号配置、前端可指定服务端地址

待开发：

- [ ] 测试节点 HTTP 时延
- [ ] 自定义 PAC 路由规则
- [ ] QUIC、SSR 协议支持

## 界面截图

<img src="http://mzzeast.shumsg.cn/FtwssiGjyR_IXalEiquQw--5ChYl" />
<p align="center">节点</p>
<img src="http://mzzeast.shumsg.cn/FlF9m8Ze5D24FlS0DfYykKCG0G3-" />
<p align="center">订阅源</p>
<img src="http://mzzeast.shumsg.cn/FkOIrdEqCXvqQEgMH166RsugmaSs" />
<p align="center">设置</p>
<details>
    <summary>点击展开更多截图</summary>
<img src="http://mzzeast.shumsg.cn/FiVwkK1H5PqTevGcVAp34GCOuERE" />
<p align="center">使用自定义PAC时可以配置路由规则</p>

</details>

## 使用

如下使用方法：

1. **使用apt-get安装**（debian、ubuntu）
   
   请确保已正确安装 v2ray-core

   官方提供了 Linux 下的一键安装脚本：

   > 运行下面的指令下载并安装 V2Ray。当 yum 或 apt-get 可用的情况下，此脚本会自动安装 unzip 和 daemon。这两个组件是安装 V2Ray 的必要组件。如果你使用的系统不支持 yum 或 apt-get，请自行安装 unzip 和 daemon

   ```bash
   curl -L -s https://install.direct/go.sh | sudo -E bash -s - --source jsdelivr
   ```
   
   准备完毕后：

   ```bash
   # add public key
   wget -qO - https://cdn.jsdelivr.net/gh/mzz2017/V2RayA@apt/key/public-key.asc | sudo apt-key add -

   # add V2RayA's repository
   sudo add-apt-repository 'deb https://cdn.jsdelivr.net/gh/mzz2017/V2RayA@apt/ v2raya main'
   sudo apt-get update

   # install V2RayA
   sudo apt-get install v2raya
   ```
   
   V2RayA服务端正常运行后，就可在[Web-GUI](https://v2raya.mzz.pub)使用了（或手动部署 Web-GUI）。
2. **使用yay/yaourt安装**（archlinux、manjaro）
   
   由于v2raya发布在AUR中，而pacman不支持AUR，因此建议使用主流的yay或yaourt作为替代方案

   ```bash
   # install yay
   sudo pacman -Sy yay
   ```

   当yay或yaourt可用时，可通过yay或yaourt安装v2raya

   ```bash
   # assume command yay is available
   yay v2raya
   ```

   V2RayA服务端正常运行后，就可在[Web-GUI](https://v2raya.mzz.pub)使用了（或手动部署 Web-GUI）。
3. **使用二进制文件 / 安装包**（支持常见linux系统）

   请确保已正确安装 v2ray-core

   官方提供了 Linux 下的一键安装脚本：

   > 运行下面的指令下载并安装 V2Ray。当 yum 或 apt-get 可用的情况下，此脚本会自动安装 unzip 和 daemon。这两个组件是安装 V2Ray 的必要组件。如果你使用的系统不支持 yum 或 apt-get，请自行安装 unzip 和 daemon

   ```bash
   curl -L -s https://install.direct/go.sh | sudo -E bash -s - --source jsdelivr
   ```

   准备完毕后，可下载[Releases](https://github.com/mzz2017/V2RayA/releases)中的二进制文件启动V2RayA服务端，或下载安装包进行安装。
   
   V2RayA服务端正常运行后，就可在[Web-GUI](https://v2raya.mzz.pub)使用了（或手动部署 Web-GUI）。
4. 拉取源码，**使用 docker-compose 部署**。

   ```bash
   git clone https://github.com/mzz2017/V2RayA.git
   cd V2RayA
   docker-compose up -d --build
   ```
5. **使用 docker 命令部署**。注意在启动v2ray及v2raya容器时添加`--privileged --network host`以支持全局透明代理。

   我们同步发行[Docker 镜像](https://hub.docker.com/r/mzz2017/v2raya)，如果无法使用 docker-compose，可以参考[docker-compose.yml](https://github.com/mzz2017/V2RayA/blob/master/docker-compose.yml)并使用 docker 命令自行搭建。
6. 当然，你也可以选择拉取源码，**通过 golang 启动**：
   
   该方法同样需要正确安装v2ray-core，详情见上

   ```bash
   git clone https://github.com/mzz2017/V2RayA.git
   cd V2RayA/service
   export GOPROXY=https://goproxy.io # set goproxy.io as the proxy of go modules
   sudo go run main.go
   ```

   注意，尽管 golang 具有交叉编译的特性，但由于项目使用了 linux commands，导致该方法不支持 windows。若想在 windows 体验，可借助 Docker 或 WSL。

   

默认使用的四个端口分别为：

2017: V2RayA 后端端口

20170: SOCKS 协议

20171: HTTP 协议

20172: 带 PAC 的 HTTP 协议

其他端口：

12345: tproxy（全局透明代理所需）

### 在不同运行环境下程序表现将不同

由于 docker 容器对 systemd 的限制性，在 docker 中将采用 pid 共享进程命名空间，volumes 共享存储空间，更新配置后通过结束进程触发 v2ray 容器的重启来更新配置，以无 inbounds 的配置代替断开连接，这是一种折中方案，会有如下影响：

1. 在更换配置时略有卡顿

### 服务端支持 Windows、MacOSX 吗

目前仅在 Linux 进行过测试，并计划优先适配 Linux。目前尚未验证在 Windows 及 MacOSX 上存在的问题。

实际上 Windows 和 MacOSX 上已经存在很多优秀的 V2Ray 客户端，若无特殊需求，建议选择这些客户端。

### 已知问题

- 在使用 GoLand 进行开发调试时，**如果开启了全局透明代理**，由于进程捕获不了 GoLand 的结束 signal，在进程退出后将无法恢复正常网络，因此建议使用`killall ___go_build_V2R`来结束进程。如已无法正常上网，恢复网络的一种简单可行方法是重新启动程序并关闭全局透明代理。不开启全局透明代理时，GoLand调试将不受影响。

## 注意

1. 应用不会将任何用户数据保存在云端，所有用户数据存放在用户本地配置文件中。若服务端运行于 docker，则当 docker 容器被清除时配置也将随之消失，请做好备份。

2. 提供的[GUI demo](https://v2raya.mzz.pub)是由[Netlify](https://app.netlify.com/)在本 Github 项目自动部署完成的，如果担心安全性可以自行部署。

3. **不要将本项目用于不合法用途。**

# 在 docker 环境中开发

```bash
docker-compose -f docker-compose.dev.yml up --build
```

gin 会监测文件改动并热重载，见[codegangsta/gin](https://github.com/codegangsta/gin)。

# 相似项目

[v2raywebui/V2RayWebUI](https://github.com/v2raywebui/V2RayWebUI)

[jiangxufeng/v2rayL](https://github.com/jiangxufeng/v2rayL)

[NoOne-hub/v2ray_client](https://github.com/NoOne-hub/v2ray_client)

# 感谢

[hq450/fancyss](https://github.com/hq450/fancyss)

[xlzd/quickdown](https://github.com/xlzd/quickdown)

[ToutyRater/v2ray-guide](https://github.com/ToutyRater/v2ray-guide/blob/master/routing/sitedata.md)

[jiangxufeng/v2rayL](https://github.com/jiangxufeng/v2rayL)

# 协议

[![License: GPL v3](https://img.shields.io/badge/License-GPL%20v3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
