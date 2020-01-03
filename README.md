# V2RayA

V2RayA 是一个支持全局透明代理的 V2Ray Linux 客户端，同时兼容SS、SSR协议。

V2RayA 致力于提供最简单的操作，满足绝大部分需求。

同时兼容V2Ray、SS、SSR意味着用户不再需要在不同工具之间切换，你甚至可以使用一个混合协议的订阅。

虽然 V2RayA 是一个 Web 客户端，它也支持以 PWA(Progressive Web App)的方式享受桌面端应用的体验。[食用方法](https://www.ithome.com/0/414/429.htm)

得益于Web客户端的优势，你不仅可以将其用于本地计算机，还可以轻松地将它部署在路由器或NAS上。

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

待开发：

- [ ] 自定义 PAC 路由规则
- [ ] QUIC、auth_chain\*支持
- [ ] 透明代理重定向备选方案
- [ ] 日志

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
<p align="center">使用自定义PAC时可以配置路由规则（待实现）</p>


</details>

## 使用

### 从软件源安装

如下使用方法：

1. **使用apt-get安装**（debian、ubuntu）
   
   请确保已正确安装 v2ray-core

   我们提供了 Linux 下的一键安装脚本（在官方脚本基础上增加了ustc镜像源）：

   > 运行下面的指令下载并安装 V2Ray。当 yum 或 apt-get 可用的情况下，此脚本会自动安装 unzip 和 daemon。这两个组件是安装 V2Ray 的必要组件。如果你使用的系统不支持 yum 或 apt-get，请自行安装 unzip 和 daemon

   ```bash
   curl -L -s https://github.com/mzz2017/V2RayA/raw/master/install/go.sh | sudo -E bash -s - --source ustc
   ```
   
   准备完毕后：

   ```bash
   # add public key
   wget -qO - https://apt.v2raya.mzz.pub/key/public-key.asc | sudo apt-key add -

   # add V2RayA's repository
   sudo add-apt-repository 'deb https://apt.v2raya.mzz.pub/ v2raya main'
   sudo apt-get update

   # install V2RayA
   sudo apt-get install v2raya
   ```
   
   V2RayA服务端正常运行后，就可在[GUI demo](https://v2raya.mzz.pub)使用了（或[部署GUI](https://github.com/mzz2017/V2RayA#%E5%A6%82%E4%BD%95%E9%83%A8%E7%BD%B2GUI)）。
   
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

   V2RayA服务端正常运行后，就可在[GUI demo](https://v2raya.mzz.pub)使用了（或[部署GUI](https://github.com/mzz2017/V2RayA#%E5%A6%82%E4%BD%95%E9%83%A8%E7%BD%B2GUI)）。

### Docker方式

1. 拉取源码，**使用 docker-compose 部署**。

   ```bash
   git clone --depth=1 https://github.com/mzz2017/V2RayA.git
   cd V2RayA
   docker-compose up -d --build
   ```
   如果出现`ERROR: ...Connot start service...container...is not running`，尝试添加参数`-V`

2. **使用 docker 命令部署**。

   ```bash
   # pull stable version of v2raya
   docker pull mzz2017/v2raya:stable
   # pull latest version of v2ray
   docker pull v2ray/official
   
   # create volume to share data
   docker volume create v2raya_shared-data
   
   # run v2raya
   docker run -d \
   	--restart=always \
   	--privileged \
   	--network=host \
   	-v v2raya_shared-data:/etc/v2ray \
   	-v /etc/localtime:/etc/localtime:ro \
   	-v /etc/timezone:/etc/timezone:ro \
   	--name v2raya_backend \
   	mzz2017/v2raya:stable

   # run v2ray
   docker run -d \
   	--restart=always \
   	--privileged \
   	--network=host \
   	--pid=container:v2raya_backend \
   	-v v2raya_shared-data:/etc/v2ray \
   	--env V2RAY_LOCATION_ASSET=/etc/v2ray \
   	--name v2raya_v2ray \
   	v2ray/official \
   	sh -c "cp -rfu /usr/bin/v2ray/* /etc/v2ray/ && v2ray -config=/etc/v2ray/config.json"
   ```
   
   如果你使用MacOSX或其他不支持host模式的环境，在该情况下无法使用全局透明代理，docker命令会略有不同：
   
   ```bash
   # pull stable version of v2raya
   docker pull mzz2017/v2raya:stable
   # pull latest version of v2ray
   docker pull v2ray/official
   
   # create volume to share data
   docker volume create v2raya_shared-data
   
   # run v2raya
   docker run -d \
       -p 2017:2017 \
       -p 20170-20172:20170-20172 \
       -p 12345:12345 \
       --restart=always \
       --privileged \
       -v v2raya_shared-data:/etc/v2ray \
       --name v2raya_backend \
       mzz2017/v2raya:stable
       
   # run v2ray
   docker run -d \
       --restart=always \
       --privileged \
       --pid=container:v2raya_backend \
       --network=container:v2raya_backend \
       -v v2raya_shared-data:/etc/v2ray \
       --env V2RAY_LOCATION_ASSET=/etc/v2ray \
       --name v2raya_v2ray \
       v2ray/official \
       /bin/sh -c "cp -rfu /usr/bin/v2ray/* /etc/v2ray/ && v2ray -config=/etc/v2ray/config.json"
   ```
   
   

### 二进制文件、安装包

请确保已正确安装 v2ray-core

我们提供了 Linux 下的一键安装脚本（在官方脚本基础上增加了ustc镜像源）：

> 运行下面的指令下载并安装 V2Ray。当 yum 或 apt-get 可用的情况下，此脚本会自动安装 unzip 和 daemon。这两个组件是安装 V2Ray 的必要组件。如果你使用的系统不支持 yum 或 apt-get，请自行安装 unzip 和 daemon

```bash
curl -L -s https://github.com/mzz2017/V2RayA/raw/master/install/go.sh | sudo -E bash -s - --source ustc
```

准备完毕后，可下载[Releases](https://github.com/mzz2017/V2RayA/releases)中的二进制文件启动V2RayA服务端，或下载安装包进行安装。

V2RayA服务端正常运行后，就可在[GUI demo](https://v2raya.mzz.pub)使用了（或[部署GUI](https://github.com/mzz2017/V2RayA#%E5%A6%82%E4%BD%95%E9%83%A8%E7%BD%B2GUI)）。

### 自行编译运行

当然，你也可以选择拉取源码，**通过 golang 启动**：

该方法同样需要正确安装v2ray-core，详情见上

```bash
git clone https://github.com/mzz2017/V2RayA.git
cd V2RayA/service
export GOPROXY=https://goproxy.io # set goproxy.io as the proxy of go modules
sudo go run main.go
```

注意，尽管 golang 具有交叉编译的特性，但由于项目使用了大量 linux commands，导致该方法仍然不支持 windows。若想在 windows 体验，可尝试借助 Docker 或 WSL。

### 在路由器使用

分为以下几种情况：

#### 若v2ray能够以daemon存在

能够以daemon存在即在正确安装v2ray后，使用下述命令之一能够得到正确的反馈：

```bash
# if systemctl is available
systemctl status v2ray
# else if service is available
service v2ray status
```

那么可从软件源安装，或下载[releases](https://github.com/mzz2017/V2RayA/releases)中的对应安装包进行安装。

#### 若v2ray能够运行于docker

可参照Docker方式使用

#### 通用方法

1. 请自行安装v2ray，并确保v2ray、v2ctl均被包含在PATH中，否则请将上述文件放于`echo $PATH`中的任一目录下。

2. 下载[releases](https://github.com/mzz2017/V2RayA/releases)中最新版本的对应CPU架构的二进制文件，或自行使用golang交叉编译。

3. 使用参数`--config=V2RAYA_CONFIG_PATH --mode=common`启动V2RayA服务端，参数含义可执行`--help`查看。

   请将上述V2RAYA_CONFIG_PATH替换为一个可读写的，并且你喜欢的路径。

## 开放端口

默认使用的四个端口分别为：

2017: V2RayA 后端端口

20170: SOCKS 协议

20171: HTTP 协议

20172: 带 PAC 的 HTTP 协议

其他端口：

12345: tproxy（全局透明代理所需）

12346: ssr server（SS、SSR所需）

## 在不同运行环境下程序表现将不同

由于 docker 容器对 systemd 的限制性，在 docker 中将采用 pid 共享进程命名空间，volumes 共享存储空间，更新配置后通过结束进程触发 v2ray 容器的重启来更新配置，以无 inbounds 的配置代替断开连接，这是一种折中方案，会有如下影响：

1. 在更换配置时略有卡顿

## 如何部署GUI

一般情况下可使用[demo](https://v2raya.mzz.pub)即可满足需求，如有部署GUI的必要，可参考下述文档：

**使用docker一键部署**

```bash
docker pull mzz2017/v2raya-gui
docker run --name v2raya-gui -d -p <port>:80 mzz2017/v2raya-gui
```
将上述`<port>`替换为任一本地端口即可。

**手动部署**

见 [README](https://github.com/mzz2017/V2RayA/tree/master/gui#v2raya-gui)

## 开发相关

### 在 docker 环境中开发

```bash
docker-compose -f docker-compose.dev.yml up --build
```

gin 会监测文件改动并热重载，见[codegangsta/gin](https://github.com/codegangsta/gin)。

如果出现`ERROR: ...Connot start service...container...is not running`，尝试添加参数`-V`

### 已知问题

在使用 GoLand 进行开发调试时，**如果开启了全局透明代理**，由于进程捕获不了 GoLand 的结束 signal，在进程退出后将无法恢复正常网络，因此建议使用`killall ___go_build_V2R`来结束进程。如已无法正常上网，恢复网络的一种简单可行方法是重新启动程序并关闭全局透明代理。不开启全局透明代理时，GoLand调试将不受影响。



# 注意

1. 程序不会将任何用户数据保存在云端，所有用户数据存放在用户本地配置文件中。若服务端运行于 docker，则当 docker 容器被清除时配置也将随之消失，请做好备份。

2. 提供的[GUI demo](https://v2raya.mzz.pub)是由[Netlify](https://app.netlify.com/)在本 Github 项目自动部署完成的，如果担心安全性可以自行部署。

3. **不要将本项目用于不合法用途。**



# 感谢

[hq450/fancyss](https://github.com/hq450/fancyss)

[ToutyRater/v2ray-guide](https://github.com/ToutyRater/v2ray-guide/blob/master/routing/sitedata.md)

[nadoo/glider](https://github.com/nadoo/glider)

# 协议

[![License: GPL v3](https://img.shields.io/badge/License-GPL%20v3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)