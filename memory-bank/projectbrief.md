# v2rayA 项目概述

## 项目简介
v2rayA 是一个 V2Ray 客户端，支持在 Linux 上的全局透明代理和在 Windows/macOS 上的系统代理。

## 核心功能
- 支持多种协议：SS、SSR、Trojan(trojan-go)、Tuic、Juicity、V2Ray
- Web GUI 界面管理
- 透明代理（Linux TProxy/Redirect）
- 系统代理设置
- 订阅管理和自动更新
- 规则路由（RoutingA）
- GFWList 自动更新

## 项目目标
提供最简单的操作方式来满足大多数代理需求，可部署在本地计算机、路由器或 NAS 上。

## 技术栈概览
- **后端**: Go 1.21+, Gin Web 框架
- **前端 (旧版)**: Vue 2.7 + Buefy
- **前端 (新版)**: Nuxt 3 + Element Plus + UnoCSS
- **数据库**: BoltDB (嵌入式键值数据库)
- **代理核心**: v2ray-core / xray-core

## 部署方式
1. APT/AUR 包管理器安装
2. Docker 容器
3. Windows Scoop/Winget
4. macOS Homebrew
5. OpenWrt 软件源
6. Ubuntu Snap
7. GitHub Releases 二进制文件

## 许可证
AGPL-3.0
