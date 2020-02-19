export default {
  common: {
    setting: "设置",
    about: "关于",
    loggedAs: "正在以 <b>{username}</b> 的身份登录",
    v2rayCoreStatus: "v2ray-core状态",
    checkRunning: "检测中",
    isRunning: "正在运行",
    notRunning: "尚未运行",
    notLogin: "未登录",
    latest: "最新",
    local: "本地"
  },
  welcome: {
    docker: "V2RayA服务端正在运行于Docker环境中，Version: {version}",
    default: "V2RayA服务端正在运行，Version: {version}",
    newVersion: "检测到新版本: {version}"
  },
  v2ray: {
    start: "启动",
    stop: "关闭"
  },
  server: {
    name: "节点名",
    address: "节点地址",
    protocol: "协议",
    latency: "时延"
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
    connect: "连接",
    disconnect: "断开",
    login: "登录",
    logout: "注销",
    configure: "配置",
    cancel: "取消",
    confirm: "确定",
    saveApply: "保存并应用",
    save: "保存",
    copyLink: "复制链接"
  },
  register: {
    title: "初来乍到，创建一个管理员账号"
  },
  login: {
    title: "登录",
    username: "用户名",
    password: "密码"
  },
  setting: {
    transparentProxy: "全局透明代理",
    pacMode: "PAC模式",
    preventDnsSpoofing: "防止DNS污染",
    mux: "多路复用",
    autoUpdateSub: "自动更新订阅",
    autoUpdateGfwlist: "自动更新GFWList",
    preferModeWhenUpdate: "解析订阅链接/更新时优先使用",
    ipForwardOn: "开启IP转发",
    concurrency: "最大并发数",
    options: {
      global: "代理所有流量",
      direct: "直连模式",
      pac: "PAC模式",
      whitelistCn: "大陆白名单",
      gfwlist: "GFWList",
      sameAsPacMode: "与PAC模式一致",
      customRouting: "自定义路由规则",
      antiDnsHijack: "仅防止DNS劫持",
      forwardDnsRequest: "防止DNS污染：转发DNS请求",
      doh: "防止DNS污染：DoH(DNS-over-HTTPS)",
      default: "保持系统默认",
      on: "启用",
      off: "关闭",
      updateSubWhenStart: "服务端启动时更新订阅",
      updateGfwlistWhenStart: "服务端启动时更新GFWList",
      dependTransparentMode: "跟随全局透明代理"
    }
  },
  customAddressPort: {
    title: "地址与端口",
    serviceAddress: "服务端地址",
    portSocks5: "socks5端口",
    portHttp: "http端口",
    portHttpWithPac: "http端口(PAC模式)",
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
    messages: [
      "将SiteDat文件放于 <b>{V2RayLocationAsset}</b> 目录下，V2rayA将自动进行识别",
      '制作SiteDat文件：<a href="https://github.com/ToutyRater/V2Ray-SiteDAT">ToutyRater/V2Ray-SiteDAT</a>',
      "在选择Tags时，可按Ctrl等多选键进行多选。"
    ]
  },
  doh: {
    title: "配置DoH服务器",
    dohPriorityList: "DoH服务优先级列表",
    messages: [
      "DoH即DNS over HTTPS，能够有效避免DNS污染，但一些DoH提供商的DoH服务可能被墙，请自行选择非代理条件下直连速度最快的DoH提供商",
      "大陆较好的DoH服务有geekdns: 233py.com、红鱼: rubyfish.cn等",
      "台湾有quad101: dns.twnic.tw等",
      "美国有cloudflare: 1.0.0.1等",
      '清单：<a href="https://dnscrypt.info/public-servers" target="_blank">public-servers</a>',
      '另外，您可以在未受到DNS污染的国内服务器上自架DoH服务，以纵享丝滑。<a href="https://dnscrypt.info/implementations" target="_blank">Server Implementations</a>',
      "建议上述列表1-2行即可，留空保存可恢复默认"
    ]
  },
  egressPortWhitelist: {
    title: "出方向端口白名单",
    tcpPortWhitelist: "TCP端口白名单",
    udpPortWhitelist: "UDP端口白名单",
    messages: [
      "如果你将V2RayA架设在对外提供服务的服务器上，那么你需要注意：",
      "全局透明代理会使得所有TCP、UDP流量走代理，通过走代理的流量其源IP地址会被替换为代理服务器的IP地址，那么如果客户请求你的服务器IP地址，他却将得到从你代理服务器IP发出的回答，该回答在客户看来无疑是不合法的，从而导致服务被拒绝。",
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
    password: "密码"
  }
};
