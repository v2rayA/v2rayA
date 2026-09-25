# Быстрый старт

Эта страница в браузере позволяет настраивать v2rayA и показывает его состояние; сервис работает от имени root (администратора в Windows; без привилегий с `--lite`) и сам запускает `v2raya_core`.

## Вход

Первая зарегистрированная учётная запись становится администратором; других учётных записей нет. При регистрации нужно указать имя пользователя и пароль длиной от 6 до 32 символов; существующие учётные данные не требуются. Поэтому в общей сети ограничьте доступ к порту 2017 до регистрации или запустите сервис с `--address 127.0.0.1:2017`, чтобы разрешить только локальный доступ.

Забытый пароль сбрасывается из командной строки при остановленном сервисе:

```sh
v2raya --reset-password
```

Выполните команду от имени учётной записи сервиса и с его каталогом `--config` (в режиме lite также укажите `--lite`); в Windows сервис работает от имени SYSTEM, и его каталог отличается от каталога пользователя с повышенными правами. Снова запустите сервис и зарегистрируйте новую учётную запись.

## Импорт узлов

На странице **Прокси** кнопка **Импортировать** позволяет добавить ссылки на узлы или подписку: выберите **Ссылка на сервер** для ссылок `vmess://`, `vless://`, `ss://`, `trojan://`, `hysteria2://`, `tuic://`, `juicity://`, `anytls://`, `wireguard://`, `socks5://`, `http://` и `https://`, по одной в строке, или изображения с QR-кодом; для подписки выберите **Адрес подписки**. Ссылки ShadowsocksR не принимаются, как и ссылки Shadowsocks с потоковым шифром (`rc4-md5`, `aes-*-cfb`, `chacha20-ietf`) или `none`: ядро принимает только шифры AEAD и методы 2022-blake3. `allow_insecure` в ссылке игнорируется: v2rayA не пропускает проверку сертификата; для самоподписанного сервера укажите SHA-256 сертификата в форме узла.

У подписок своя страница — **Подписки** (на телефоне документация переезжает в меню верхней панели, чтобы освободить место). Узлы каждой подписки хранятся вместе; подписки можно обновлять вручную или по расписанию (**Настройки → Автоматически обновлять подписки**). Соседний параметр режима определяет, будет ли обновление выполняться через прокси.

## Группы

Узлы используются через группы прокси. Группа `proxy` существует всегда; остальные создаются, настраиваются и удаляются через кнопку групп в верхней панели — это единственное место для управления группами. Узел добавляется в группу из своего меню, либо выделите несколько на странице «Прокси» и выберите группу в **Добавить в группу прокси**. Группа с несколькими подключёнными участниками использует участника с наименьшей измеренной задержкой (**Авто (минимальная задержка)**); меню узла на панели позволяет закрепить одного участника, чтобы группа использовала только его, **Добавить или удалить узлы** меняет состав участников, а в настройках группы задаются адрес и интервал проверки.

В правилах маршрутизации группы обозначают исходящие подключения: по умолчанию `proxy`, а любая другая группа — под своим именем, как только у неё появится подключённый участник.

## Запуск ядра

Кнопка **Запустить** на панели или кнопка состояния вверху страницы запускает `v2raya_core` с текущей конфигурацией. Изменения на странице «Настройки» вступают в силу по нажатию **Сохранить и применить**: работающее ядро перезапускается с ними, остановленное применяет их при следующем запуске.

После запуска ядра приложения подключаются к нему через локальные входящие порты:

| Входящий порт    | Адрес             | Маршрутизация                                          |
| ---------------- | ----------------- | ------------------------------------------------------ |
| SOCKS5           | `127.0.0.1:20170` | всё через `proxy`                                      |
| HTTP             | `127.0.0.1:20171` | всё через `proxy`                                      |
| HTTP с правилами | `127.0.0.1:20172` | параметр **Режим разделения трафика для порта правил** |

Порты меняются в разделе **Настройки → Адреса и порты**; 0 закрывает входящий порт. Чтобы направлять трафик приложений через прокси без отдельной настройки каждого приложения, включите прозрачный прокси.

## Маршрутизация

В разделе **Настройки → Разделение трафика** выбирается способ разделения трафика для порта правил и прозрачного прокси: всё через прокси, кроме китайских сайтов, только GFWList через прокси или собственные правила RoutingA. RoutingA описан в отдельном разделе.

## Клавиатура

- Страница «Прокси», вид списком: `Ctrl`+`A` выделяет все показанные узлы, `Esc` снимает выделение; `/` переводит в поиск на страницах «Прокси» и «Журнал».
- Настройки: `Ctrl`+`S` сохраняет. Редактор RoutingA: `Ctrl`+`S` сохраняет, `Tab` делает отступ.
- Диалоги с многострочным полем (импорт, подписка, скрипты маршрутов, списки): `Ctrl`+`Enter` отправляет; `Esc` закрывает любой диалог.
- Журнал: при фокусе `Home`, `End`, `PgUp`, `PgDn` листают журнал.

В macOS вместо `Ctrl` — `Cmd`.

## Automatic groups and subscription updates

Enable **Automatically add available servers** in a proxy group's settings to
check every server in Proxies, including standalone nodes and all subscriptions.
The group keeps only servers that complete an HTTP(S) request through their VPN
connection. A listening TCP port or an ICMP response is not enough.

The first scan runs when enabled, then after catalog changes and at the group's
probe interval. The default for new groups is **300s**; saved intervals are
preserved. Turning the switch off keeps the current members and returns control
to manual editing. The group's existing Auto/least-ping mode chooses the server.
Subscription updates do not select a server or populate manual groups.

Enable **Automatically update subscription** for each subscription that should
refresh itself. Both minute fields are required:

- **Regular update interval:** 0 disables regular downloads; a positive whole
  number schedules unconditional updates.
- **Retry interval when all servers are unavailable:** at least 1 minute. If no
  saved candidate is reachable, refresh at this interval until one works.
  Failed or empty downloads retain the saved list. A failed subscription retries
  independently of healthy subscriptions in the same group.

Candidate health is checked after subscription/catalog changes and every 300
seconds for enabled subscription updates. Each automatic group uses its own
probe URL and interval. Timers start when enabled or the service starts.
Retries do not postpone unconditional updates. No background action starts a
manually stopped main core; temporary candidate probes can still run.

An empty automatic group has a blocking outbound. Traffic assigned to that group
is rejected instead of falling back to another group or direct access. Explicit
direct routes and other groups retain their own policy. On OpenWrt TPROXY and
Redirect, membership reloads retain existing interception when network settings
are unchanged and no custom hooks are installed. This is not a system-wide kill
switch for power loss, service termination, custom hooks, or TUN teardown.

Upgrading from the earlier Resilient recovery policy preserves accounts,
subscriptions, nodes and saved group members. Old first-server, auto-connect,
recovery and global subscription schedule flags no longer run. The new switches
default off: enable the desired group and subscription policies explicitly.
