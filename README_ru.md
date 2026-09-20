# v2rayA [![Docker Cloud Build Status](https://img.shields.io/docker/cloud/build/v2rayA/v2raya)](https://hub.docker.com/r/mzz2017/v2raya) [![Travis (.org)](https://img.shields.io/travis/v2rayA/v2rayA?label=travis-ci%20build)](https://travis-ci.org/v2rayA/v2rayA)

[**English**](https://github.com/v2rayA/v2rayA/blob/main/README.md)&nbsp;&nbsp;&nbsp;[**简体中文**](https://github.com/v2rayA/v2rayA/blob/main/README_zh.md)&nbsp;&nbsp;&nbsp;[**Русский**](https://github.com/v2rayA/v2rayA/blob/main/README_ru.md)

v2rayA — веб-клиент для собственного ядра на базе Xray с глобальным прозрачным прокси в Linux, Windows и macOS. Он поддерживает ссылки на прокси VMess, VLESS, Shadowsocks, Trojan, Hysteria2, TUIC, [Juicity](https://github.com/juicity), AnyTLS, WireGuard, SOCKS5 и HTTP(S); ShadowsocksR больше не поддерживается.

Мы стремимся сделать управление как можно проще и учесть большинство сценариев использования.

Веб-интерфейс позволяет использовать v2rayA не только на локальном компьютере, но и без труда развернуть его на маршрутизаторе или NAS.

Проект: https://github.com/v2rayA/v2rayA


## Использование

Основные способы установки v2rayA:

1. Из репозитория APT или AUR
2. Docker
3. Собственный [репозиторий scoop](https://github.com/v2rayA/v2raya-scoop) (для пользователей Windows)
4. Собственный [репозиторий homebrew](https://github.com/v2rayA/homebrew-v2raya)
5. Собственный [репозиторий OpenWrt](https://github.com/v2rayA/v2raya-openwrt) или официальный репозиторий OpenWrt (начиная с OpenWrt 22.03)
6. Microsoft winget: https://winstall.app/apps/v2rayA.v2rayA
7. Ubuntu Snap: https://snapcraft.io/v2raya
8. Бинарный файл или установочный пакет из выпусков на GitHub

Подробнее: [**v2rayA - Docs**](https://v2raya.org/en/docs/prologue/introduction/)


## Прозрачный прокси

В Linux прозрачный прокси доступен в режимах `redirect`, `tproxy` и `tun`; в Windows и macOS — в режиме `tun` или как системный прокси.

`tun` встроен в ядро: ядро открывает устройство TUN, назначает ему адрес и берёт на себя маршрут по умолчанию. Собственные соединения ядра, прямой исходящий трафик и запросы DNS-модуля к вышестоящим серверам никогда не попадают в TUN: в Linux для этого используются метки сокетов, а в Windows и macOS — привязка к физическому интерфейсу. На все DNS-запросы приложений к публичным серверам отвечает DNS-модуль ядра. v2rayA и ядро исключаются всегда; другие процессы можно исключить в настройках по имени исполняемого файла. Маршруты подключённых сетей и статические маршруты всегда обходят TUN.

Известное ограничение: в Windows и macOS приложение, которое обращается напрямую к DNS-серверу в локальной сети, по-прежнему обходит TUN. Системный DNS-резолвер направлен в TUN и этим ограничением не затронут.

## Снимки экрана

Панель управления: ядро, текущий узел, трафик в реальном времени, режимы прозрачного прокси и маршрутизации, задержка узлов и подписки — каждый элемент на отдельной плитке.

<img src="docs/images/screenshot.png" alt="Панель управления v2rayA в светлой и тёмной темах" width="100%">

Редактор RoutingA: правила можно редактировать списком с помощью редактора правил или как текст с номерами строк, подсветкой и построчной проверкой; синтаксис всегда под рукой; доступны импорт и экспорт.

<img src="docs/images/routinga.png" alt="Редактор RoutingA в светлой и тёмной темах" width="100%">

## Примечания

1. Программа не хранит пользовательские данные в облаке: все они хранятся локально.
2. **Не используйте этот проект в противоправных целях.**

## Благодарности

[hq450/fancyss](https://github.com/hq450/fancyss)

[ToutyRater/v2ray-guide](https://github.com/ToutyRater/v2ray-guide/blob/master/routing/sitedata.md)

[nadoo/glider](https://github.com/nadoo/glider)

[Loyalsoldier/v2ray-rules-dat](https://github.com/Loyalsoldier/v2ray-rules-dat)

[zfl9/ss-tproxy](https://github.com/zfl9/ss-tproxy/blob/master/ss-tproxy)

## Динамика звёзд

[![Stargazers over time](https://starchart.cc/v2rayA/v2rayA.svg)](https://starchart.cc/v2rayA/v2rayA)

## Лицензия

[![License: AGPL v3-only](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
