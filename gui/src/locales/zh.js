export default {
  common: {
    setting: "设置",
    about: "关于",
    loggedAs: "正在以 <b>{username}</b> 的身份登录",
    v2rayCoreStatus: "v2ray-core状态",
    checkRunning: "检测中",
    isRunning: "正在运行",
    notRunning: "就绪",
    notLogin: "未登录",
    latest: "最新",
    local: "本地",
    success: "成功",
    fail: "失败",
    message: "提示",
    none: "无",
    optional: "可选",
    loadBalance: "负载均衡",
    log: "日志"
  },
  welcome: {
    title: "初来乍到，请多关照",
    docker: "v2rayA服务端正在运行于Docker环境中，Version: {version}",
    default: "v2rayA服务端正在运行，Version: {version}",
    newVersion: "检测到新版本: {version}",
    messages: [
      "我们发现你还没有创建或导入任何节点、订阅。",
      "我们支持以节点地址或者订阅地址的方式导入，也支持手动创建节点，快来试试吧！"
    ]
  },
  v2ray: {
    start: "启动",
    stop: "关闭"
  },
  server: {
    name: "节点名",
    address: "节点地址",
    protocol: "协议",
    latency: "时延",
    lastSeenTime: "上次存活时间",
    lastTryTime: "上次测试时间",
    messages: {
      notAllowInsecure:
        "根据 {name} 的官方文档，如果你使用 {name}，将不允许 AllowInsecure",
      notRecommend:
        "根据 {name} 的官方文档，如果你使用 {name}，不推荐开启 AllowInsecure"
    }
  },
  InSecureConfirm: {
    title: "检测到不安全的配置",
    message:
      "即将保存的配置中<b>AllowInsecure</b>被设置为true，除非你知道你在做什么，否则贸然开启可能导致数据泄漏！是否继续？",
    confirm: "我知道我在做什么",
    cancel: "取消"
  },
  subscription: {
    host: "域名",
    remarks: "别名",
    timeLastUpdate: "上次更新时间",
    numberServers: "节点数"
  },
  operations: {
    name: "操作",
    update: "更新",
    modify: "修改",
    share: "分享",
    view: "查看",
    delete: "删除",
    create: "创建",
    import: "导入",
    inBatch: "批量",
    connect: "连接",
    disconnect: "断开",
    select: "选择",
    login: "登录",
    logout: "注销",
    configure: "配置",
    cancel: "取消",
    confirm: "确定",
    confirm2: "确认无误",
    saveApply: "保存并应用",
    save: "保存",
    copyLink: "复制链接",
    helpManual: "查看帮助",
    yes: "是",
    no: "否",
    switchSite: "切换至备用站点",
    addOutbound: "新增一个出站 (outbound)"
  },
  register: {
    title: "初来乍到，创建一个管理员账号",
    messages: [
      "请记住您创建的管理员账号，用于登录该管理页面。",
      "账号信息位于本地，我们不会上传任何信息到服务器。",
      "如不慎忘记密码，使用 v2raya --reset-password 重置。"
    ]
  },
  login: {
    title: "登录",
    username: "用户名",
    password: "密码"
  },
  setting: {
    transparentProxy: "透明代理/系统代理",
    transparentType: "透明代理/系统代理实现方式",
    pacMode: "规则端口的分流模式",
    preventDnsSpoofing: "防止DNS污染",
    specialMode: "特殊模式",
    mux: "多路复用",
    autoUpdateSub: "自动更新订阅",
    autoUpdateGfwlist: "自动更新GFWList",
    preferModeWhenUpdate: "解析订阅链接/更新时优先使用",
    ipForwardOn: "开启IP转发",
    portSharingOn: "开启端口分享",
    concurrency: "最大并发数",
    options: {
      global: "不进行分流",
      direct: "直连模式",
      pac: "跟随规则端口",
      whitelistCn: "大陆白名单模式",
      gfwlist: "GFWList模式",
      sameAsPacMode: "分流规则与规则端口所选模式一致",
      customRouting: "自定义路由规则",
      antiDnsHijack: "仅防止DNS劫持(快速)",
      forwardDnsRequest: "转发DNS请求",
      doh: "DoH(DNS-over-HTTPS)",
      default: "保持系统默认",
      on: "启用",
      off: "关闭",
      updateSubWhenStart: "服务端启动时更新订阅",
      updateSubAtIntervals: "每隔一段时间更新订阅（单位：小时）",
      updateGfwlistWhenStart: "服务端启动时更新GFWList",
      updateGfwlistAtIntervals: "每隔一段时间更新GFWList（单位：小时）",
      dependTransparentMode: "跟随透明代理/系统代理",
      closed: "关闭",
      advanced: "自定义高级设置"
    },
    messages: {
      gfwlist: "该时间是指本地文件最后修改时间，因此可能会领先最新版本",
      transparentProxy:
        "全局代理开启后，无需经过额外设置，任何TCP流量均会经过V2RayA。另外，如需作为网关使得连接本机的其他主机或docker也享受代理，请勾选“开启局域网共享”。",
      transparentType:
        "★tproxy: 支持udp，不支持docker。★redirect: docker友好，不支持udp，需要占用本地53端口以应对dns污染。",
      pacMode:
        "该选项设置规则分流端口所使用的路由模式。默认情况下规则分流端口为20172，HTTP协议。",
      preventDnsSpoofing:
        "★转发DNS查询: 通过代理服务器转发DNS请求。" +
        "★DoH(v2ray-core: 4.22.0+): DNS over HTTPS。",
      specialMode:
        "★supervisor：监控dns污染，提前拦截，利用v2ray-core的sniffing解决污染。★fakedns：使用fakedns策略加速解析。",
      tcpFastOpen:
        "简化TCP握手流程以加速建立连接，可能会增加封包的特征。若系统不支持可能会导致无法正常连接。",
      mux:
        "复用TCP连接以减少握手次数，但会影响吞吐量大的使用场景，如观看视频、下载、测速。当前仅支持vmess节点。可能会增加特征造成断流。",
      confirmEgressPorts: `<p>您正在对不同子网下的机器设置透明代理，请确认不走代理的出方向端口。</p>
                          <p>当前设置的端口白名单为：</p>
                          <p>TCP: {tcpPorts}</p>
                          <p>UDP: {udpPorts}</p>`,
      xtlsNotWithWs: `xtls无法和websocket共存`,
      grpcShouldWithTls: `gRPC必须启用TLS`,
      ssPluginImpl:
        "★默认：使用 simple-obfs 时为等效传输层，v2ray-plugin 时为链式。" +
        "★链式：shadowsocks 流量会被转发至独立的插件。" +
        "★等效传输层：直接由 v2ray/xray 核心的传输层处理。"
    }
  },
  customAddressPort: {
    title: "地址与端口",
    serviceAddress: "服务端地址",
    portSocks5: "socks5端口",
    portHttp: "http端口",
    portSocks5WithPac: "socks5端口(带分流规则)",
    portHttpWithPac: "http端口(带分流规则)",
    portVlessGrpc: "VLESS-GRPC端口(带分流规则)",
    portVlessGrpcLink: "VLESS-GRPC端口链接",
    messages: [
      "如需修改后端运行地址(默认0.0.0.0:2017)，可添加环境变量<code>V2RAYA_ADDRESS</code>或添加启动参数<code>--address</code>。",
      "docker模式下如果未使用<code>--privileged --network host</code>参数启动容器，可通过修改端口映射修改socks5、http端口。",
      "docker模式下不能正确判断端口占用，请确保输入的端口未被其他程序占用。",
      "如将端口设为0则表示关闭该端口。"
    ]
  },
  customRouting: {
    title: "自定义路由规则",
    defaultRoutingRule: "默认路由规则",
    sameAsDefaultRule: "与默认规则相同",
    appendRule: "追加规则",
    direct: "直连",
    proxy: "代理",
    block: "拦截",
    rule: "规则",
    domainFile: "域名文件",
    typeRule: "规则类型",
    messages: {
      0: "将SiteDat文件放于 <b>{V2RayLocationAsset}</b> 目录下，V2rayA将自动进行识别",
      1: '制作SiteDat文件：<a href="https://github.com/ToutyRater/V2Ray-SiteDAT">ToutyRater/V2Ray-SiteDAT</a>',
      2: "在选择Tags时，可按Ctrl等多选键进行多选。",
      noSiteDatFileFound: "未在{V2RayLocationAsset}中发现siteDat文件",
      emptyRuleNotPermitted: "不能存在tags为空的规则，请检查"
    }
  },
  doh: {
    title: "配置DoH服务器",
    dohPriorityList: "DoH服务优先级列表",
    messages: [
      "DoH即DNS over HTTPS，能够有效避免DNS污染，一些国内的DoH提供商自己本身也有被污染的可能，另外，一些DoH提供商的DoH服务可能被墙，请自行选择非代理条件下直连速度较快且效果较好的DoH提供商",
      "大陆较好的DoH服务有阿里dns, geekdns, rubyfish等",
      "台湾有quad101: dns.twnic.tw等",
      "美国有cloudflare: 1.0.0.1等",
      '清单：<a href="https://dnscrypt.info/public-servers" target="_blank">public-servers</a>',
      '另外，您可以在自己的国内服务器上自架DoH服务，以纵享丝滑 <a href="https://github.com/facebookexperimental/doh-proxy" target="_blank">doh-proxy</a>。在这种情况下，建议同时运行服务端(doh-proxy/doh-httpproxy, 提供对外服务)和客户端(doh-stub, 连接至doh.opendns.com)并将他们串联，因为通常情况下在普遍受到污染的地区找到一个完全不受污染的服务器是很困难的。',
      "建议上述列表1-2行即可，留空保存可恢复默认"
    ]
  },
  dns: {
    title: "配置DNS服务器",
    internalQueryServers: "域名查询服务器",
    externalQueryServers: "国外域名查询服务器",
    messages: [
      "“@:(dns.internalQueryServers)” 用于查询国内域名，而 “@:(dns.externalQueryServers)” 用于查询国外域名。",
      "如果将 “@:(dns.externalQueryServers)” 留空，“@:(dns.internalQueryServers)” 将会负责查询所有域名。"
    ]
  },
  egressPortWhitelist: {
    title: "出方向端口白名单",
    tcpPortWhitelist: "TCP端口白名单",
    udpPortWhitelist: "UDP端口白名单",
    messages: [
      "如果你将v2rayA架设在对外提供服务的服务器A上，连接了代理服务器B，那么你需要注意：",
      "透明代理会使得所有TCP、UDP流量走代理，通过走代理的流量其源IP地址会被替换为代理服务器B的IP地址，那么如果有客户向你的服务器A发出请求，他却将得到从你代理服务器B发出的回答，该回答在客户看来无疑是不合法的，从而导致连接被拒绝。",
      "因此，需要将服务器提供的对外服务端口包含在白名单中，使其不走代理。如ssh(22)、v2raya({v2rayaPort})。",
      "如不对外提供服务或仅对局域网内主机提供服务，则可不设置白名单。",
      "格式：22表示端口22，20170:20172表示20170到20172三个端口。"
    ]
  },
  configureServer: {
    title: "配置节点 | 节点",
    servername: "节点名称",
    port: "端口号",
    forceTLS: "强制开启TLS",
    noObfuscation: "不伪装",
    httpObfuscation: "伪装为HTTP",
    srtpObfuscation: "伪装视频通话(SRTP)",
    utpObfuscation: "伪装为BT下载(uTP)",
    wechatVideoObfuscation: "伪装为微信视频通话",
    dtlsObfuscation: "伪装为DTLS1.2数据包",
    wireguardObfuscation: "伪装为WireGuard数据包",
    hostObfuscation: "域名(host)",
    pathObfuscation: "路径(path)",
    seedObfuscation: "混淆种子",
    username: "用户名",
    password: "密码",
    origin: "原版"
  },
  configureSubscription: {
    title: "订阅配置"
  },
  import: {
    message: "填入节点链接或订阅地址：",
    batchMessage: "一行一个节点链接:",
    qrcodeError: "找不到有效的二维码，请重新尝试"
  },
  delete: {
    title: "确认删除",
    message: "确定要<b>删除</b>这些节点/订阅吗？注意，该操作是不可逆的。"
  },
  latency: {
    message: "时延测试往往需要花费较长时间，请耐心等待"
  },
  version: {
    higherVersionNeeded: "该操作需要 v2rayA 的版本高于{version}",
    v2rayInvalid:
      "检测到 geosite.dat, geoip.dat 文件或 v2ray-core 可能未正确安装，请检查",
    lowCoreVersion: "v2ray-core 版本过低，可能会出现非预期错误"
  },
  about: `<p>v2rayA是V2Ray的一个Web客户端，前端使用Vue.js构建，后端使用Golang构建。</p>
          <p class="about-small">默认端口：</p>
          <p class="about-small">2017: v2rayA后端端口</p>
          <p class="about-small">20170: SOCKS协议</p>
          <p class="about-small">20171: HTTP协议</p>
          <p class="about-small">20172: 带分流规则的HTTP协议</p>
          <p class="about-small">其他端口：</p>
          <p class="about-small">32345: tproxy，透明代理所需 </p>
          <p class="about-small">32346: 插件协议端口，如trojan、ssr和pingtunnel</p>
          <p>在使用中如果发现任何问题，欢迎<a href="https://github.com/v2rayA/v2rayA/issues">提出issue</a>。</p>`,
  axios: {
    messages: {
      optimizeBackend: "您是否需要调整服务端地址？",
      noBackendFound: "未在 {url} 检测到v2rayA服务端，请确定v2rayA正常运行",
      cannotCommunicate: [
        "无法通信。如果您的服务端已正常运行，且端口正常开放，原因可能是当前浏览器不允许https站点访问http资源，您可以尝试切换为http备用站点。",
        "无法通信。火狐浏览器不允许https站点访问http资源，您可以换用Chrome浏览器或切换为http备用站点。"
      ]
    },
    urls: {
      usage:
        "https://github.com/v2rayA/v2rayA/wiki/%E4%BD%BF%E7%94%A8%E6%96%B9%E6%B3%95"
    }
  },
  routingA: {
    messages: ["点击“查看帮助”按钮以获取帮助"]
  },
  outbound: {
    addMessage: "请输入你想要添加的出站(outbound)名称：",
    deleteMessage:
      '确定要<b>删除</b>出站 "{outboundName}" 吗？注意，该操作是不可逆的。'
  },
  driver: {
    welcome: [
      "首先导入节点服务器",
      "初次使用，还没有任何节点服务器，在这里导入或创建节点服务器。"
    ],
    tabs: [
      "订阅与节点服务器",
      "导入订阅、节点服务器后，在这里切换和管理你的订阅、普通节点以及订阅节点。"
    ]
  },
  log: {
    logModalTitle: "查看日志",
    refreshInterval: "刷新间隔",
    seconds: "秒"
  }
};
