export default {
  common: {
    setting: "Setting",
    about: "About",
    loggedAs: "Logged as <b>{username}</b>",
    v2rayCoreStatus: "Status of v2ray-core",
    checkRunning: "Checking",
    isRunning: "Running",
    notRunning: "Stopped",
    notLogin: "Login please",
    latest: "Latest",
    local: "Local"
  },
  welcome: {
    docker: "V2RayA service is running in Docker，Version: {version}",
    default: "V2RayA service is running，Version: {version}",
    newVersion: "Detected new version: {version}"
  },
  v2ray: {
    start: "Start",
    stop: "Stop"
  },
  server: {
    name: "Server Name",
    address: "Server Address",
    protocol: "Protocol",
    latency: "Latency"
  },
  subscription: {
    host: "Host",
    remarks: "Remarks",
    timeLastUpdate: "Datetime of Last Update",
    numberServers: "Number of Servers"
  },
  operations: {
    name: "Operations",
    update: "Update",
    modify: "Modify",
    share: "Share",
    view: "View",
    delete: "Delete",
    create: "Create",
    import: "Import",
    connect: "Connect",
    disconnect: "Disconnect",
    login: "Login",
    logout: "Logout",
    configure: "Configure",
    cancel: "Cancel",
    saveApply: "Save and Apply",
    confirm: "Confirm",
    save: "Save",
    copyLink: "COPY LINK"
  },
  register: {
    title: "Nice to meet you! Create an admin account now"
  },
  login: {
    title: "Login",
    username: "Username",
    password: "Password"
  },
  setting: {
    transparentProxy: "Transparent Proxy",
    pacMode: "PAC Mode",
    preventDnsSpoofing: "Prevent DNS Spoofing",
    mux: "Multiplex",
    autoUpdateSub: "Automatically Update Subscriptions",
    autoUpdateGfwlist: "Automatically Update GFWList",
    preferModeWhenUpdate: "Mode when Upadate Subscriptions and GFWList",
    ipForwardOn: "Setup IP Forward",
    concurrency: "Concurrency",
    options: {
      global: "Proxy All Traffic",
      direct: "Direct",
      pac: "PAC Mode",
      whitelistCn: "Proxy except CN Sites",
      gfwlist: "Proxy Only GFWList",
      sameAsPacMode: "The Same as PAC Mode",
      customRouting: "Customized Routing",
      antiDnsHijack: "Prevent DNS Hijack Only",
      forwardDnsRequest: "Prevent DNS Spoofing: Forward DNS Request",
      doh: "Prevent DNS Spoofing: DoH(dns-over-https)",
      default: "Keep Default",
      on: "On",
      off: "Off",
      updateSubWhenStart: "Update Subscriptions When Service Starts",
      updateGfwlistWhenStart: "Update GFWList When Service Starts",
      dependTransparentMode: "Depend on Transparent Mode"
    }
  },
  customAddressPort: {
    title: "Address and Ports",
    serviceAddress: "Address of Service",
    portSocks5: "Port of SOCKS5",
    portHttp: "Port of HTTP",
    portHttpWithPac: "Port of HTTP(with PAC)",
    messages: [
      "Service address default as 0.0.0.0:2017 can be changed by setting environment variable <code>V2RAYA_ADDRESS</code> and command argument<code>--address</code>.",
      "If you start v2raya docker container with port mapping instead of <code>--network host</code>, you can remapping ports in this way.",
      "We can not judge port occupations in docker mode. Confirm ports are free.",
      "Put zero means to close this port."
    ]
  },
  customRouting: {
    title: "Customize Routing Rule",
    defaultRoutingRule: "Default Routing Rule",
    sameAsDefaultRule: "the same as default rule",
    appendRule: "Append Rule",
    direct: "Direct",
    proxy: "Proxy",
    block: "Block",
    rule: "Rule",
    domainFile: "Domain File",
    typeRule: "Type of Rule",
    messages: [
      "V2RayA will recognize all SiteDat file in <b>{V2RayLocationAsset}</b>",
      'To make a SiteDat file by yourself: <a href="https://github.com/ToutyRater/V2Ray-SiteDAT">ToutyRater/V2Ray-SiteDAT</a>',
      "Multi-select is supported."
    ]
  },
  doh: {
    title: "Configure DoH Server",
    dohPriorityList: "Priority list of DoH Servers",
    messages: [
      "DoH(DNS over HTTPS) can be used to prevent dns spoofing, but some public DoH servers are not access in specific areas. You should choose the fastest and stablest DoH server.",
      "Awesome public DoH servers in Mainland China include geekdns, rubyfish, etc",
      "In taiwan area include quad101, etc",
      "USA: cloudflare, dns.google, etc",
      'Checklist：<a href="https://dnscrypt.info/public-servers" target="_blank">public-servers</a>',
      'Besides, setting up DoH service at your own server doesn\'t suffer dns spoofing is suggested and well-behaved int most cases. <a href="https://dnscrypt.info/implementations" target="_blank">Server Implementations</a>',
      "Optimally, place one or two lines above. The list will restore to default after saving with empty content."
    ]
  },
  egressPortWhitelist: {
    title: "Egress Port Whitelist",
    tcpPortWhitelist: "TCP Port Whitelist",
    udpPortWhitelist: "UDP Port Whitelist",
    messages: [
      "If V2RayA is setup on a server providing service to clients, pay attention:",
      "Transparent proxy will force all TCP and UDP traffic to pass through proxy server, where source IP address will be replaced with IP address of proxy server. Moreover, if some clients send requests to your server's IP, they will received responses from your proxy server's IP weirdly, which is illegal.",
      "To resolve it, we need to add those service ports to whitelist so that not pass through proxy.For examples, ssh(22)、v2raya({v2rayaPort}).",
      "Obviously, if the server does not provide any service, you can skip configuring.",
      "Formatting：22 means port 22，20170:20172 means three ports 20170 to 20172."
    ]
  },
  configureServer: {
    title: "Configure Server | Server",
    servername: "Servername",
    port: "Port",
    forceTLS: "forcibly TLS on",
    noObfuscation: "No obfuscation",
    httpObfuscation: "Obfuscated as HTTP",
    srtpObfuscation: "Obfuscated as Video Calls (SRTP)",
    utpObfuscation: "Obfuscated as Bittorrent (uTP)",
    wechatVideoObfuscation: "Obfuscated as Wechat Video Calls",
    dtlsObfuscation: "Obfuscated as DTLS1.2 Packets",
    wireguardObfuscation: "Obfuscated as WireGuard Packets",
    hostObfuscation: "Host",
    pathObfuscation: "Path",
    password: "Password"
  },
  configureSubscription: {
    title: "Configure Subscription"
  }
};
