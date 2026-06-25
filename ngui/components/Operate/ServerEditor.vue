<script lang="ts" setup>
import { ref, reactive, watch, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance } from 'element-plus'

const { t } = useI18n()

// ==================== Props & Emits ====================
const props = defineProps<{
  modelValue: boolean          // dialog visibility
  server?: any                 // { id, _type, sub } for edit mode
  readonly?: boolean           // view-only mode
}>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
  'save': []
}>()

// ==================== Utility: Base64 ====================
function utf8ToBase64(str: string): string {
  return btoa(encodeURIComponent(str).replace(/%([0-9A-F]{2})/g, (_, p1) =>
    String.fromCharCode(Number('0x' + p1))
  ))
}

function base64ToUtf8(str: string): string {
  return decodeURIComponent(atob(str).split('').map(c =>
    '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2)
  ).join(''))
}

function base64EncodeURI(str: string): string {
  return utf8ToBase64(str).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
}

function base64DecodeURI(str: string): string {
  str = str.replace(/-/g, '+').replace(/_/g, '/')
  while (str.length % 4) str += '='
  return base64ToUtf8(str)
}

// ==================== Utility: URL Parser ====================
interface ParsedURL {
  source: string
  username: string
  password: string
  protocol: string
  host: string
  port: number | string
  query: string
  params: Record<string, string>
  hash: string
  path: string
}

function parseURL(u: string): ParsedURL {
  let url = u
  let protocol = ''
  let fakeProto = false

  if (url.indexOf('://') === -1) {
    url = 'http://' + url
  } else {
    protocol = url.substring(0, url.indexOf('://'))
    switch (protocol) {
      case 'http':
      case 'https':
      case 'ws':
      case 'wss':
        break
      default:
        url = 'http' + url.substring(url.indexOf('://'))
        fakeProto = true
    }
  }

  const a = document.createElement('a')
  a.href = url

  const result: ParsedURL = {
    source: u,
    username: a.username,
    password: a.password,
    protocol: fakeProto ? protocol : a.protocol.replace(':', ''),
    host: a.hostname,
    port: a.port ? parseInt(a.port) : (protocol === 'https' || protocol === 'wss' ? 443 : 80),
    query: a.search,
    params: (() => {
      const ret: Record<string, string> = {}
      const seg = a.search.replace(/^\?/, '').split('&')
      for (const s of seg) {
        if (!s) continue
        const eq = s.indexOf('=')
        if (eq === -1) continue
        const key = s.substring(0, eq)
        const val = decodeURIComponent(s.substring(eq + 1))
        if (ret[key]) {
          if (Array.isArray(ret[key])) {
            (ret[key] as any).push(val)
          } else {
            ret[key] = [ret[key], val] as any
          }
        } else {
          ret[key] = val
        }
      }
      return ret
    })(),
    hash: a.hash.replace('#', ''),
    path: a.pathname.replace(/^([^/])/, '/$1'),
  }
  a.remove()
  return result
}

function generateURL({ username, password, protocol, host, port, params, hash, path }: {
  protocol: string
  username?: string
  password?: string
  host?: string
  port?: string | number
  params?: Record<string, any>
  hash?: string
  path?: string
}): string {
  let url = `${protocol || 'http'}://`
  if (username) {
    url += encodeURIComponent(username)
    if (password) url += ':' + encodeURIComponent(password)
    url += '@'
  }
  url += host || ''
  if (port && port !== 80 && port !== 443) url += ':' + port
  if (path) url += path.startsWith('/') ? path : '/' + path
  if (params && Object.keys(params).length > 0) {
    const query = Object.entries(params)
      .filter(([, v]) => v !== '' && v !== false && v !== undefined && v !== null)
      .map(([k, v]) => `${k}=${encodeURIComponent(String(v))}`)
      .join('&')
    if (query) url += '?' + query
  }
  if (hash) url += '#' + encodeURIComponent(hash)
  return url
}

// ==================== Variant ====================
function variant(): string {
  if (typeof window !== 'undefined')
    return localStorage.getItem('variant')?.toLowerCase() || 'xray'
  return 'xray'
}

// ==================== Form Data ====================
const tabChoice = ref(0)

const v2rayFormRef = ref<FormInstance>()
const wireguardFormRef = ref<FormInstance>()
const ssFormRef = ref<FormInstance>()
const ssrFormRef = ref<FormInstance>()
const trojanFormRef = ref<FormInstance>()
const juicityFormRef = ref<FormInstance>()
const tuicFormRef = ref<FormInstance>()
const hysteria2FormRef = ref<FormInstance>()
const httpFormRef = ref<FormInstance>()
const socks5FormRef = ref<FormInstance>()
const anytlsFormRef = ref<FormInstance>()

// V2Ray (VMess / VLESS)
const v2ray = reactive({
  ps: '',
  add: '',
  port: '',
  id: '',
  flow: '',
  aid: '',
  net: 'tcp',
  type: 'none',
  host: '',
  path: '',
  tls: 'none',
  quicSecurity: 'none',
  fp: '',
  pbk: '',
  sid: '',
  spx: '',
  alpn: '',
  sni: '',
  scy: 'auto',
  v: '',
  pinnedPeerCertSha256: '',
  verifyPeerCertByName: '',
  protocol: 'vmess' as 'vmess' | 'vless',
  key: 'none',
  xhttpMode: 'auto',
  xhttpRawJson: '',
  maxEarlyData: '',
  earlyDataHeaderName: '',
  multiMode: false,
  idleTimeout: '',
  healthCheckTimeout: '',
  permitWithoutStream: false,
  initialWindowsSize: '',
})

// WireGuard
const wireguard = reactive({
  name: '',
  address: '',
  port: '',
  publicKey: '',
  privateKey: '',
  localAddress: '',
  dns: '',
  mtu: '',
  allowedIPs: '',
  persistentKeepalive: '',
  preSharedKey: '',
  endpoint: '',
})

// Shadowsocks
const ss = reactive({
  method: '2022-blake3-aes-128-gcm',
  plugin: '',
  obfs: 'http',
  tls: '',
  path: '/',
  mode: 'websocket',
  host: '',
  password: '',
  server: '',
  port: '',
  name: '',
  protocol: 'ss',
  impl: '',
  backend: '',
  plugin_opts: '',
})

// ShadowsocksR
const ssr = reactive({
  method: 'aes-128-cfb',
  password: '',
  server: '',
  port: '',
  name: '',
  proto: 'origin',
  protoParam: '',
  obfs: 'plain',
  obfsParam: '',
  protocol: 'ssr',
})

// Trojan
const trojan = reactive({
  name: '',
  server: '',
  peer: '',
  host: '',
  path: '',
  pinnedPeerCertSha256: '',
  verifyPeerCertByName: '',
  port: '',
  password: '',
  method: 'origin' as 'origin' | 'shadowsocks',
  ssCipher: 'aes-128-gcm',
  ssPassword: '',
  net: 'tcp',
  obfs: 'none',
  protocol: 'trojan',
  backend: '',
  alpn: '',
})

// Juicity
const juicity = reactive({
  name: '',
  server: '',
  port: '',
  sni: '',
  cc: 'bbr',
  uuid: '',
  password: '',
  pinnedCertchainSha256: '',
  fingerprint: '',
  protocol: 'juicity',
})

// TUIC
const tuic = reactive({
  name: '',
  server: '',
  port: '',
  sni: '',
  cc: 'bbr',
  uuid: '',
  password: '',
  pinnedPeerCertSha256: '',
  verifyPeerCertByName: '',
  disableSni: false,
  alpn: 'h3',
  fingerprint: '',
  udpRelayMode: 'native',
  protocol: 'tuic',
})

// Hysteria2
const hysteria2 = reactive({
  name: '',
  server: '',
  port: '',
  password: '',
  sni: '',
  obfs: 'none',
  obfsPassword: '',
  pinnedPeerCertSha256: '',
  verifyPeerCertByName: '',
  up: '',
  down: '',
  protocol: 'hysteria2',
})

// HTTP
const http = reactive({
  username: '',
  password: '',
  host: '',
  port: '',
  protocol: 'http',
  name: '',
})

// SOCKS5
const socks5 = reactive({
  username: '',
  password: '',
  host: '',
  port: '',
  protocol: 'socks5',
  name: '',
})

// AnyTLS
const anytls = reactive({
  name: '',
  host: '',
  port: '',
  auth: '',
  sni: '',
  alpn: '',
  fingerprint: '',
  pinnedPeerCertSha256: '',
  verifyPeerCertByName: '',
  protocol: 'anytls',
})

// ==================== Form Validation Rules ====================
const requiredRule = { required: true, message: t('configureServer.servername'), trigger: 'blur' }

// ==================== URL Resolver (parse sharing URLs -> form data) ====================
function resolveURL(url: string): any {
  const lower = url.toLowerCase()

  if (lower.startsWith('vmess://')) {
    let obj = JSON.parse(base64ToUtf8(url.substring(url.indexOf('://') + 3)))
    obj.ps = decodeURIComponent(obj.ps)
    obj.tls = obj.tls || 'none'
    obj.type = obj.type || 'none'
    obj.scy = obj.scy || 'auto'
    obj.protocol = obj.protocol || 'vmess'
    return obj
  }

  if (lower.startsWith('vless://')) {
    let u = parseURL(url)
    const o: any = {
      ps: decodeURIComponent(u.hash),
      add: u.host,
      port: u.port,
      id: decodeURIComponent(u.username),
      flow: u.params.flow || '',
      net: u.params.type || 'tcp',
      type: u.params.headerType || 'none',
      host: u.params.host || '',
      path: u.params.path || u.params.serviceName || '',
      alpn: u.params.alpn ? decodeURIComponent(u.params.alpn) : '',
      sni: u.params.sni || '',
      tls: u.params.security || 'none',
      quicSecurity: u.params.quicSecurity || 'none',
      fp: u.params.fp || '',
      pbk: u.params.pbk || '',
      sid: u.params.sid || '',
      spx: u.params.spx || '',
      pinnedPeerCertSha256: u.params.pinnedPeerCertSha256 || '',
      verifyPeerCertByName: u.params.verifyPeerCertByName || '',
      key: u.params.key || '',
      xhttpMode: u.params.xhttpMode || 'auto',
      xhttpRawJson: u.params.xhttpRawJson || '',
      maxEarlyData: u.params.maxEarlyData || '',
      earlyDataHeaderName: u.params.earlyDataHeaderName || '',
      multiMode: u.params.multiMode === 'true' || u.params.multiMode === '1',
      idleTimeout: u.params.idleTimeout || '',
      healthCheckTimeout: u.params.healthCheckTimeout || '',
      permitWithoutStream: u.params.permitWithoutStream === 'true' || u.params.permitWithoutStream === '1',
      initialWindowsSize: u.params.initialWindowsSize || '',
      protocol: 'vless',
    }
    if (o.net === 'mkcp' || o.net === 'kcp') {
      o.path = u.params.seed || ''
    }
    return o
  }

  if (lower.startsWith('ss://')) {
    let u = parseURL(url)
    let userinfo = u.username
    let method = '', password = ''
    try {
      let decoded = base64ToUtf8(userinfo)
      let idx = decoded.indexOf(':')
      if (idx > -1) {
        method = decoded.substring(0, idx)
        password = decoded.substring(idx + 1)
      }
    } catch (e) {
      method = userinfo
    }
    const ssPlugin = u.params.plugin || ''
    return {
      method,
      password,
      server: u.host,
      port: u.port,
      name: decodeURIComponent(u.hash || ''),
      plugin: ssPlugin.split(';')[0] || '',
      plugin_opts: ssPlugin.split(';').slice(1).join(';') || '',
      protocol: 'ss',
      backend: u.params['v2raya-backend'] || '',
    }
  }

  if (lower.startsWith('ssr://')) {
    url = base64DecodeURI(url.substring(6))
    let arr = url.split('/?')
    let query = arr[1].split('&')
    let m: Record<string, string> = {}
    for (let param of query) {
      let [key, val] = param.split('=', 2)
      val = base64DecodeURI(val)
      m[key] = val
    }
    let pre = arr[0].split(':')
    if (pre.length > 6) {
      pre[pre.length - 6] = pre.slice(0, pre.length - 5).join(':')
      pre = pre.slice(pre.length - 6)
    }
    pre[5] = base64DecodeURI(pre[5])
    return {
      method: pre[3],
      password: pre[5],
      server: pre[0],
      port: pre[1],
      name: m['remarks'] || '',
      proto: pre[2],
      protoParam: m['protoparam'] || '',
      obfs: pre[4],
      obfsParam: m['obfsparam'] || '',
      protocol: 'ssr',
    }
  }

  if (lower.startsWith('trojan://') || lower.startsWith('trojan-go://')) {
    let u = parseURL(url)
    const o: any = {
      password: decodeURIComponent(u.username),
      server: u.host,
      port: u.port,
      name: decodeURIComponent(u.hash),
      peer: u.params.peer || u.params.sni || '',
      pinnedPeerCertSha256: u.params.pinnedPeerCertSha256 || '',
      verifyPeerCertByName: u.params.verifyPeerCertByName || '',
      method: 'origin',
      net: u.params.type || 'tcp',
      obfs: 'none',
      ssCipher: 'aes-128-gcm',
      ssPassword: '',
      path: u.params.path || u.params.serviceName || '',
      alpn: u.params.alpn || '',
      protocol: 'trojan',
      backend: u.params['v2raya-backend'] || '',
    }
    if (lower.startsWith('trojan-go://')) {
      if (u.params.encryption?.startsWith('ss;')) {
        const fields = u.params.encryption.split(';')
        o.method = 'shadowsocks'
        o.ssCipher = fields[1] || 'aes-128-gcm'
        o.ssPassword = fields[2] || ''
      }
      if (u.params.type === 'ws') {
        o.obfs = 'websocket'
        o.host = u.params.host || ''
        o.path = u.params.path || '/'
      }
    }
    return o
  }

  if (lower.startsWith('juicity://')) {
    let u = parseURL(url)
    return {
      name: decodeURIComponent(u.hash),
      uuid: decodeURIComponent(u.username),
      password: decodeURIComponent(u.password),
      server: u.host,
      port: u.port,
      sni: u.params.sni || '',
      pinnedCertchainSha256: u.params.pinned_certchain_sha256 || '',
      cc: u.params.congestion_control || 'bbr',
      fingerprint: u.params.fp || '',
      protocol: 'juicity',
    }
  }

  if (lower.startsWith('tuic://')) {
    let u = parseURL(url)
    return {
      name: decodeURIComponent(u.hash),
      uuid: decodeURIComponent(u.username),
      password: decodeURIComponent(u.password),
      server: u.host,
      port: u.port,
      sni: u.params.sni || '',
      pinnedPeerCertSha256: u.params.pinnedPeerCertSha256 || '',
      verifyPeerCertByName: u.params.verifyPeerCertByName || '',
      disableSni: u.params.disable_sni === 'true' || u.params.disable_sni === '1',
      alpn: u.params.alpn || 'h3',
      cc: u.params.congestion_control || 'bbr',
      fingerprint: u.params.fp || '',
      udpRelayMode: u.params.udp_relay_mode || 'native',
      protocol: 'tuic',
    }
  }

  if (lower.startsWith('hysteria2://') || lower.startsWith('hy2://')) {
    let u = parseURL(url)
    return {
      name: decodeURIComponent(u.hash),
      password: decodeURIComponent(u.username),
      server: u.host,
      port: u.port,
      sni: u.params.sni || '',
      pinnedPeerCertSha256: u.params.pinnedPeerCertSha256 || '',
      verifyPeerCertByName: u.params.verifyPeerCertByName || '',
      obfs: u.params.obfs || 'none',
      obfsPassword: u.params['obfs-password'] || '',
      up: u.params.up || '',
      down: u.params.down || '',
      protocol: 'hysteria2',
    }
  }

  if (lower.startsWith('http://') || lower.startsWith('https://')) {
    let u = parseURL(url)
    return {
      username: decodeURIComponent(u.username),
      password: decodeURIComponent(u.password),
      host: u.host,
      port: u.port,
      protocol: u.protocol,
      name: decodeURIComponent(u.hash),
    }
  }

  if (lower.startsWith('socks5://')) {
    let u = parseURL(url)
    return {
      username: decodeURIComponent(u.username),
      password: decodeURIComponent(u.password),
      host: u.host,
      port: u.port,
      protocol: u.protocol,
      name: decodeURIComponent(u.hash),
    }
  }

  if (lower.startsWith('anytls://')) {
    let u = parseURL(url)
    let auth = u.username ? decodeURIComponent(u.username) : ''
    let sni = u.params.peer || u.params.sni || ''
    return {
      name: decodeURIComponent(u.hash),
      host: u.host,
      port: u.port,
      auth,
      sni,
      alpn: u.params.alpn || '',
      fingerprint: u.params.fp || '',
      pinnedPeerCertSha256: u.params.pinnedPeerCertSha256 || '',
      verifyPeerCertByName: u.params.verifyPeerCertByName || '',
      protocol: 'anytls',
    }
  }

  if (lower.startsWith('wireguard://')) {
    let u = parseURL(url)
    return {
      name: decodeURIComponent(u.hash),
      address: u.host,
      port: u.port,
      publicKey: decodeURIComponent(u.username),
      privateKey: u.params.privateKey || '',
      localAddress: u.params.localAddress || '',
      dns: u.params.dns || '',
      mtu: u.params.mtu || '',
      allowedIPs: u.params.allowedIPs || '',
      persistentKeepalive: u.params.persistentKeepalive || '',
      preSharedKey: u.params.preSharedKey || '',
      endpoint: u.params.endpoint || '',
    }
  }

  return null
}

// ==================== URL Generator (form data -> sharing URL) ====================
function generateServerURL(srcObj: any): string | null {
  let query: Record<string, any>
  let tmp: any

  switch (srcObj.protocol) {
    case 'vless':
    case 'vmess': {
      if (srcObj.protocol === 'vless') {
        query = {
          type: srcObj.net,
          flow: srcObj.flow || '',
          security: srcObj.tls,
          fp: srcObj.fp || '',
          path: srcObj.path,
          host: srcObj.host,
          headerType: srcObj.type,
          sni: srcObj.sni,
          pinnedPeerCertSha256: srcObj.pinnedPeerCertSha256 || '',
          verifyPeerCertByName: srcObj.verifyPeerCertByName || '',
        }
        if (srcObj.alpn) query.alpn = srcObj.alpn
        if (srcObj.net === 'ws') {
          if (srcObj.maxEarlyData) query.maxEarlyData = srcObj.maxEarlyData
          if (srcObj.earlyDataHeaderName) query.earlyDataHeaderName = srcObj.earlyDataHeaderName
        }
        if (srcObj.net === 'grpc') {
          query.serviceName = srcObj.path
          if (srcObj.multiMode) query.multiMode = srcObj.multiMode
          if (srcObj.idleTimeout) query.idleTimeout = srcObj.idleTimeout
          if (srcObj.healthCheckTimeout) query.healthCheckTimeout = srcObj.healthCheckTimeout
          if (srcObj.permitWithoutStream) query.permitWithoutStream = srcObj.permitWithoutStream
          if (srcObj.initialWindowsSize) query.initialWindowsSize = srcObj.initialWindowsSize
        }
        if (srcObj.net === 'mkcp' || srcObj.net === 'kcp') query.seed = srcObj.path
        if (srcObj.net === 'quic') {
          query.key = srcObj.key
          query.quicSecurity = srcObj.quicSecurity
        }
        if (srcObj.net === 'xhttp') {
          query.xhttpMode = srcObj.xhttpMode
          if (srcObj.xhttpRawJson) query.xhttpRawJson = srcObj.xhttpRawJson
        }
        if (query.security === 'reality') {
          query.pbk = srcObj.pbk
          query.sid = srcObj.sid
          query.spx = srcObj.spx
        }
        return generateURL({
          protocol: 'vless',
          username: srcObj.id,
          host: srcObj.add,
          port: srcObj.port,
          hash: srcObj.ps,
          params: query,
        })
      }

      // vmess
      const obj = { ...srcObj }
      switch (obj.net) {
        case 'kcp':
        case 'tcp':
        case 'quic':
          break
        default:
          obj.type = ''
      }
      switch (obj.net) {
        case 'ws':
        case 'h2':
        case 'http':
        case 'quic':
        case 'grpc':
        case 'kcp':
        case 'mkcp':
          break
        default:
          if (!(obj.net === 'tcp' && obj.type === 'http')) obj.path = ''
      }
      return 'vmess://' + utf8ToBase64(JSON.stringify(obj))
    }

    case 'ss': {
      tmp = `ss://${utf8ToBase64(`${srcObj.method}:${srcObj.password}`)}@${srcObj.server}:${srcObj.port}/`
      if (srcObj.plugin) {
        if (srcObj.plugin_opts) {
          // Preserve plugin_opts as-is (from edit mode or URL import)
          tmp += `?plugin=${encodeURIComponent(srcObj.plugin + ';' + srcObj.plugin_opts)}`
        } else {
          // Reconstruct from individual form fields
          const plugin = [srcObj.plugin]
          if (srcObj.plugin === 'v2ray-plugin') {
            if (srcObj.tls) plugin.push('tls')
            if (srcObj.mode !== 'websocket') plugin.push('mode=' + srcObj.mode)
            if (srcObj.host) plugin.push('host=' + srcObj.host)
            if (srcObj.path) {
              if (!srcObj.path.startsWith('/')) srcObj.path = '/' + srcObj.path
              plugin.push('path=' + srcObj.path)
            }
            if (srcObj.impl) plugin.push('impl=' + srcObj.impl)
          } else {
            plugin.push('obfs=' + srcObj.obfs)
            plugin.push('obfs-host=' + srcObj.host)
            if (srcObj.obfs === 'http') plugin.push('obfs-path=' + srcObj.path)
            if (srcObj.impl) plugin.push('impl=' + srcObj.impl)
          }
          tmp += `?plugin=${encodeURIComponent(plugin.join(';'))}`
        }
      }
      if (srcObj.backend) {
        tmp += (tmp.includes('?') ? '&' : '?') + `v2raya-backend=${encodeURIComponent(srcObj.backend)}`
      }
      if (srcObj.name) tmp += `#${encodeURIComponent(srcObj.name)}`
      return tmp
    }

    case 'ssr': {
      return `ssr://${utf8ToBase64(
        `${srcObj.server}:${srcObj.port}:${srcObj.proto}:${srcObj.method}:${srcObj.obfs}:${base64EncodeURI(srcObj.password)}/?remarks=${base64EncodeURI(srcObj.name)}&protoparam=${base64EncodeURI(srcObj.protoParam || '')}&obfsparam=${base64EncodeURI(srcObj.obfsParam || '')}`
      )}`
    }

    case 'trojan': {
      query = {
        type: srcObj.net,
        pinnedPeerCertSha256: srcObj.pinnedPeerCertSha256 || '',
        verifyPeerCertByName: srcObj.verifyPeerCertByName || '',
      }
      if (srcObj.peer) query.sni = srcObj.peer
      if (srcObj.alpn) query.alpn = srcObj.alpn

      let proto = 'trojan'
      if (srcObj.method !== 'origin' || srcObj.obfs !== 'none') {
        proto = 'trojan-go'
        query.type = srcObj.obfs === 'none' ? 'original' : 'ws'
        if (srcObj.method === 'shadowsocks') {
          query.encryption = `ss;${srcObj.ssCipher};${srcObj.ssPassword}`
        }
        if (query.type === 'ws') {
          query.host = srcObj.host || ''
          query.path = srcObj.path || '/'
        }
      }
      if (srcObj.net === 'grpc') query.serviceName = srcObj.path
      if (srcObj.net === 'mkcp' || srcObj.net === 'kcp') query.seed = srcObj.path
      if (srcObj.backend) query['v2raya-backend'] = srcObj.backend

      return generateURL({
        protocol: proto,
        username: srcObj.password,
        host: srcObj.server,
        port: srcObj.port,
        hash: srcObj.name,
        params: query,
      })
    }

    case 'juicity': {
      query = {
        congestion_control: srcObj.cc,
      }
      if (srcObj.sni) query.sni = srcObj.sni
      if (srcObj.pinnedCertchainSha256) query.pinned_certchain_sha256 = srcObj.pinnedCertchainSha256
      if (srcObj.fingerprint) query.fp = srcObj.fingerprint
      return generateURL({
        protocol: 'juicity',
        username: srcObj.uuid,
        password: srcObj.password,
        host: srcObj.server,
        port: srcObj.port,
        hash: srcObj.name,
        params: query,
      })
    }

    case 'tuic': {
      query = {
        pinnedPeerCertSha256: srcObj.pinnedPeerCertSha256 || '',
        verifyPeerCertByName: srcObj.verifyPeerCertByName || '',
        congestion_control: srcObj.cc,
        disable_sni: srcObj.disableSni,
        alpn: srcObj.alpn,
        udp_relay_mode: srcObj.udpRelayMode,
      }
      if (srcObj.sni) query.sni = srcObj.sni
      if (srcObj.fingerprint) query.fp = srcObj.fingerprint
      return generateURL({
        protocol: 'tuic',
        username: srcObj.uuid,
        password: srcObj.password,
        host: srcObj.server,
        port: srcObj.port,
        hash: srcObj.name,
        params: query,
      })
    }

    case 'hysteria2': {
      query = {
        pinnedPeerCertSha256: srcObj.pinnedPeerCertSha256 || '',
        verifyPeerCertByName: srcObj.verifyPeerCertByName || '',
      }
      if (srcObj.sni) query.sni = srcObj.sni
      if (srcObj.obfs !== 'none') {
        query.obfs = srcObj.obfs
        query['obfs-password'] = srcObj.obfsPassword
      }
      if (srcObj.up) query.up = srcObj.up
      if (srcObj.down) query.down = srcObj.down
      return generateURL({
        protocol: 'hysteria2',
        username: srcObj.password,
        host: srcObj.server,
        port: srcObj.port,
        hash: srcObj.name,
        params: query,
      })
    }

    case 'http':
    case 'https': {
      tmp = {
        protocol: srcObj.protocol + '-proxy',
        host: srcObj.host,
        port: srcObj.port,
        hash: srcObj.name,
      }
      if (srcObj.username && srcObj.password) {
        Object.assign(tmp, { username: srcObj.username, password: srcObj.password })
      }
      return generateURL(tmp)
    }

    case 'socks5': {
      tmp = {
        protocol: 'socks5',
        host: srcObj.host,
        port: srcObj.port,
        hash: srcObj.name,
      }
      if (srcObj.username && srcObj.password) {
        Object.assign(tmp, { username: srcObj.username, password: srcObj.password })
      }
      return generateURL(tmp)
    }

    case 'anytls': {
      const q: Record<string, any> = {}
      if (srcObj.sni) q.peer = srcObj.sni
      if (srcObj.alpn) q.alpn = srcObj.alpn
      if (srcObj.pinnedPeerCertSha256) q.pinnedPeerCertSha256 = srcObj.pinnedPeerCertSha256
      if (srcObj.verifyPeerCertByName) q.verifyPeerCertByName = srcObj.verifyPeerCertByName
      if (srcObj.fingerprint) q.fp = srcObj.fingerprint
      return generateURL({
        protocol: 'anytls',
        username: srcObj.auth,
        host: srcObj.host,
        port: srcObj.port,
        hash: srcObj.name,
        params: q,
      })
    }

    case 'wireguard': {
      const q: Record<string, any> = {}
      if (srcObj.privateKey) q.privateKey = srcObj.privateKey
      if (srcObj.localAddress) q.localAddress = srcObj.localAddress
      if (srcObj.dns) q.dns = srcObj.dns
      if (srcObj.mtu) q.mtu = srcObj.mtu
      if (srcObj.allowedIPs) q.allowedIPs = srcObj.allowedIPs
      if (srcObj.persistentKeepalive) q.persistentKeepalive = srcObj.persistentKeepalive
      if (srcObj.preSharedKey) q.preSharedKey = srcObj.preSharedKey
      if (srcObj.endpoint) q.endpoint = srcObj.endpoint
      return generateURL({
        protocol: 'wireguard',
        username: srcObj.publicKey,
        host: srcObj.address,
        port: srcObj.port,
        hash: srcObj.name,
        params: q,
      })
    }
  }
  return null
}

// ==================== Submit Handler ====================
const isSubmitting = ref(false)

async function handleSubmit() {
  // Validate visible form
  let valid = true
  const formRefs: (FormInstance | undefined)[] = [
    v2rayFormRef.value,
    wireguardFormRef.value,
    ssFormRef.value,
    ssrFormRef.value,
    trojanFormRef.value,
    juicityFormRef.value,
    tuicFormRef.value,
    hysteria2FormRef.value,
    httpFormRef.value,
    socks5FormRef.value,
    anytlsFormRef.value,
  ]

  const formRef = formRefs[tabChoice.value]
  if (formRef) {
    try {
      await formRef.validate()
    } catch {
      valid = false
    }
  }

  if (!valid) return

  // Check mandatory fields for v2ray tab
  if (tabChoice.value === 0) {
    if (!v2ray.add || !v2ray.port || !v2ray.id) {
      ElMessage.warning(t('configureServer.servername'))
      return
    }
  }

  let coded = ''
  let srcObj: any = null

  switch (tabChoice.value) {
    case 0:
      coded = generateServerURL(v2ray) || ''
      break
    case 1:
      coded = generateServerURL(wireguard) || ''
      break
    case 2:
      coded = generateServerURL(ss) || ''
      break
    case 3:
      coded = generateServerURL(ssr) || ''
      break
    case 4:
      coded = generateServerURL(trojan) || ''
      break
    case 5:
      coded = generateServerURL(juicity) || ''
      break
    case 6:
      coded = generateServerURL(tuic) || ''
      break
    case 7:
      coded = generateServerURL(hysteria2) || ''
      break
    case 8:
      coded = generateServerURL(http) || ''
      break
    case 9:
      coded = generateServerURL(socks5) || ''
      break
    case 10:
      coded = generateServerURL(anytls) || ''
      break
  }

  if (!coded) {
    ElMessage.error('生成节点链接失败')
    return
  }

  isSubmitting.value = true
  try {
    const body: any = { url: coded }
    if (props.server) {
      body.which = {
        id: props.server.id,
        _type: props.server._type,
        sub: props.server.sub !== undefined ? props.server.sub : undefined,
      }
    }

    const { data } = await useV2Fetch('import').post(body).json()
    if (data.value?.code === 'SUCCESS') {
      ElMessage.success(t('common.success'))
      emit('save')
      emit('update:modelValue', false)
    }
  } finally {
    isSubmitting.value = false
  }
}

// ==================== Network Change Handler ====================
function handleNetworkChange() {
  v2ray.type = 'none'
  if (v2ray.tls === 'none' && v2ray.net === 'grpc') {
    ElMessage.warning(t('setting.messages.grpcShouldWithTls'))
    setTimeout(() => {
      v2ray.tls = 'tls'
    }, 100)
  }
}

function handleV2rayProtocolSwitch() {
  // Reset type when switching protocol
}

// ==================== Edit Mode: fetch and populate ====================
const isEditMode = computed(() => !!props.server)
const dialogVisible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
})

async function loadServerData() {
  if (!props.server) return

  const params = JSON.stringify({
    id: props.server.id,
    _type: props.server._type,
    sub: props.server.sub !== undefined ? props.server.sub : undefined,
  })

  const { data } = await useV2Fetch(`sharingAddress?touch=${encodeURIComponent(params)}`).get().json()
  if (data.value?.code !== 'SUCCESS') return

  const sharingAddr = data.value.data.sharingAddress
  if (!sharingAddr) return

  const lower = sharingAddr.toLowerCase()

  if (lower.startsWith('vmess://') || lower.startsWith('vless://')) {
    const resolved = resolveURL(sharingAddr)
    Object.assign(v2ray, resolved)
    tabChoice.value = 0
  } else if (lower.startsWith('wireguard://')) {
    const resolved = resolveURL(sharingAddr)
    Object.assign(wireguard, resolved)
    tabChoice.value = 1
  } else if (lower.startsWith('ss://')) {
    const resolved = resolveURL(sharingAddr)
    Object.assign(ss, resolved)
    tabChoice.value = 2
  } else if (lower.startsWith('ssr://')) {
    const resolved = resolveURL(sharingAddr)
    Object.assign(ssr, resolved)
    tabChoice.value = 3
  } else if (lower.startsWith('trojan://') || lower.startsWith('trojan-go://')) {
    const resolved = resolveURL(sharingAddr)
    Object.assign(trojan, resolved)
    tabChoice.value = 4
  } else if (lower.startsWith('juicity://')) {
    const resolved = resolveURL(sharingAddr)
    Object.assign(juicity, resolved)
    tabChoice.value = 5
  } else if (lower.startsWith('tuic://')) {
    const resolved = resolveURL(sharingAddr)
    Object.assign(tuic, resolved)
    tabChoice.value = 6
  } else if (lower.startsWith('hysteria2://') || lower.startsWith('hy2://')) {
    const resolved = resolveURL(sharingAddr)
    Object.assign(hysteria2, resolved)
    tabChoice.value = 7
  } else if (lower.startsWith('http://') || lower.startsWith('https://')) {
    const resolved = resolveURL(sharingAddr)
    Object.assign(http, resolved)
    tabChoice.value = 8
  } else if (lower.startsWith('socks5://')) {
    const resolved = resolveURL(sharingAddr)
    Object.assign(socks5, resolved)
    tabChoice.value = 9
  } else if (lower.startsWith('anytls://')) {
    const resolved = resolveURL(sharingAddr)
    Object.assign(anytls, resolved)
    tabChoice.value = 10
  }
}

// Watch dialog opening
watch(() => props.modelValue, (val) => {
  if (val && isEditMode.value) {
    loadServerData()
  }
})

// ==================== Reset forms for create mode ====================
function resetForms() {
  // Reset all form data to defaults...
  // This is handled by the parent creating a new instance each time
}

// ==================== Computed for visibility ====================
const showV2rayTLS = computed(() => v2ray.type !== 'dtls')
const showV2raySNI = computed(() => v2ray.tls !== 'none')
const showV2rayFP = computed(() => v2ray.tls === 'tls' || v2ray.tls === 'reality')
const showV2rayALPN = computed(() => v2ray.tls === 'tls')
const showV2rayFlow = computed(() => v2ray.protocol === 'vless' && v2ray.tls !== 'none')
const showV2rayReality = computed(() => v2ray.tls === 'reality')
const showV2rayTypeTCP = computed(() => v2ray.net === 'tcp')
const showV2rayQUICSecurity = computed(() => v2ray.protocol === 'vless' && v2ray.net === 'quic')
const showV2rayTypeKCPQUIC = computed(() => v2ray.net === 'kcp' || v2ray.net === 'quic')
const showV2rayHost = computed(() =>
  v2ray.net === 'ws' || v2ray.net === 'h2' || v2ray.net === 'xhttp' ||
  v2ray.tls === 'tls' || (v2ray.net === 'tcp' && v2ray.type === 'http')
)
const showV2rayPath = computed(() =>
  v2ray.net === 'ws' || v2ray.net === 'h2' || (v2ray.net === 'tcp' && v2ray.type === 'http')
)
const showV2rayWS = computed(() => v2ray.net === 'ws')
const showV2rayKCP = computed(() => v2ray.net === 'mkcp' || v2ray.net === 'kcp')
const showV2rayGRPC = computed(() => v2ray.net === 'grpc')
const showV2rayXHTTP = computed(() => v2ray.net === 'xhttp')
const showV2rayQUIC = computed(() => v2ray.net === 'quic')

// SS computed
const showSSPlugin = computed(() => ss.plugin === 'simple-obfs' || ss.plugin === 'v2ray-plugin')
const showSSSimpleObfs = computed(() => ss.plugin === 'simple-obfs')
const showSSV2rayPlugin = computed(() => ss.plugin === 'v2ray-plugin')
const showSSHost = computed(() =>
  (ss.plugin === 'simple-obfs' && (ss.obfs === 'http' || ss.obfs === 'tls')) || ss.plugin === 'v2ray-plugin'
)
const showSSPath = computed(() =>
  (ss.plugin === 'simple-obfs' && ss.obfs === 'http') || ss.plugin === 'v2ray-plugin'
)

// Trojan computed
const showTrojanSS = computed(() => trojan.method === 'shadowsocks')
const showTrojanObfsWS = computed(() => trojan.obfs === 'websocket')
const showTrojanWSH2 = computed(() => trojan.net === 'ws' || trojan.net === 'h2')
const showTrojanALPN = computed(() => true)
const showTrojanKCP = computed(() => trojan.net === 'mkcp' || trojan.net === 'kcp')
const showTrojanGRPC = computed(() => trojan.net === 'grpc')
</script>

<template>
  <ElDialog
    v-model="dialogVisible"
    :title="readonly ? '查看节点' : '配置节点'"
    width="560px"
    :close-on-click-modal="false"
    destroy-on-close
  >
    <ElTabs v-model="tabChoice" type="border-card" class="server-editor-tabs">
      <!-- ==================== V2Ray Tab ==================== -->
      <ElTabPane label="V2RAY">
        <ElForm ref="v2rayFormRef" :model="v2ray" label-width="120px" label-position="top" size="default">
          <ElFormItem label="Protocol">
            <ElSelect v-model="v2ray.protocol" @change="handleV2rayProtocolSwitch">
              <ElOption value="vmess" label="VMESS" />
              <ElOption value="vless" label="VLESS" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem label="名称">
            <ElInput v-model="v2ray.ps" placeholder="节点名称" />
          </ElFormItem>
          <ElFormItem label="地址" required>
            <ElInput v-model="v2ray.add" placeholder="IP / HOST" />
          </ElFormItem>
          <ElFormItem label="端口" required>
            <ElInput v-model="v2ray.port" placeholder="端口号" type="number" />
          </ElFormItem>
          <ElFormItem label="ID" required>
            <ElInput v-model="v2ray.id" placeholder="UserID" />
          </ElFormItem>
          <ElFormItem v-if="v2ray.protocol === 'vmess'" label="AlterID">
            <ElInput v-model="v2ray.aid" placeholder="AlterID" type="number" min="0" max="65535" />
          </ElFormItem>
          <ElFormItem v-if="v2ray.protocol === 'vmess'" label="Security">
            <ElSelect v-model="v2ray.scy">
              <ElOption value="auto" label="Auto" />
              <ElOption value="aes-256-gcm" label="aes-256-gcm" />
              <ElOption value="aes-128-gcm" label="aes-128-gcm" />
              <ElOption value="chacha20-poly1305" label="chacha20-poly1305" />
              <ElOption value="xchacha20-poly1305" label="xchacha20-poly1305" />
              <ElOption value="none" label="none" />
              <ElOption value="zero" label="zero" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem v-show="showV2rayTLS" label="TLS">
            <ElSelect v-model="v2ray.tls" @change="handleNetworkChange">
              <ElOption value="none" label="关闭" />
              <ElOption value="tls" label="tls" />
              <ElOption v-if="variant() === 'xray'" value="reality" label="reality" />
              <ElOption v-if="variant() === 'xray'" value="xtls" label="xtls" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem v-if="showV2raySNI" label="SNI">
            <ElInput v-model="v2ray.sni" placeholder="SNI" />
          </ElFormItem>
          <ElFormItem v-show="showV2rayFP" label="uTLS fingerprint">
            <ElSelect v-model="v2ray.fp">
              <ElOption value="" label="empty" />
              <ElOption value="chrome" label="chrome" />
              <ElOption value="firefox" label="firefox" />
              <ElOption value="safari" label="safari" />
              <ElOption value="ios" label="ios" />
              <ElOption value="android" label="android" />
              <ElOption value="edge" label="edge" />
              <ElOption value="random" label="random" />
              <ElOption value="randomized" label="randomized" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem v-show="showV2rayALPN" label="Alpn">
            <ElInput v-model="v2ray.alpn" placeholder="h3,h2,http/1.1" />
          </ElFormItem>
          <ElFormItem v-if="showV2rayFlow" label="Flow">
            <ElInput v-model="v2ray.flow" placeholder="Flow" />
          </ElFormItem>
          <!-- Reality fields -->
          <ElFormItem v-show="showV2rayReality" label="pbk">
            <ElInput v-model="v2ray.pbk" placeholder="Public Key" />
          </ElFormItem>
          <ElFormItem v-show="showV2rayReality" label="sid">
            <ElInput v-model="v2ray.sid" placeholder="Short ID" />
          </ElFormItem>
          <ElFormItem v-show="showV2rayReality" label="spx">
            <ElInput v-model="v2ray.spx" placeholder="SpiderX" />
          </ElFormItem>
          <ElFormItem label="固定证书 SHA256">
            <ElInput v-model="v2ray.pinnedPeerCertSha256" :placeholder="$t('pinnedPeerCertSha256Placeholder')" class="mb-2" />
          </ElFormItem>
          <ElFormItem label="证书验证域名">
            <ElInput v-model="v2ray.verifyPeerCertByName" :placeholder="$t('verifyPeerCertByNamePlaceholder')" />
          </ElFormItem>
          <ElFormItem label="Network">
            <ElSelect v-model="v2ray.net" @change="handleNetworkChange">
              <ElOption value="tcp" label="TCP" />
              <ElOption value="kcp" label="mKCP" />
              <ElOption value="ws" label="WebSocket" />
              <ElOption value="h2" label="HTTP/2" />
              <ElOption value="grpc" label="gRPC" />
              <ElOption value="quic" label="QUIC" />
              <ElOption value="xhttp" label="XHTTP" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem v-show="showV2rayTypeTCP" label="Type">
            <ElSelect v-model="v2ray.type">
              <ElOption value="none" label="不伪装" />
              <ElOption value="http" label="伪装为HTTP" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem v-show="showV2rayQUICSecurity" label="QUIC Security">
            <ElSelect v-model="v2ray.quicSecurity">
              <ElOption value="none" label="none" />
              <ElOption value="aes-128-gcm" label="aes-128-gcm" />
              <ElOption value="chacha20-poly1305" label="chacha20-poly1305" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem v-show="showV2rayTypeKCPQUIC" label="Type">
            <ElSelect v-model="v2ray.type">
              <ElOption value="none" label="不伪装" />
              <ElOption value="srtp" label="伪装视频通话(SRTP)" />
              <ElOption value="utp" label="伪装为BT下载(uTP)" />
              <ElOption value="wechat-video" label="伪装为微信视频通话" />
              <ElOption value="dtls" label="伪装为DTLS1.2(强制TLS)" />
              <ElOption value="wireguard" label="伪装为WireGuard" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem v-show="showV2rayHost" label="Host">
            <ElInput v-model="v2ray.host" placeholder="域名(host)" />
          </ElFormItem>
          <ElFormItem v-show="showV2rayPath" label="Path">
            <ElInput v-model="v2ray.path" placeholder="路径(path)" />
          </ElFormItem>
          <ElFormItem v-show="showV2rayWS" label="Max Early Data">
            <ElInput v-model="v2ray.maxEarlyData" type="number" placeholder="Max Early Data" />
          </ElFormItem>
          <ElFormItem v-show="showV2rayWS" label="Early Data Header Name">
            <ElInput v-model="v2ray.earlyDataHeaderName" placeholder="Early Data Header Name" />
          </ElFormItem>
          <ElFormItem v-show="showV2rayKCP" label="Seed">
            <ElInput v-model="v2ray.path" placeholder="混淆种子" />
          </ElFormItem>
          <ElFormItem v-show="showV2rayGRPC" label="Service Name">
            <ElInput v-model="v2ray.path" placeholder="Service Name" />
          </ElFormItem>
          <ElFormItem v-show="showV2rayGRPC" label="MultiMode">
            <ElSwitch v-model="v2ray.multiMode" />
          </ElFormItem>
          <ElFormItem v-show="showV2rayGRPC" label="Idle Timeout">
            <ElInput v-model="v2ray.idleTimeout" type="number" placeholder="Idle Timeout (s)" />
          </ElFormItem>
          <ElFormItem v-show="showV2rayGRPC" label="Health Check Timeout">
            <ElInput v-model="v2ray.healthCheckTimeout" type="number" placeholder="Health Check Timeout (s)" />
          </ElFormItem>
          <ElFormItem v-show="showV2rayGRPC" label="Permit Without Stream">
            <ElSwitch v-model="v2ray.permitWithoutStream" />
          </ElFormItem>
          <ElFormItem v-show="showV2rayGRPC" label="Initial Windows Size">
            <ElInput v-model="v2ray.initialWindowsSize" type="number" placeholder="Initial Windows Size" />
          </ElFormItem>
          <ElFormItem v-show="showV2rayXHTTP" label="Path">
            <ElInput v-model="v2ray.path" placeholder="路径(path)" />
          </ElFormItem>
          <ElFormItem v-show="showV2rayXHTTP" label="Mode">
            <ElSelect v-model="v2ray.xhttpMode">
              <ElOption value="auto" label="auto" />
              <ElOption value="packet-up" label="packet-up" />
              <ElOption value="stream-up" label="stream-up" />
              <ElOption value="stream-one" label="stream-one" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem v-show="showV2rayXHTTP" label="Extra Raw JSON">
            <ElInput v-model="v2ray.xhttpRawJson" placeholder="{XHTTPObject}" />
          </ElFormItem>
          <ElFormItem v-show="showV2rayQUIC" label="Key">
            <ElInput v-model="v2ray.key" placeholder="密码" />
          </ElFormItem>
        </ElForm>
      </ElTabPane>

      <!-- ==================== WireGuard Tab ==================== -->
      <ElTabPane label="WireGuard">
        <ElForm ref="wireguardFormRef" :model="wireguard" label-width="120px" label-position="top">
          <ElFormItem label="名称">
            <ElInput v-model="wireguard.name" placeholder="节点名称" />
          </ElFormItem>
          <ElFormItem label="地址" required>
            <ElInput v-model="wireguard.address" placeholder="IP / HOST" />
          </ElFormItem>
          <ElFormItem label="端口" required>
            <ElInput v-model="wireguard.port" placeholder="端口号" type="number" />
          </ElFormItem>
          <ElFormItem label="Public Key" required>
            <ElInput v-model="wireguard.publicKey" placeholder="Public Key" />
          </ElFormItem>
          <ElFormItem label="Private Key" required>
            <ElInput v-model="wireguard.privateKey" placeholder="Private Key" />
          </ElFormItem>
          <ElFormItem label="地址 (本地)">
            <ElInput v-model="wireguard.localAddress" placeholder="CIDR, e.g. 10.0.0.1/24" />
          </ElFormItem>
          <ElFormItem label="DNS">
            <ElInput v-model="wireguard.dns" placeholder="DNS Server" />
          </ElFormItem>
          <ElFormItem label="MTU">
            <ElInput v-model="wireguard.mtu" type="number" placeholder="MTU" />
          </ElFormItem>
          <ElFormItem label="Allowed IPs">
            <ElInput v-model="wireguard.allowedIPs" placeholder="0.0.0.0/0, ::/0" />
          </ElFormItem>
          <ElFormItem label="Persistent Keepalive">
            <ElInput v-model="wireguard.persistentKeepalive" type="number" placeholder="Persistent Keepalive (s)" />
          </ElFormItem>
          <ElFormItem label="Pre-shared Key">
            <ElInput v-model="wireguard.preSharedKey" placeholder="Pre-shared Key" />
          </ElFormItem>
          <ElFormItem label="Endpoint">
            <ElInput v-model="wireguard.endpoint" placeholder="Endpoint (可选，默认同Address:Port)" />
          </ElFormItem>
        </ElForm>
      </ElTabPane>

      <!-- ==================== SS Tab ==================== -->
      <ElTabPane label="SS">
        <ElForm ref="ssFormRef" :model="ss" label-width="120px" label-position="top">
          <ElFormItem label="名称">
            <ElInput v-model="ss.name" placeholder="节点名称" />
          </ElFormItem>
          <ElFormItem label="地址" required>
            <ElInput v-model="ss.server" placeholder="IP / HOST" />
          </ElFormItem>
          <ElFormItem label="端口" required>
            <ElInput v-model="ss.port" placeholder="端口号" type="number" />
          </ElFormItem>
          <ElFormItem label="密码" required>
            <ElInput v-model="ss.password" placeholder="密码" />
          </ElFormItem>
          <ElFormItem label="加密方式">
            <ElSelect v-model="ss.method">
              <ElOption value="2022-blake3-aes-128-gcm" label="2022-blake3-aes-128-gcm" />
              <ElOption value="2022-blake3-aes-256-gcm" label="2022-blake3-aes-256-gcm" />
              <ElOption value="2022-blake3-chacha20-poly1305" label="2022-blake3-chacha20-poly1305" />
              <ElOption value="aes-128-gcm" label="aes-128-gcm" />
              <ElOption value="aes-256-gcm" label="aes-256-gcm" />
              <ElOption value="chacha20-poly1305" label="chacha20-poly1305" />
              <ElOption value="chacha20-ietf-poly1305" label="chacha20-ietf-poly1305" />
              <ElOption value="plain" label="plain" />
              <ElOption value="none" label="none" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem label="插件">
            <ElSelect v-model="ss.plugin">
              <ElOption value="" label="关闭" />
              <ElOption value="simple-obfs" label="simple-obfs" />
              <ElOption value="v2ray-plugin" label="v2ray-plugin" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem v-if="showSSPlugin" label="实现方式">
            <ElSelect v-model="ss.impl">
              <ElOption value="" label="默认" />
              <ElOption value="chained" label="chained" />
              <ElOption value="transport" label="transport" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem v-show="showSSSimpleObfs" label="Obfs">
            <ElSelect v-model="ss.obfs">
              <ElOption value="http" label="http" />
              <ElOption value="tls" label="tls" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem v-show="showSSV2rayPlugin" label="Mode">
            <ElSelect v-model="ss.mode">
              <ElOption value="websocket" label="websocket" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem v-show="showSSV2rayPlugin" label="TLS">
            <ElSelect v-model="ss.tls">
              <ElOption value="" label="关闭" />
              <ElOption value="tls" label="tls" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem v-if="showSSHost" label="Host">
            <ElInput v-model="ss.host" placeholder="(可选)" />
          </ElFormItem>
          <ElFormItem v-if="showSSPath" label="Path">
            <ElInput v-model="ss.path" placeholder="/" />
          </ElFormItem>
          <ElFormItem label="后端">
            <ElSelect v-model="ss.backend">
              <ElOption value="" label="系统默认" />
              <ElOption value="v2ray" label="v2ray" />
            </ElSelect>
          </ElFormItem>
        </ElForm>
      </ElTabPane>

      <!-- ==================== SSR Tab ==================== -->
      <ElTabPane label="SSR">
        <ElForm ref="ssrFormRef" :model="ssr" label-width="120px" label-position="top">
          <ElFormItem label="名称">
            <ElInput v-model="ssr.name" placeholder="节点名称" />
          </ElFormItem>
          <ElFormItem label="地址" required>
            <ElInput v-model="ssr.server" placeholder="IP / HOST" />
          </ElFormItem>
          <ElFormItem label="端口" required>
            <ElInput v-model="ssr.port" placeholder="端口号" type="number" />
          </ElFormItem>
          <ElFormItem label="密码" required>
            <ElInput v-model="ssr.password" placeholder="密码" />
          </ElFormItem>
          <ElFormItem label="加密方式">
            <ElSelect v-model="ssr.method">
              <ElOption v-for="m in ['aes-128-cfb','aes-192-cfb','aes-256-cfb','aes-128-ctr','aes-192-ctr','aes-256-ctr','aes-128-ofb','aes-192-ofb','aes-256-ofb','des-cfb','bf-cfb','cast5-cfb','rc4-md5','chacha20','chacha20-ietf','salsa20','camellia-128-cfb','camellia-192-cfb','camellia-256-cfb','idea-cfb','rc2-cfb','seed-cfb','none']" :key="m" :value="m" :label="m" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem label="协议">
            <ElSelect v-model="ssr.proto">
              <ElOption value="origin" label="origin" />
              <ElOption value="verify_sha1" label="verify_sha1" />
              <ElOption value="auth_sha1_v4" label="auth_sha1_v4" />
              <ElOption value="auth_aes128_md5" label="auth_aes128_md5" />
              <ElOption value="auth_aes128_sha1" label="auth_aes128_sha1" />
              <ElOption value="auth_chain_a" label="auth_chain_a" />
              <ElOption value="auth_chain_b" label="auth_chain_b" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem v-if="ssr.proto !== 'origin'" label="协议参数">
            <ElInput v-model="ssr.protoParam" placeholder="(可选)" />
          </ElFormItem>
          <ElFormItem label="混淆">
            <ElSelect v-model="ssr.obfs">
              <ElOption value="plain" label="plain" />
              <ElOption value="http_simple" label="http_simple" />
              <ElOption value="http_post" label="http_post" />
              <ElOption value="random_head" label="random_head" />
              <ElOption value="tls1.2_ticket_auth" label="tls1.2_ticket_auth" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem v-if="ssr.obfs !== 'plain'" label="混淆参数">
            <ElInput v-model="ssr.obfsParam" placeholder="(可选)" />
          </ElFormItem>
        </ElForm>
      </ElTabPane>

      <!-- ==================== Trojan Tab ==================== -->
      <ElTabPane label="Trojan">
        <ElForm ref="trojanFormRef" :model="trojan" label-width="120px" label-position="top">
          <ElFormItem label="名称">
            <ElInput v-model="trojan.name" placeholder="节点名称" />
          </ElFormItem>
          <ElFormItem label="地址" required>
            <ElInput v-model="trojan.server" placeholder="IP / HOST" />
          </ElFormItem>
          <ElFormItem label="端口" required>
            <ElInput v-model="trojan.port" placeholder="端口号" type="number" />
          </ElFormItem>
          <ElFormItem label="密码" required>
            <ElInput v-model="trojan.password" placeholder="密码" />
          </ElFormItem>
          <ElFormItem label="协议">
            <ElSelect v-model="trojan.method">
              <ElOption value="origin" label="原版" />
              <ElOption value="shadowsocks" label="shadowsocks" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem v-if="showTrojanSS" label="Shadowsocks 加密">
            <ElSelect v-model="trojan.ssCipher">
              <ElOption value="aes-128-gcm" label="aes-128-gcm" />
              <ElOption value="aes-256-gcm" label="aes-256-gcm" />
              <ElOption value="chacha20-poly1305" label="chacha20-poly1305" />
              <ElOption value="chacha20-ietf-poly1305" label="chacha20-ietf-poly1305" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem v-if="showTrojanSS" label="Shadowsocks 密码">
            <ElInput v-model="trojan.ssPassword" placeholder="shadowsocks密码" />
          </ElFormItem>
          <ElFormItem label="固定证书 SHA256">
            <ElInput v-model="trojan.pinnedPeerCertSha256" :placeholder="$t('pinnedPeerCertSha256Placeholder')" class="mb-2" />
          </ElFormItem>
          <ElFormItem label="证书验证域名">
            <ElInput v-model="trojan.verifyPeerCertByName" :placeholder="$t('verifyPeerCertByNamePlaceholder')" />
          </ElFormItem>
          <ElFormItem label="SNI(Peer)">
            <ElInput v-model="trojan.peer" placeholder="SNI(Peer)" />
          </ElFormItem>
          <ElFormItem label="Network">
            <ElSelect v-model="trojan.net">
              <ElOption value="tcp" label="TCP" />
              <ElOption value="kcp" label="mKCP" />
              <ElOption value="ws" label="WebSocket" />
              <ElOption value="h2" label="HTTP/2" />
              <ElOption value="grpc" label="gRPC" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem label="Obfs">
            <ElSelect v-model="trojan.obfs">
              <ElOption value="none" label="不伪装" />
              <ElOption value="websocket" label="websocket" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem v-show="showTrojanObfsWS" label="Websocket Host">
            <ElInput v-model="trojan.host" />
          </ElFormItem>
          <ElFormItem v-show="showTrojanObfsWS" label="Websocket Path">
            <ElInput v-model="trojan.path" placeholder="/" />
          </ElFormItem>
          <ElFormItem v-show="showTrojanWSH2" label="Host">
            <ElInput v-model="trojan.host" placeholder="域名(host)" />
          </ElFormItem>
          <ElFormItem v-show="showTrojanWSH2" label="Path">
            <ElInput v-model="trojan.path" placeholder="路径(path)" />
          </ElFormItem>
          <ElFormItem v-show="showTrojanALPN" label="Alpn">
            <ElInput v-model="trojan.alpn" placeholder="h2,http/1.1" />
          </ElFormItem>
          <ElFormItem v-show="showTrojanKCP" label="Seed">
            <ElInput v-model="trojan.path" placeholder="混淆种子" />
          </ElFormItem>
          <ElFormItem v-show="showTrojanGRPC" label="Service Name">
            <ElInput v-model="trojan.path" placeholder="Service Name" />
          </ElFormItem>
          <ElFormItem label="后端">
            <ElSelect v-model="trojan.backend">
              <ElOption value="" label="系统默认" />
              <ElOption value="v2ray" label="v2ray" />
            </ElSelect>
          </ElFormItem>
        </ElForm>
      </ElTabPane>

      <!-- ==================== Juicity Tab ==================== -->
      <ElTabPane label="Juicity">
        <ElForm ref="juicityFormRef" :model="juicity" label-width="120px" label-position="top">
          <ElFormItem label="名称">
            <ElInput v-model="juicity.name" placeholder="节点名称" />
          </ElFormItem>
          <ElFormItem label="地址" required>
            <ElInput v-model="juicity.server" placeholder="IP / HOST" />
          </ElFormItem>
          <ElFormItem label="端口" required>
            <ElInput v-model="juicity.port" placeholder="端口号" type="number" />
          </ElFormItem>
          <ElFormItem label="UUID" required>
            <ElInput v-model="juicity.uuid" placeholder="UUID" />
          </ElFormItem>
          <ElFormItem label="密码" required>
            <ElInput v-model="juicity.password" placeholder="密码" />
          </ElFormItem>
          <ElFormItem label="拥塞控制">
            <ElSelect v-model="juicity.cc">
              <ElOption value="bbr" label="bbr" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem label="SNI">
            <ElInput v-model="juicity.sni" placeholder="SNI" />
          </ElFormItem>
          <ElFormItem label="Pinned Cert Chain Sha256">
            <ElInput v-model="juicity.pinnedCertchainSha256" placeholder="Pinned Cert Chain Sha256" />
          </ElFormItem>
          <ElFormItem label="uTLS fingerprint">
            <ElSelect v-model="juicity.fingerprint">
              <ElOption value="" label="empty" />
              <ElOption value="chrome" label="chrome" />
              <ElOption value="firefox" label="firefox" />
              <ElOption value="safari" label="safari" />
              <ElOption value="ios" label="ios" />
              <ElOption value="android" label="android" />
              <ElOption value="edge" label="edge" />
              <ElOption value="random" label="random" />
              <ElOption value="randomized" label="randomized" />
            </ElSelect>
          </ElFormItem>
        </ElForm>
      </ElTabPane>

      <!-- ==================== Tuic Tab ==================== -->
      <ElTabPane label="Tuic">
        <ElForm ref="tuicFormRef" :model="tuic" label-width="120px" label-position="top">
          <ElFormItem label="名称">
            <ElInput v-model="tuic.name" placeholder="节点名称" />
          </ElFormItem>
          <ElFormItem label="地址" required>
            <ElInput v-model="tuic.server" placeholder="IP / HOST" />
          </ElFormItem>
          <ElFormItem label="端口" required>
            <ElInput v-model="tuic.port" placeholder="端口号" type="number" />
          </ElFormItem>
          <ElFormItem label="UUID" required>
            <ElInput v-model="tuic.uuid" placeholder="UUID" />
          </ElFormItem>
          <ElFormItem label="密码" required>
            <ElInput v-model="tuic.password" placeholder="密码" />
          </ElFormItem>
          <ElFormItem label="拥塞控制">
            <ElSelect v-model="tuic.cc">
              <ElOption value="bbr" label="bbr" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem label="固定证书 SHA256">
            <ElInput v-model="tuic.pinnedPeerCertSha256" :placeholder="$t('pinnedPeerCertSha256Placeholder')" class="mb-2" />
          </ElFormItem>
          <ElFormItem label="证书验证域名">
            <ElInput v-model="tuic.verifyPeerCertByName" :placeholder="$t('verifyPeerCertByNamePlaceholder')" />
          </ElFormItem>
          <ElFormItem label="DisableSni">
            <ElSelect v-model="tuic.disableSni">
              <ElOption :value="false" label="否" />
              <ElOption :value="true" label="是" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem v-if="!tuic.disableSni" label="SNI">
            <ElInput v-model="tuic.sni" placeholder="SNI" />
          </ElFormItem>
          <ElFormItem label="ALPN">
            <ElInput v-model="tuic.alpn" placeholder="h3" />
          </ElFormItem>
          <ElFormItem label="uTLS fingerprint">
            <ElSelect v-model="tuic.fingerprint">
              <ElOption value="" label="empty" />
              <ElOption value="chrome" label="chrome" />
              <ElOption value="firefox" label="firefox" />
              <ElOption value="safari" label="safari" />
              <ElOption value="ios" label="ios" />
              <ElOption value="android" label="android" />
              <ElOption value="edge" label="edge" />
              <ElOption value="random" label="random" />
              <ElOption value="randomized" label="randomized" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem label="UDP relay mode">
            <ElSelect v-model="tuic.udpRelayMode">
              <ElOption value="native" label="native" />
              <ElOption value="quic" label="quic" />
            </ElSelect>
          </ElFormItem>
        </ElForm>
      </ElTabPane>

      <!-- ==================== Hysteria2 Tab ==================== -->
      <ElTabPane label="Hysteria2">
        <ElForm ref="hysteria2FormRef" :model="hysteria2" label-width="120px" label-position="top">
          <ElFormItem label="名称">
            <ElInput v-model="hysteria2.name" placeholder="节点名称" />
          </ElFormItem>
          <ElFormItem label="地址" required>
            <ElInput v-model="hysteria2.server" placeholder="IP / HOST" />
          </ElFormItem>
          <ElFormItem label="端口" required>
            <ElInput v-model="hysteria2.port" placeholder="端口号" type="number" />
          </ElFormItem>
          <ElFormItem label="密码" required>
            <ElInput v-model="hysteria2.password" placeholder="密码" />
          </ElFormItem>
          <ElFormItem label="固定证书 SHA256">
            <ElInput v-model="hysteria2.pinnedPeerCertSha256" :placeholder="$t('pinnedPeerCertSha256Placeholder')" class="mb-2" />
          </ElFormItem>
          <ElFormItem label="证书验证域名">
            <ElInput v-model="hysteria2.verifyPeerCertByName" :placeholder="$t('verifyPeerCertByNamePlaceholder')" />
          </ElFormItem>
          <ElFormItem label="SNI">
            <ElInput v-model="hysteria2.sni" placeholder="SNI" />
          </ElFormItem>
          <ElFormItem label="Obfs">
            <ElSelect v-model="hysteria2.obfs">
              <ElOption value="none" label="none" />
              <ElOption value="salamander" label="salamander" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem v-if="hysteria2.obfs !== 'none'" label="Obfs Password">
            <ElInput v-model="hysteria2.obfsPassword" placeholder="Obfs Password" />
          </ElFormItem>
          <ElFormItem label="上行带宽 (up)">
            <ElInput v-model="hysteria2.up" placeholder="例如: 100M" />
          </ElFormItem>
          <ElFormItem label="下行带宽 (down)">
            <ElInput v-model="hysteria2.down" placeholder="例如: 200M" />
          </ElFormItem>
        </ElForm>
      </ElTabPane>

      <!-- ==================== HTTP Tab ==================== -->
      <ElTabPane label="HTTP">
        <ElForm ref="httpFormRef" :model="http" label-width="120px" label-position="top">
          <ElFormItem label="协议">
            <ElSelect v-model="http.protocol">
              <ElOption value="http" label="HTTP" />
              <ElOption value="https" label="HTTPS" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem label="名称">
            <ElInput v-model="http.name" placeholder="节点名称" />
          </ElFormItem>
          <ElFormItem label="地址" required>
            <ElInput v-model="http.host" placeholder="IP / HOST" />
          </ElFormItem>
          <ElFormItem label="端口" required>
            <ElInput v-model="http.port" placeholder="端口号" type="number" />
          </ElFormItem>
          <ElFormItem label="用户名">
            <ElInput v-model="http.username" placeholder="用户名" />
          </ElFormItem>
          <ElFormItem label="密码">
            <ElInput v-model="http.password" placeholder="密码" />
          </ElFormItem>
        </ElForm>
      </ElTabPane>

      <!-- ==================== SOCKS5 Tab ==================== -->
      <ElTabPane label="SOCKS5">
        <ElForm ref="socks5FormRef" :model="socks5" label-width="120px" label-position="top">
          <ElFormItem label="名称">
            <ElInput v-model="socks5.name" placeholder="节点名称" />
          </ElFormItem>
          <ElFormItem label="地址" required>
            <ElInput v-model="socks5.host" placeholder="IP / HOST" />
          </ElFormItem>
          <ElFormItem label="端口" required>
            <ElInput v-model="socks5.port" placeholder="端口号" type="number" />
          </ElFormItem>
          <ElFormItem label="用户名">
            <ElInput v-model="socks5.username" placeholder="用户名" />
          </ElFormItem>
          <ElFormItem label="密码">
            <ElInput v-model="socks5.password" placeholder="密码" />
          </ElFormItem>
        </ElForm>
      </ElTabPane>

      <!-- ==================== AnyTLS Tab ==================== -->
      <ElTabPane label="AnyTLS">
        <ElForm ref="anytlsFormRef" :model="anytls" label-width="120px" label-position="top">
          <ElFormItem label="名称">
            <ElInput v-model="anytls.name" placeholder="节点名称" />
          </ElFormItem>
          <ElFormItem label="地址" required>
            <ElInput v-model="anytls.host" placeholder="IP / HOST" />
          </ElFormItem>
          <ElFormItem label="端口" required>
            <ElInput v-model="anytls.port" placeholder="端口号" type="number" />
          </ElFormItem>
          <ElFormItem label="认证密钥" required>
            <ElInput v-model="anytls.auth" placeholder="Authentication Key" />
          </ElFormItem>
          <ElFormItem label="SNI(Peer)">
            <ElInput v-model="anytls.sni" placeholder="SNI / Peer (可选)" />
          </ElFormItem>
          <ElFormItem label="ALPN">
            <ElInput v-model="anytls.alpn" placeholder="h2,http/1.1" />
          </ElFormItem>
          <ElFormItem label="uTLS fingerprint">
            <ElSelect v-model="anytls.fingerprint">
              <ElOption value="" label="empty" />
              <ElOption value="chrome" label="chrome" />
              <ElOption value="firefox" label="firefox" />
              <ElOption value="safari" label="safari" />
              <ElOption value="ios" label="ios" />
              <ElOption value="android" label="android" />
              <ElOption value="edge" label="edge" />
              <ElOption value="random" label="random" />
              <ElOption value="randomized" label="randomized" />
            </ElSelect>
          </ElFormItem>
          <ElFormItem label="固定证书 SHA256">
            <ElInput v-model="anytls.pinnedPeerCertSha256" :placeholder="$t('pinnedPeerCertSha256Placeholder')" class="mb-2" />
          </ElFormItem>
          <ElFormItem label="证书验证域名">
            <ElInput v-model="anytls.verifyPeerCertByName" :placeholder="$t('verifyPeerCertByNamePlaceholder')" />
          </ElFormItem>
        </ElForm>
      </ElTabPane>
    </ElTabs>

    <template #footer>
      <span v-if="!readonly" class="dialog-footer">
        <ElButton @click="dialogVisible = false">{{ t('operations.cancel') }}</ElButton>
        <ElButton type="primary" :loading="isSubmitting" @click="handleSubmit">
          {{ t('operations.saveApply') }}
        </ElButton>
      </span>
      <span v-else class="dialog-footer">
        <ElButton @click="dialogVisible = false">{{ t('operations.cancel') }}</ElButton>
      </span>
    </template>
  </ElDialog>
</template>

<style scoped>
.server-editor-tabs :deep(.el-tabs__content) {
  max-height: 60vh;
  overflow-y: auto;
}
</style>
