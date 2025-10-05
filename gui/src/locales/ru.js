export default {
  common: {
    outboundSetting: "Настройка исходящего узла",
    setting: "Настройки",
    about: "О программе",
    loggedAs: "Текущий логин: <b>{username}</b>",
    v2rayCoreStatus: "Статус v2ray-core",
    checkRunning: "Проверка",
    isRunning: "Работает",
    notRunning: "Готово",
    notLogin: "Пожалуйста, войдите",
    latest: "Последняя",
    local: "Текущая",
    success: "УСПЕХ",
    fail: "ПРОВАЛ",
    message: "Сообщение",
    none: "нет",
    optional: "необязательно",
    loadBalance: "Load Balance",
    log: "Журнал",
  },
  welcome: {
    title: "Добро пожаловать",
    docker: "v2rayA сервис запускается в Docker. Версия: {version}",
    default: "v2rayA сервис запущен. Версия: {version}",
    newVersion: "Обнаружена новая версия: {version}",
    messages: [
      "Здесь нет ни одного сервера.",
      "Вы можете создать или добавить сервер, а также подписку.",
    ],
  },
  v2ray: {
    start: "Пуск",
    stop: "Стоп",
  },
  server: {
    name: "Название сервера",
    address: "Адрес сервера",
    protocol: "Протокол",
    latency: "Задержка",
    lastSeenTime: "Last seen time",
    lastTryTime: "Last try time",
    messages: {
      notAllowInsecure:
        "Согласно документации {name}, если вы используйте {name}, опция AllowInsecure запрещена.",
      notRecommend:
        "Согласно документации {name}, если вы используете {name}, опция AllowInsecure не рекомендуется.",
    },
  },
  InSecureConfirm: {
    title: "Найдена небезопасная настройка",
    message:
      "В конфигурации установлена опция <b>AllowInsecure</b> в значении true. Это может вызвать риски безопасности. Вы уверены, что хотите продолжить?",
    confirm: "Понимаю, что делаю",
    cancel: "отмена",
  },
  subscription: {
    host: "Хост",
    remarks: "Примечания",
    timeLastUpdate: "Время последнего обновления",
    numberServers: "Число серверов",
    subscription: "Подписка",
    autoSelect: "Автоматически подключаться к новым серверам после автоматичкского обновления подписки",
  },
  operations: {
    name: "Операции",
    update: "Обновить",
    modify: "Изменить",
    share: "Поделиться",
    view: "Посмотреть",
    delete: "Удалить",
    create: "Создать",
    import: "Импортировать",
    inBatch: "Группа",
    connect: "Подключиться",
    disconnect: "Отключиться",
    select: "Выбрать",
    login: "Войти",
    logout: "Выйти",
    configure: "Настройка",
    cancel: "Отмена",
    saveApply: "Сохранить и применить",
    confirm: "Подтвердить",
    confirm2: "Точно подтвердить",
    save: "Сохранить",
    copyLink: "Скопировать ссылку",
    helpManual: "Помощь и руководство",
    yes: "Да",
    no: "Нет",
    switchSite: "Переключится в альтернативный сайт",
    addOutbound: "Добавить исходящий узел",
    domainsExcluded: "Исключённые домены"
  },
  register: {
    title: "Создайте учётную запись администратора",
    messages: [
      "Запомните данные учётной записи администратора. Это важно для входа в систему",
      "Информация о учётной записи хранится локально. Мы никогда не отправляем данные на любой сервер.",
      "Если забыли пароль, используйте v2raya --reset-password для сброса.",
    ],
  },
  login: {
    title: "Вход в систему",
    username: "Имя пользователя",
    password: "Пароль",
  },
  setting: {
    transparentProxy: "Прозрачный прокси/Системный прокси",
    transparentType: "Реализация прозрачного прокси/Системного прокси",
    pacMode: "Режим разделения трафика на порте с правилами",
    preventDnsSpoofing: "Предотвратить DNS-спуфинг",
    specialMode: "Специальный режим",
    mux: "Мультиплекс",
    autoUpdateSub: "Автоматически обновлять подписки",
    autoUpdateGfwlist: "Автоматически обновлять GFWList",
    preferModeWhenUpdate: "Режим обновления подписок и GFWList",
    ipForwardOn: "IP форвардинг",
    portSharingOn: "Port Sharing",
    concurrency: "Параллелизм",
    inboundSniffing: "Сниффер",
    options: {
      global: "Не разделять трафик",
      direct: "Напрямую",
      pac: "Зависит от порта правил",
      whitelistCn: "Использовать прокси кроме CN сайтов",
      gfwlist: "Использовать прокси только для сайтов GFWList",
      sameAsPacMode:
        "Режим разделения трафика такой же, как у порта с правилами",
      customRouting: "Настраиваемая адресация",
      antiDnsHijack: "Только защита от перехвата DNS (быстро)",
      forwardDnsRequest: "Перенаправлять DNS запросы",
      doh: "DoH (dns-over-https)",
      default: "По-умолчанию",
      on: "Включено",
      off: "Выключено",
      updateSubWhenStart: "Обновлять подписки при запуске сервиса",
      updateSubAtIntervals: "Обновлять подписки регулярно (в часах)",
      updateGfwlistWhenStart: "Обновлять GFWList при запуске сервиса",
      updateGfwlistAtIntervals: "Обновлять GFWList регулярно (в часах)",
      dependTransparentMode: "Следовать за Прозрачным прокси/Системным Прокси",
      closed: "Выключено",
      advanced: "Расширенная настройка",
      leastPing: "С наименьшей задержкой",
    },
    messages: {
      inboundSniffing: "Анализировать входящий трафик. Если эта опция выключена, часть трафик может быть не перенаправлена корректно.",
      gfwlist:
        "Основывается на времени изменения файла, которое иногда бывает после последней версии в Интернете.",
      transparentProxy:
        "Если прозрачный прокси включен, это не нужно дополнительно настраивать, и весь TCP-трафик будет проходить через v2rayA. При предоставлении прокси-сервиса другим компьютерам и при использовании docker в качестве шлюза необходимо включить опцию 'Share in LAN'.",
      transparentType:
        "★tproxy: поддерживает UDP, но не поддерживает docker. ★redirect: подходит для docker, но не поддерживает UDP и требуется занять локальный порт 53 для защиты dns от загрязнения.",
      pacMode: `Здесь вы можете выбрать правила разделения трафика для порта с правилами. По-умолчанию, порт для разделения трафика это 20172 с протоколом HTTP.`,
      preventDnsSpoofing:
        "★Перенаправлять DNS запросы: DNS запросы будут отправляться через прокси-сервер." +
        "★DoH(dns-over-https, v2ray-core: 4.22.0+): DNS over HTTPS.",
      specialMode:
        "★supervisor：Мониторинг загрязнения dns, перехват заранее, используется механизм сниффинга v2ray-core для предотвращения загрязнения. ★fakedns：Использование стратегии fakedns для ускорения резолвинга.",
      tcpFastOpen:
        "Упростить процесс TCP рукопожатия для ускорения установки соединения. Есть риск распознавания характеристик пакетов. Это может помешать установить соединение если ваша система не подергивает это",
      mux:
        "Мультиплексировать TCP соединения чтобы снизить количество рукопожатий, но это влияет на варианты использования с высоким трафиком, такие как просмотр видео, скачивание, и проверка скорости. " +
        "Есть риск распознавания характеристик пакетов. Поддерживается только в vmess.",
      confirmEgressPorts: `<p>Вы настраиваете прозрачный прокси-сервер между локальными сетями, подтвердите белый список портов на выходе.</p>
                          <p>Белый список:</p>
                          <p>TCP: {tcpPorts}</p>
                          <p>UDP: {udpPorts}</p>`,
      grpcShouldWithTls: `gRPC должен быть с TLS`,
      ssPluginImpl:
        "★по-умолчанию: 'transport' для simple-obfs, 'chained' для v2ray-plugin." +
        "★chained: shadowsocks трафик будет перенаправлен в отдельный плагин." +
        "★transport: обработка транспортным слоем ядра v2ray/xray напрямую.",
    },
  },
  customAddressPort: {
    title: "Адреса и порты",
    serviceAddress: "Адрес сервиса",
    portSocks5: "Порт SOCKS5",
    portHttp: "Порт HTTP",
    portSocks5WithPac: "Порт SOCKS5 (с правилами)",
    portHttpWithPac: "Порт HTTP (с правилами)",
    portVmess: "Порт VMess (с правилами)",
    portVmessLink: "Ссылка порта VMess",
    messages: [
      "Адрес сервера по-умолчанию 0.0.0.0:2017 может быть изменен через переменную среды <code>V2RAYA_ADDRESS</code> и через аргумент<code>--address</code> команды v2raya.",
      "Если вы запускайте v2raya docker контейнер с назначением портов вместо <code>--network host</code>, вы можете переназначить порты здесь.",
      "Мы не можем судить о захвате портов в docker режиме. Подтвердите это сами.",
      "Ноль означает закрытие этого порта.",
    ],
  },
  customRouting: {
    title: "Настройка порта с правилами",
    defaultRoutingRule: "Правило адресации по-умолчанию",
    sameAsDefaultRule: "такой-же как порт по-умолчанию",
    appendRule: "Вставить правило",
    direct: "Напрямую",
    proxy: "Прокси",
    block: "Блокировать",
    rule: "Правило",
    domainFile: "Файл с доменами",
    typeRule: "Тип правила",
    messages: {
      0: "v2rayA будет распознавать все файлы SiteDat в <b>{V2RayLocationAsset}</b>",
      1: 'Чтобы создать SiteDat файл самому: <a href="https://github.com/ToutyRater/V2Ray-SiteDAT">ToutyRater/V2Ray-SiteDAT</a>',
      2: "Поддерживается Multi-select",
      noSiteDatFileFound: "SiteDat файл не найден в {V2RayLocationAsset}",
      emptyRuleNotPermitted: "Пустые правила не разрешаются",
    },
  },
  doh: {
    title: "Настройка DoH сервера",
    dohPriorityList: "Список приоритетных DoH серверов",
    messages: [
      "DoH (DNS over HTTPS) может эффективно избегать DNS загрязнения. Но некоторые DoH провайдеры сами могут быть загрязнены. В дополнение к этому, некоторые DoH сервисы могут быть заблокированы местными сетевыми провайдерами. Пожалуйста выберите самый быстрый и стабильный DoH сервер.",
      "Отличные публичные DoH сервера в Материковом Китае: alidns, geekdns, rubyfish, и т.д",
      "На территории Тайваня: quad101, и т.д",
      "США: cloudflare, dns.google, etc",
      'Список вариантов:<a href="https://dnscrypt.info/public-servers" target="_blank">публичных серверов</a>',
      'Кроме того, настройка службы DoH на своем сервере хорошо работает во многих случаях <a href="https://github.com/facebookexperimental/doh-proxy" target="_blank">doh-proxy</a>. В этом случае, рекомендуется запустить сервер (doh-proxy/doh-httpproxy), предоставлявший сервис и клиент (doh-stub) подключающийся к doh.opendns.com и соединить их в связке, потому что вам сложно найти сервер который не загрязнем в целом загрязненном регионе.',
      "Оптимально, поместите одну или две строчки выше. Список будет восстановлен по-умолчанию, после сохранения с пустым содержимым",
    ],
  },
  dns: {
    title: "Настройка DNS сервера",
    internalQueryServers: "Серверы запросов домена",
    externalQueryServers: "Внешние серверы запросов домена",
    messages: [
      '"@:(dns.internalQueryServers)" применяются для просмотра доменных имен в Китае, в то время как "@:(dns.externalQueryServers)" используется для других доменов.',
      '"@:(dns.internalQueryServers)" будет использован для всех доменных имен если "@:(dns.externalQueryServers)" пуст.',
    ],
  },
  egressPortWhitelist: {
    title: "Настройка белового списка портов",
    tcpPortWhitelist: "Белый список портов TCP",
    udpPortWhitelist: "Белый список портов UDP",
    messages: [
      "Если v2rayA настроен на сервере A, который подключается к прокси-серверу B, обратите внимание:",
      "Прозрачный прокси будет перенаправлять весь TCP и UDP трафик через прокси-сервер B, где IP адрес источника будет заменен на адрес прокси B. Более того, если некоторые клиенты отправят запросы на сервер A который предоставляет сервис, они получат странные ответы от сервера прокси B, что является неверным.",
      "Чтобы решить это, нужно добавить эти порты сервиса в белый список, чтобы не пропускать через прокси. Например: ssh(22)、v2raya({v2rayaPort}).",
      "Очевидно, что если сервер не предоставляет никаких сервисов, вы можете пропустить эту с настройку.",
      "Форматирование：22 означает порт 22，20170:20172 означает три порта с 20170 по 20172.",
    ],
  },
  configureServer: {
    title: "Настройка сервера | Сервер",
    servername: "Название сервера",
    port: "Порт",
    forceTLS: "Принудительно включить TLS",
    noObfuscation: "Без обфускации",
    httpObfuscation: "Обфускация под HTTP",
    srtpObfuscation: "Обфускация под видео-звонки (SRTP)",
    utpObfuscation: "Обфускация под Bittorrent (uTP)",
    wechatVideoObfuscation: "Обфускация под видео-звонки Wechat",
    dtlsObfuscation: "Обфускация под пакеты DTLS1.2",
    wireguardObfuscation: "Обфускация под пакеты WireGuard",
    hostObfuscation: "Хост",
    pathObfuscation: "Путь",
    seedObfuscation: "Сид",
    username: "Имя пользователя",
    password: "Пароль",
    origin: "origin",
    pinnedCertchainSha256: "pinned certificate chain sha256",
  },
  configureSubscription: {
    title: "Настройка подписки",
  },
  import: {
    message: "Введите ссылку на сервер или адрес подписки:",
    batchMessage: "Одна ссылка сервера на строку:",
    qrcodeError: "Не удалось найти верный QRCode, пожалуйста попробуйте снова",
  },
  delete: {
    title: "Подтвердить УДАЛЕНИЕ",
    message:
      "Вы уверены что хотите <b>УДАЛИТЬ</b> эти сервера/подписки? Это нельзя отменить.",
  },
  latency: {
    message:
      "Проверка задержки занимает как правило одну или две минуты. Пожалуйста подождите.",
  },
  version: {
    higherVersionNeeded:
      "Эта операция требует версию v2rayA выше чем {version}",
    v2rayInvalid:
      "geosite.dat, geoip.dat или v2ray-core могут быть установлены некорректно",
    v2rayNotV5:
      "Версия v2ray-core не соответствует v5. Используйте v5 или понизьте версию v2rayA до v1.5",
  },
  about: `<p>v2rayA это графический веб-клиент V2Ray.</p>
          <p class="about-small">Порты по-умолчанию:</p>
          <p class="about-small">2017: порт сервиса v2rayA</p>
          <p class="about-small">20170: протокол SOCKS</p>
          <p class="about-small">20171: протокол HTTP</p>
          <p class="about-small">20172: протокол HTTP с правилами разделения трафика</p>
          <p class="about-small">Другие порты:</p>
          <p class="about-small">32345: tproxy, нужно для прозрачного прокси </p>
          <p>Все данные хранятся локально, вместо хранения на облаке.</p>
          <p>В случае найденных проблем, о них можно сообщить в разделе <a href="https://github.com/v2rayA/v2rayA/issues">ошибок</a>.</p>
          <p>Документация: <a href="https://v2raya.org">https://v2raya.org</a></p>`,
  axios: {
    messages: {
      optimizeBackend: "Изменить адрес v2rayA сервиса?",
      noBackendFound:
        "Не удалось найти v2rayA по адресу {url}. Убедитесь, что v2rayA запущен по этому адресу.",
      cannotCommunicate: [
        "Не удалось подключиться. Если сервис запущен и порты открыты верно, причина в том, что текущий браузер не разрешает https сайтам подключаться к http ресурсам, вы можете попробовать Chrome или подключиться к альтернативному http сайту.",
        "Не удалось подключиться. Firefox не разрешает https сайтам подключаться к http ресурсам, вы можете попробовать подключиться к альтернативному http сайту.",
      ],
    },
    urls: {
      usage: "https://github.com/v2rayA/v2rayA/wiki/Usage",
    },
  },
  routingA: {
    messages: ["нажмите кнопку 'Помощь и руководство' для помощи"],
  },
  outbound: {
    addMessage: "Введите имя выходного узла, который вы хотите добавить",
    deleteMessage:
      'Вы уверены, что хотите <b>УДАЛИТЬ</b> выходной узел "{outboundName}"? Это нельзя отменить.',
  },
  log: {
    logModalTitle: "Просмотр журнала",
    refreshInterval: "Интервал обновления",
    seconds: "секунд",
    autoScoll: "Авто-пролистывание",
  },
  domainsExcluded: {
    title: "Исключённые домены",
    messages: [
      "Список доменных имён. Если анализатор трафика обнаруживает доменное имя из списка, адрес назначения не будет сбрасываться."
    ],
    formName: "Список исключённых доменов",
    formPlaceholder: "courier.push.apple.com\nMijia Cloud\ndlg.io.mi.com"
  },
};
