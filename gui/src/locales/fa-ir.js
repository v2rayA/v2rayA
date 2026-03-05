export default {
  common: {
    outboundSetting: "تنظیمات خروجی",
    setting: "تنظیمات",
    about: "درباره",
    loggedAs: "خوش آمدید <b>{username}</b>",
    v2rayCoreStatus: "وضعیت v2ray-core",
    checkRunning: "بررسی",
    isRunning: "درحال اجرا",
    notRunning: "آماده",
    notLogin: "لطفا وارد شوید",
    latest: "آخرین",
    local: "محلی",
    success: "موفق",
    fail: "عدم موفقیت",
    message: "پیام",
    none: "هیچ یک",
    optional: "اختیاری",
    loadBalance: "متعادل کردن",
    log: "گزارش ها",
  },
  welcome: {
    title: "خوش آمدی",
    docker: "v2rayA بر روی Docker درحال اجرا است. نسخه: {version}",
    default: "v2rayA درحال اجرا است. نسخه: {version}",
    newVersion: "ورژن جدید موجود است: {version}",
    messages: [
      "سروری وجود ندارد.",
      "با گزینه وارد کردن (import) یا subscription یک سرور اضافه کنید.",
    ],
  },
  v2ray: {
    start: "شروع",
    stop: "خاتمه",
  },
  server: {
    name: "نام سرور",
    address: "آدرس سرور",
    protocol: "پروتکل",
    latency: "تاخیر",
    lastSeenTime: "آخرین بار مراجعه",
    lastTryTime: "آخرین زمان تست کردن",
    messages: {
      notAllowInsecure:
        "According to the docs of {name}, if you use {name}, AllowInsecure will be forbidden.",
      notRecommend:
        "According to the docs of {name}, if you use {name}, AllowInsecure is not recommend.",
    },
  },
  InSecureConfirm: {
    title: "پیکربندی خطرناک شناسایی شد",
    message:
      "تنظیمات به <b>AllowInsecure</b> تغییر کرد. ممکن از خطراتی داشته باشد. آیا از ادامه مطمعن هستید?",
    confirm: "من میدانم چه کار دارم میکنم",
    cancel: "لغو",
  },
  subscription: {
    host: "هاست",
    remarks: "Remarks (لینک اشتراک)",
    timeLastUpdate: "تاریخ آخرین به روز رسانی",
    numberServers: "تعداد سرورها",
  },
  operations: {
    name: "عملیات",
    update: "بروزرسانی",
    modify: "تغییر",
    share: "اشتراک گذاری",
    view: "دیدن",
    delete: "حذف",
    create: "ایجاد",
    import: "وارد کردن",
    inBatch: "به صورت دسته ای",
    connect: "اتصال",
    disconnect: "قطع اتصال",
    select: "انتخاب",
    login: "وارد شدن",
    logout: "خارج شدن",
    configure: "پیکربندی شود",
    cancel: "لغو شود",
    saveApply: "ذخیره و اعمال شود",
    confirm: "تایید",
    confirm2: "با دقت تایید شود",
    save: "ذخیره",
    copyLink: "کپی لینک",
    helpManual: "راهنما",
    yes: "بله",
    no: "خیر",
    switchSite: "به سایت جایگزین بروید",
    addOutbound: "یک خروجی اضافه کنید",
  },
  register: {
    title: "ابتدا یک اکانت ایجاد کنید",
    messages: [
      "به یاد داشته باشید که این اکانت برای وارد شدن استفاده خواهد شد.",
      "اکانت بر روی سیستم شما ذخیره می شود و به سرور ارسال نمی شود.",
      "درصورت فراموشی رمز از دستور --reset-password برای بازیابی رمز استفاده کنید.",
    ],
  },
  login: {
    title: "وارد شدن",
    username: "نام کاربری",
    password: "رمز عبور",
  },
  setting: {
    transparentProxy: "Transparent Proxy/System Proxy",
    transparentType: "Transparent Proxy/System Proxy Implementation",
    logLevel: "Log Level",
    pacMode: "Traffic Splitting Mode of Rule Port",
    preventDnsSpoofing: "جلوگیری از هک DNS",
    specialMode: "حالت ویژه",
    mux: "Multiplex",
    autoUpdateSub: "بروزرسانی خودکار اشتراک ها",
    autoUpdateGfwlist: "بروزرسانی خودکار GFWList",
    preferModeWhenUpdate: "Mode when Update Subscriptions and GFWList",
    ipForwardOn: "IP Forward",
    portSharingOn: "به اشتراک گذاری پورت",
    concurrency: "همزمانی",
    options: {
      trace: "Trace",
      debug: "Debug",
      info: "Info",
      warn: "Warn",
      error: "Error",
      global: "عدم تقسیم ترافیک",
      direct: "Direct",
      pac: "بسته به پورت تصمیم گیری شود.",
      whitelistCn: "Proxy except CN Sites",
      gfwlist: "فقط GFWList را پروکسی کن.",
      sameAsPacMode: "Traffic Splitting Mode is the Same as the Rule Port",
      customRouting: "Customized Routing",
      antiDnsHijack: "Prevent DNS Hijack Only (fast)",
      forwardDnsRequest: "Forward DNS Request",
      doh: "DoH(dns-over-https)",
      default: "پیشفرض",
      on: "روشن",
      off: "خاموش",
      updateSubWhenStart: "آپدیت اشتراک ها موقع اجرای نرم افزار",
      updateSubAtIntervals: "آپدیت اشتراک ها هر چند وقت یکبار (واحد: ساعت)",
      updateGfwlistWhenStart: "وقتی نرم افزار اجرا شد GFWList را آپدیت کن",
      updateGfwlistAtIntervals:
        "هر چند وقت یکبار GFWList را آپدیت کن (واحد: ساعت(",
      dependTransparentMode: "Follows Transparent Proxy/System Proxy",
      closed: "خاموش",
      advanced: "تنظیمات پیشرفته",
      leastPing: "اول سرور با پینگ کمتر",
    },
    messages: {
      gfwlist:
        "Based on modified time of file which sometimes is after latest version online.",
      transparentProxy:
        "If transparent proxy on, no extra configure needed and all TCP traffic will pass through the v2rayA. Providing proxy service to other computers and docker as the gateway should make option 'Share in LAN' on.",
      transparentType:
        "★tproxy: support UDP, but not support docker. ★redirect: friendly for docker, but does not support UDP and need to occupy local port 53 for dns anti-pollution.",
      pacMode: `Here you can set the splitting traffic rule of the rule port. By default, "Rule of Splitting Traffic" port is 20172 and HTTP protocol.`,
      preventDnsSpoofing:
        "★Forward DNS Request: DNS requests will be forwarded by proxy server." +
        "★DoH(dns-over-https, v2ray-core: 4.22.0+): DNS over HTTPS.",
      specialMode:
        "★supervisor：Monitor dns pollution, intercept in advance, use the sniffing mechanism of v2ray-core to prevent pollution. ★fakedns：Use the fakens strategy to speed up the resolving.",
      tcpFastOpen:
        "Simplify TCP handshake process to speed up connection establishment. Risk of emphasizing characteristics of packets exists. It may cause failed to connect if your system does not support it.",
      mux:
        "Multiplexing TCP connections to reduce the number of handshake, but it will affect the use cases with high throughput, such as watching videos, downloading, and test speed. " +
        "Risk of emphasizing characteristics of packets exists. Support vmess only now.",
      confirmEgressPorts: `<p>You are setting up transparent proxy across LANs, confirm egress port whitelist.</p>
                          <p>Whitelist:</p>
                          <p>TCP: {tcpPorts}</p>
                          <p>UDP: {udpPorts}</p>`,
      grpcShouldWithTls: `gRPC must be with TLS`,
      ssPluginImpl:
        "★default: 'transport' for simple-obfs, 'chained' for v2ray-plugin." +
        "★chained: shadowsocks traffic will be redirect to standalone plugin." +
        "★transport: processed by the transport layer of v2ray/xray core directly.",
    },
  },
  customAddressPort: {
    title: "آدرس و پورت ها",
    serviceAddress: "آدرس سرویس",
    portSocks5: "پورت SOCKS5",
    portHttp: "پورت HTTP",
    portSocks5WithPac: "پورت SOCKS5(with Rule)",
    portHttpWithPac: "پورت HTTP(with Rule)",
    portVmess: "پورت VMess(with Rule)",
    portVmessLink: "لینک پورت VMess",
    messages: [
      "Service address default as 0.0.0.0:2017 can be changed by setting environment variable <code>V2RAYA_ADDRESS</code> and command argument<code>--address</code>.",
      "If you start v2raya docker container with port mapping instead of <code>--network host</code>, you can remapping ports in this way.",
      "We cannot judge port occupations in docker mode. Confirm it by yourself.",
      "صفر به معنای بستن پورت است.",
    ],
  },
  customRouting: {
    title: "Customize Routing Rule",
    defaultRoutingRule: "پیشفرض Routing Rule",
    sameAsDefaultRule: "the same as default rule",
    appendRule: "اضافه کردن Rule",
    direct: "مستقیم",
    proxy: "پروکسی",
    block: "مسدود کردن",
    rule: "شرط",
    domainFile: "فایل دامنه ها",
    typeRule: "انواع شرط",
    messages: {
      0: "v2rayA تمام اطلاعات درون <b>{V2RayLocationAsset}</b> را بررسی خواهد کرد.",
      1: 'To make a SiteDat file by yourself: <a href="https://github.com/ToutyRater/V2Ray-SiteDAT">ToutyRater/V2Ray-SiteDAT</a>',
      2: "می توانید چند گزینه را انتخاب کنید.",
      noSiteDatFileFound: "siteDat در فایل یافت نشد. {V2RayLocationAsset}",
      emptyRuleNotPermitted: "شرط ها نمی تواند خالی باشد.",
    },
  },
  doh: {
    title: "پیکربندی سرور DoH",
    dohPriorityList: "Priority list of DoH Servers",
    messages: [
      "DoH (DNS over HTTPS) can effectively avoid DNS pollution. But some native DoH providers may themselves be contaminated sometimes. In addition, some DoH services may be blocked by local network providers. Please choose the fastest and stablest DoH server.",
      "Awesome public DoH servers in Mainland China include alidns, geekdns, rubyfish, etc",
      "In taiwan area include quad101, etc",
      "USA: cloudflare, dns.google, etc",
      'Checklist：<a href="https://dnscrypt.info/public-servers" target="_blank">public-servers</a>',
      'Besides, setting up DoH service at your own native server is suggested and well-behaved in most cases <a href="https://github.com/facebookexperimental/doh-proxy" target="_blank">doh-proxy</a>. In this case, it is recommended to run the server(doh-proxy/doh-httpproxy) providing service and client(doh-stub) connecting to doh.opendns.com at the same time and connect them in series, because you can hardly find a server that is not polluted in a generally contaminated region.',
      "Optimally, place one or two lines above. The list will restore to default after saving with empty content.",
    ],
  },
  dns: {
    title: "پیکربندی سرور DNS",
    internalQueryServers: "Domain Query Servers",
    externalQueryServers: "External Domain Query Servers",
    messages: [
      '"@:(dns.internalQueryServers)" are designed to be used to look up domain names in China, while "@:(dns.externalQueryServers)" be used to look up others.',
      '"@:(dns.internalQueryServers)" will be used to look up all domain names if "@:(dns.externalQueryServers)" is empty.',
    ],
  },
  egressPortWhitelist: {
    title: "Egress Port Whitelist",
    tcpPortWhitelist: "TCP پورت های مجاز",
    udpPortWhitelist: "UDP پورت های مجاز",
    messages: [
      "If v2rayA is setup on a server A which connected with a proxy server B, pay attention:",
      "Transparent proxy will force all TCP and UDP traffic to pass through proxy server B, where source IP address will be replaced with proxy B's. Moreover, if some clients send requests to server A that provides service, they will received responses from your proxy B's IP address weirdly, which is illegal.",
      "To resolve it, we need to add those service ports to whitelist so that not pass through proxy.For examples, ssh(22)、v2raya({v2rayaPort}).",
      "Obviously, if the server does not provide any service, you can skip configuring.",
      "Formatting：22 means port 22，20170:20172 means three ports 20170 to 20172.",
    ],
  },
  configureServer: {
    title: "پیکربندی سرور | سرور",
    servername: "نام سرور",
    port: "پورت",
    forceTLS: "forcibly TLS on",
    noObfuscation: "No obfuscation",
    httpObfuscation: "Obfuscated as HTTP",
    srtpObfuscation: "Obfuscated as Video Calls (SRTP)",
    utpObfuscation: "Obfuscated as Bittorrent (uTP)",
    wechatVideoObfuscation: "Obfuscated as Wechat Video Calls",
    dtlsObfuscation: "Obfuscated as DTLS1.2 Packets",
    wireguardObfuscation: "Obfuscated as WireGuard Packets",
    hostObfuscation: "هاست",
    pathObfuscation: "مسیر",
    seedObfuscation: "Seed",
    username: "نام کاربری",
    password: "رمز عبور",
    origin: "origin",
  },
  configureSubscription: {
    title: "پیکربندی اشتراک",
  },
  import: {
    message: "Input a server link or subscription address:",
    batchMessage: "One server link per line:",
    qrcodeError: "Failed to find a valid QRCode, please try again",
  },
  delete: {
    title: "تایید و پاک کردن",
    message:
      "آیا از پاک کردن <b>DELETE</b> اطمینان دارید؟ این عمل غیرقابل بازگشت است.",
  },
  latency: {
    message: "تست latency ممکن است چند دقیقه طول بکشد. لطفا صبور باشید.",
  },
  version: {
    higherVersionNeeded: "این عملیات به نسخه بالاتر از {version} نیاز دارد.",
    v2rayInvalid:
      "geosite.dat, geoip.dat یا v2ray-core ممکن است به درستی نصب نشده باشد.",
    v2rayNotV5:
      "ورژن v2ray-core برار v5 نیست. از v5 استفاده کنید یا به v2rayA نسخه v1.5 بازگردانی کنید.",
  },
  about: `<p>v2rayA یک رابطه کاربری برای V2Ray است.</p>
          <p class="about-small">پورت های پیشفرض:</p>
          <p class="about-small">2017: v2rayA service port</p>
          <p class="about-small">20170: SOCKS پروتکل</p>
          <p class="about-small">20171: HTTP پروتکل</p>
          <p class="about-small">20172: HTTP protocol with "Rule of Splitting Traffic"</p>
          <p class="about-small">دیگر پورت ها：</p>
          <p class="about-small">32345: tproxy, needed by transparent proxy </p>
          <p>تمام اطلاعات به صورت محلی ذخیره میشود نه فضای ابری</p>
          <p>می توانید مشکلات را به لینک روبرو گزارش دهید : <a href="https://github.com/v2rayA/v2rayA/issues">issues</a>.</p>
          <p>داکیومنت ها : <a href="https://v2raya.org">https://v2raya.org</a></p>`,
  axios: {
    messages: {
      optimizeBackend: "Adjust v2rayA service address？",
      noBackendFound:
        "نمی توان v2rayA را در آدرس {url} پیدا کرد. اطمینان حاصل کنید که v2rayA روی این آدرس درحال اجرا است.",
      cannotCommunicate: [
        "Cannot communicate. If your service is running and ports open correctly, the reason may be that current browser does not allow https sites to access http resources, you can try using Chrome or switching to alternate http site.",
        "Cannot communicate. Firefox does not allow https sites to access http resources, you can try switching to alternate http sites.",
      ],
    },
    urls: {
      usage: "https://github.com/v2rayA/v2rayA/wiki/Usage",
    },
  },
  routingA: {
    messages: ["برای کمک برو روی گزینه 'Help&Manual' کلیک کنید."],
  },
  outbound: {
    addMessage: "لطفاً نام خروجی را که می خواهید اضافه کنید وارد کنید:",
    deleteMessage:
      'از حذف <b>DELETE</b> اطمینان دارید "{outboundName}"? این عمل غیرقابل بازگشت است',
  },
  log: {
    logModalTitle: "دیدن لاگ ها",
    logsLabel: "لاگ ها",
    refreshInterval: "Refresh Interval",
    seconds: "ثانیه",
    autoScoll: "Auto Scroll",
    category: "Category",
    categories: {
      all: "All",
      error: "Error",
      warn: "Warn",
      info: "Info",
      debug: "Debug",
      trace: "Trace",
      other: "Other",
    },
  },
};
