# V2Ray

V2Ray 是 Project V 下的一个工具。Project V 包含一系列工具，帮助你打造专属的定制网络体系。而 V2Ray 属于最核心的一个。 简单地说，V2Ray 是一个与 Shadowsocks 类似的代理软件，但比Shadowsocks更具优势

V2Ray 用户手册：https://www.v2ray.com

V2Ray 项目地址：https://github.com/v2ray/v2ray-core

# V2RayA

V2RayA是一个V2Ray的Web GUI。

尽管v2ray的GUI很多，但在Linux上好用的却寥寥无几。[jiangxufeng/v2rayL](https://github.com/jiangxufeng/v2rayL)是目前Linux上较好的一个Linux GUI，但暂时无法满足我对用户体验的较高要求，因此开此项目。

## 功能清单

- [x] 检查/启动/关闭V2Ray服务
- [x] 导入vmess地址
- [x] 导入ss地址
- [x] 导入订阅地址
- [x] 删除节点/删除订阅
- [x] 连接节点
- [ ] 支持PAC模式（GFWList、大陆白名单、自定义规则）
- [ ] 登陆
- [ ] 手动添加节点
- [ ] 二维码、地址分享
- [ ] 测试节点Ping时延
- [ ] 测试节点HTTP时延

## 使用(under development)

有如下使用方法：

1. 在本地用docker-compose部署service，在[GUI](https://v2ray.mzz.pub)使用（或手动部署前端GUI）。
   
   ```bash
   docker-compose up -d
   ```
   
1. 用docker拉取镜像部署service，在[GUI](https://v2ray.mzz.pub)使用（或手动部署前端GUI）。

   ```bash
   docker pull mzz2017/V2RayA
   docker run -d -p 1080-1082:1080-1082 --restart=always mzz2017/V2Ray
   ```

其中三个端口分别为：

1080: SOCKS协议

1081: HTTP协议

1082: 带PAC的HTTP协议

用户可通过docker将上述端口映射到本地的任意端口。

## 注意

应用不会将任何用户数据保存在云端，所有用户数据存放在docker容器中，当docker容器被清除时配置也将随之消失。

提供的[GUI demo](https://v2raya.mzz.pub)是由[Render](https://render.com/)在本Github项目自动部署完成的，如果担心安全性可以自行部署。

不要将本项目用于不合法用途，作者仅将该项目用于学习研究和内网穿透的用途。

# 感谢

[jiangxufeng/v2rayL](https://github.com/jiangxufeng/v2rayL)

[2dust/v2rayN](https://github.com/2dust/v2rayN)

[hq450/fancyss](https://github.com/hq450/fancyss)

# Similar projects

[v2raywebui/V2RayWebUI](https://github.com/v2raywebui/V2RayWebUI)

# 协议

[![License: GPL v3](https://img.shields.io/badge/License-GPL%20v3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)