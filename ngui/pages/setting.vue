<script lang="ts" setup>
definePageMeta({ middleware: ['auth'] })

const { t } = useI18n()

// ====================== Data Fetching ======================
let setting = $ref<any>({})
let systemInfo = $ref<any>({ os: '', isRoot: false, tinytunSupported: false })
let remoteGFWListVersion = $ref('')
let localGFWListVersion = $ref('')

const activeNames = $ref(['gfwlist', 'transparent', 'routing', 'advanced', 'autoUpdate'])

// Fetch settings
const { data: settingRes } = await useV2Fetch<any>('setting').json()
if (settingRes.value?.data) {
  setting = settingRes.value.data.setting || {}
  localGFWListVersion = settingRes.value.data.localGFWListVersion || ''
  system.value.gfwlist = localGFWListVersion
}

// Fetch remote GFWList version
const { data: gfwVerRes } = await useV2Fetch<any>('remoteGFWListVersion').json()
if (gfwVerRes.value?.data) {
  remoteGFWListVersion = gfwVerRes.value.data.remoteGFWListVersion || ''
}

// Fetch version info
const { data: verRes } = await useV2Fetch<any>('version').json()
if (verRes.value?.data) {
  systemInfo = verRes.value.data
  system.value.os = systemInfo.os
  system.value.isRoot = systemInfo.isRoot
}

// ====================== Ports ======================
let portsData = $ref<any>({})
let showPortsDialog = $ref(false)
const loadPorts = async () => {
  const { data } = await useV2Fetch<any>('ports').json()
  if (data.value?.code === 'SUCCESS') portsData = data.value.data
  showPortsDialog = true
}
const savePorts = async () => {
  const { data } = await useV2Fetch<any>('ports').put(portsData).json()
  if (data.value?.code === 'SUCCESS') {
    ElMessage.success(t('common.success'))
    showPortsDialog = false
  }
}

// ====================== DNS Rules ======================
let dnsRules = $ref<any[]>([])
let showDnsDialog = $ref(false)
const DEFAULT_DNS_RULES = [
  { server: 'localhost', domains: 'geosite:private', outbound: 'direct' },
  { server: '223.5.5.5', domains: 'geosite:cn', outbound: 'direct' },
  { server: '8.8.8.8', domains: '', outbound: 'proxy' },
]
const loadDnsRules = async () => {
  const { data } = await useV2Fetch<any>('dnsRules').json()
  if (data.value?.code === 'SUCCESS' && data.value.data?.rules?.length) {
    dnsRules = data.value.data.rules.map((r: any) => ({
      server: r.server || '', domains: r.domains || '', outbound: r.outbound || 'direct',
    }))
  } else {
    dnsRules = DEFAULT_DNS_RULES.map(r => ({ ...r }))
  }
  showDnsDialog = true
}
const addDnsRule = () => { dnsRules.push({ server: '', domains: '', outbound: 'direct' }) }
const removeDnsRule = (i: number) => { dnsRules.splice(i, 1) }
const resetDnsRules = () => { dnsRules = DEFAULT_DNS_RULES.map(r => ({ ...r })) }
const saveDnsRules = async () => {
  const valid = dnsRules.filter((r: any) => r.server.trim())
  if (!valid.length) { ElMessage.warning('至少需要一条规则'); return }
  const { data } = await useV2Fetch<any>('dnsRules').put(valid).json()
  if (data.value?.code === 'SUCCESS') { ElMessage.success(t('common.success')); showDnsDialog = false }
}

// ====================== RoutingA ======================
let routingAText = $ref('')
let showRoutingADialog = $ref(false)
const loadRoutingA = async () => {
  const { data } = await useV2Fetch<any>('routingA').json()
  if (data.value?.code === 'SUCCESS') routingAText = data.value.data.routingA || ''
  showRoutingADialog = true
}
const saveRoutingA = async () => {
  const { data } = await useV2Fetch<any>('routingA').put({ routingA: routingAText }).json()
  if (data.value?.code === 'SUCCESS') { ElMessage.success(t('common.success')); showRoutingADialog = false }
}

// ====================== Custom Inbound ======================
let inbounds = $ref<any[]>([])
let inboundForm = $ref({ tag: '', protocol: 'socks', port: '' })
let inboundAdding = $ref(false)
let showInboundDialog = $ref(false)
const loadInbounds = async () => {
  const { data } = await useV2Fetch<any>('customInbound').json()
  if (data.value?.code === 'SUCCESS') inbounds = data.value.data.inbounds || []
  showInboundDialog = true
}
const addInbound = async () => {
  if (!inboundForm.tag || !inboundForm.port) { ElMessage.warning('请填写完整'); return }
  inboundAdding = true
  try {
    const { data } = await useV2Fetch<any>('customInbound').post({
      tag: inboundForm.tag.trim(), protocol: inboundForm.protocol, port: Number(inboundForm.port),
    }).json()
    if (data.value?.code === 'SUCCESS') {
      inbounds = data.value.data.inbounds || []
      inboundForm = { tag: '', protocol: 'socks', port: '' }
    }
  } finally { inboundAdding = false }
}
const deleteInbound = async (tag: string) => {
  try {
    await ElMessageBox.confirm(`确定删除入站 "${tag}"？`, t('operations.delete'), { type: 'warning' })
  } catch { return }
  const { data } = await useV2Fetch<any>('customInbound').delete({ tag }).json()
  if (data.value?.code === 'SUCCESS') inbounds = data.value.data.inbounds || []
}

// ====================== Sniffing Excluded Domains ======================
let sniffingDomains = $ref('')
let showSniffingDialog = $ref(false)
const loadSniffingDomains = async () => {
  const { data } = await useV2Fetch<any>('domainsExcluded').json()
  if (data.value?.code === 'SUCCESS') sniffingDomains = data.value.data.domainsExcluded || ''
  showSniffingDialog = true
}
const saveSniffingDomains = async () => {
  const { data } = await useV2Fetch<any>('domainsExcluded').put({ domainsExcluded: sniffingDomains }).json()
  if (data.value?.code === 'SUCCESS') { ElMessage.success(t('common.success')); showSniffingDialog = false }
}

// ====================== TPROXY White IP Groups ======================
let tproxyWhiteIps = $ref('')
let showTproxyDialog = $ref(false)
const loadTproxyWhiteIps = async () => {
  const { data } = await useV2Fetch<any>('tproxyWhiteIpGroups').json()
  if (data.value?.code === 'SUCCESS') tproxyWhiteIps = data.value.data.tproxyWhiteIpGroups || ''
  showTproxyDialog = true
}
const saveTproxyWhiteIps = async () => {
  const { data } = await useV2Fetch<any>('tproxyWhiteIpGroups').put({ tproxyWhiteIpGroups: tproxyWhiteIps }).json()
  if (data.value?.code === 'SUCCESS') { ElMessage.success(t('common.success')); showTproxyDialog = false }
}

// ====================== Network Interfaces (for TUN bypass) ======================
let availableInterfaces = $ref<any[]>([])
let tunBypassInterfacesList = $ref<string[]>([])
let tunBypassCustom = $ref('')
const tunBypassInterfacesComputed = computed({
  get: () => {
    const parts: string[] = [...tunBypassInterfacesList]
    if (tunBypassCustom.trim()) {
      parts.push(...tunBypassCustom.split(',').map(s => s.trim()).filter(s => s.length))
    }
    return [...new Set(parts)].join(',')
  },
  set: (val: string) => {
    const parts = val ? val.split(',').map(s => s.trim()).filter(s => s.length) : []
    const known = (availableInterfaces || []).map((i: any) => i.name)
    tunBypassInterfacesList = parts.filter(p => known.includes(p))
    tunBypassCustom = parts.filter(p => !known.includes(p)).join(',')
  },
})
const fetchNetworkInterfaces = async () => {
  const { data } = await useV2Fetch<any>('networkInterfaces').json()
  if (data.value?.data?.interfaces) {
    availableInterfaces = data.value.data.interfaces
    if (setting.tunBypassInterfaces) tunBypassInterfacesComputed.value = setting.tunBypassInterfaces
  }
}

// Initialize TUN bypass from setting
if (setting.tunBypassInterfaces) {
  tunBypassInterfacesComputed.value = setting.tunBypassInterfaces
}

// Watch transparentType for TUN init
watch(() => setting.transparentType, (val) => {
  if (val === 'tun' && systemInfo.tinytunSupported) fetchNetworkInterfaces()
})

// ====================== TUN Route Script ======================
let showTunRouteDialog = $ref(false)
let tunRouteShellType = $ref('')
let tunRouteShellPath = $ref('')
let tunSetupScript = $ref('')
let tunTeardownScript = $ref('')
const openTunRouteScript = () => {
  tunRouteShellType = setting.tunRouteShellType || ''
  tunRouteShellPath = setting.tunRouteShellPath || ''
  tunSetupScript = setting.tunSetupScript || ''
  tunTeardownScript = setting.tunTeardownScript || ''
  showTunRouteDialog = true
}
const saveTunRouteScript = () => {
  setting.tunRouteShellType = tunRouteShellType
  setting.tunRouteShellPath = tunRouteShellPath
  setting.tunSetupScript = tunSetupScript
  setting.tunTeardownScript = tunTeardownScript
  showTunRouteDialog = false
}

// ====================== TUN Exclude Processes ======================
let showTunExcludeDialog = $ref(false)
let tunExcludeProcessesInput = $ref('')
const openTunExcludeProcesses = () => {
  tunExcludeProcessesInput = setting.tunExcludeProcesses || ''
  showTunExcludeDialog = true
}
const saveTunExcludeProcesses = () => {
  setting.tunExcludeProcesses = tunExcludeProcessesInput
  showTunExcludeDialog = false
}

// ====================== Update GFWList ======================
const updateGFWList = async () => {
  const { data } = await useV2Fetch<any>('gfwList').put().json()
  if (data.value?.code === 'SUCCESS') {
    system.value.gfwlist = data.value.data.localGFWListVersion
    localGFWListVersion = data.value.data.localGFWListVersion
    ElMessage.success(t('common.success'))
  }
}

// ====================== Save Settings ======================
const updateSetting = async () => {
  // Prepare payload
  const payload: any = {
    ...setting,
    pacAutoUpdateIntervalHour: parseInt(setting.pacAutoUpdateIntervalHour) || 0,
    subscriptionAutoUpdateIntervalHour: parseInt(setting.subscriptionAutoUpdateIntervalHour) || 0,
    mux: parseInt(setting.mux) || 8,
    tunBypassInterfaces: tunBypassInterfacesComputed.value,
  }
  // Remove system fields that shouldn't be sent
  delete payload.lite

  const { data } = await useV2Fetch<any>('setting').put(payload).json()
  if (data.value?.code === 'SUCCESS') ElMessage.success(t('common.success'))
}

// ====================== Computed ======================
const lite = computed(() => system.value.lite && String(system.value.lite) !== 'false')
const showTunOptions = computed(() =>
  setting.transparent !== 'close' && setting.transparentType === 'tun' && systemInfo.tinytunSupported
)
const showTproxyOptions = computed(() =>
  setting.transparent !== 'close' && (setting.transparentType === 'tproxy' || setting.transparentType === 'redirect')
)
</script>

<template>
  <div class="max-w-4xl mx-auto px-4 py-6">
    <h1 class="text-xl font-bold mb-4">{{ $t('common.setting') }}</h1>

    <ElCollapse v-model="activeNames">
      <!-- ==================== GFWList ==================== -->
      <ElCollapseItem title="GFWList" name="gfwlist">
        <div class="flex items-center gap-4 flex-wrap">
          <span>{{ $t('common.latest') }}:</span>
          <ElLink href="https://github.com/v2rayA/dist-v2ray-rules-dat/releases" target="_blank" type="primary">
            {{ remoteGFWListVersion }}
          </ElLink>
          <span>{{ $t('common.local') }}:</span>
          <span class="font-mono">{{ localGFWListVersion || $t('common.none') }}</span>
          <ElButton size="small" @click="updateGFWList">{{ $t('operations.update') }}</ElButton>
        </div>
      </ElCollapseItem>

      <!-- ==================== 透明代理 ==================== -->
      <ElCollapseItem :title="$t('setting.transparentProxy')" name="transparent">
        <ElForm label-width="160px" label-position="left" size="default">
          <ElFormItem :label="$t('setting.transparentProxy')">
            <div class="flex items-center gap-2 flex-wrap">
              <ElSelect v-model="setting.transparent" size="small" style="width:220px">
                <ElOption value="close" :label="$t('setting.options.off')" />
                <ElOption value="proxy" :label="`${$t('setting.options.on')}: ${$t('setting.options.global')}`" />
                <ElOption value="whitelist" :label="`${$t('setting.options.on')}: ${$t('setting.options.whitelistCn')}`" />
                <ElOption value="gfwlist" :label="`${$t('setting.options.on')}: ${$t('setting.options.gfwlist')}`" />
                <ElOption value="pac" :label="`${$t('setting.options.on')}: ${$t('setting.options.sameAsPacMode')}`" />
              </ElSelect>
              <ElCheckboxButton v-show="!lite" v-model="setting.ipforward" :label="$t('setting.ipForwardOn')" size="small" />
              <ElCheckboxButton v-model="setting.portSharing" :label="$t('setting.portSharingOn')" size="small" />
            </div>
          </ElFormItem>

          <ElFormItem v-if="setting.transparent !== 'close'" :label="$t('setting.transparentType')">
            <div class="flex items-center gap-2 flex-wrap">
              <ElSelect v-model="setting.transparentType" size="small" style="width:180px">
                <ElOption v-show="!lite && systemInfo.os === 'linux'" value="redirect" label="redirect" />
                <ElOption v-show="!lite && systemInfo.os === 'linux'" value="tproxy" label="tproxy" />
                <ElOption v-show="!lite" value="tun" :disabled="!systemInfo.tinytunSupported"
                  :label="`tun (TinyTun)${!systemInfo.tinytunSupported ? ' — 未集成' : ''}`" />
                <ElOption
                  v-show="!(systemInfo.isRoot && (systemInfo.os === 'linux' || systemInfo.os === 'darwin'))"
                  value="system_proxy" label="system proxy" />
              </ElSelect>

              <!-- TPROXY white IP groups -->
              <ElButton v-if="setting.transparentType === 'tproxy'" size="small" @click="loadTproxyWhiteIps">
                {{ $t('operations.tproxyWhiteIpGroups') || 'TPROXY IP 白名单' }}
              </ElButton>

              <!-- TUN auto route -->
              <ElCheckboxButton v-if="showTunOptions" v-model="setting.tunAutoRoute" :label="$t('setting.tunAutoRoute')" size="small" />
              <ElButton v-if="showTunOptions && !setting.tunAutoRoute" size="small" @click="openTunRouteScript">
                配置 TUN 路由脚本
              </ElButton>
            </div>
          </ElFormItem>

          <!-- TPROXY excluded interfaces -->
          <ElFormItem
            v-if="showTproxyOptions"
            :label="$t('setting.tproxyExcludedInterfaces') || 'TPROXY 排除接口'">
            <ElInput v-model="setting.tproxyExcludedInterfaces" size="small" style="width:320px"
              placeholder="docker*, veth*, wg*, ppp*, br-*" />
          </ElFormItem>

          <!-- TUN bypass interfaces -->
          <ElFormItem v-if="showTunOptions" label="TUN 绕过接口">
            <div class="flex items-center gap-2 flex-wrap">
              <ElSelect
                v-model="tunBypassInterfacesList"
                multiple
                size="small"
                placeholder="选择接口"
                style="min-width:200px"
                :disabled="!availableInterfaces.filter((i: any) => !i.isLoopback).length"
              >
                <ElOption
                  v-for="iface in availableInterfaces.filter((i: any) => !i.isLoopback)"
                  :key="iface.name" :value="iface.name"
                  :label="`${iface.name}${iface.addrs?.length ? ' (' + iface.addrs.join(', ') + ')' : ''}`"
                />
              </ElSelect>
              <ElInput v-model="tunBypassCustom" size="small" style="width:200px"
                placeholder="自定义接口, 逗号分隔" />
            </div>
          </ElFormItem>

          <!-- TUN process backend -->
          <ElFormItem v-if="showTunOptions && systemInfo.os === 'linux'" label="TUN 进程后端">
            <ElSelect v-model="setting.tunProcessBackend" size="small" style="width:160px">
              <ElOption value="" label="TUN (默认)" />
              <ElOption value="ebpf" label="eBPF" />
            </ElSelect>
          </ElFormItem>

          <!-- TUN exclude processes -->
          <ElFormItem v-if="showTunOptions" label="TUN 排除进程">
            <div class="flex items-center gap-2">
              <ElInput :model-value="setting.tunExcludeProcesses" size="small" readonly style="width:240px"
                placeholder="排除的进程列表" />
              <ElButton size="small" @click="openTunExcludeProcesses">{{ $t('operations.configure') }}</ElButton>
            </div>
          </ElFormItem>
        </ElForm>
      </ElCollapseItem>

      <!-- ==================== 路由规则 ==================== -->
      <ElCollapseItem :title="$t('setting.pacMode')" name="routing">
        <ElForm label-width="160px" label-position="left" size="default">
          <ElFormItem :label="$t('setting.pacMode')">
            <div class="flex items-center gap-2">
              <ElSelect v-model="setting.pacMode" size="small" style="width:200px">
                <ElOption value="whitelist" :label="$t('setting.options.whitelistCn')" />
                <ElOption value="gfwlist" :label="$t('setting.options.gfwlist')" />
                <ElOption value="routingA" label="RoutingA" />
              </ElSelect>
              <ElButton v-if="setting.pacMode === 'routingA'" size="small" @click="loadRoutingA">
                {{ $t('operations.configure') }}
              </ElButton>
            </div>
          </ElFormItem>

          <ElFormItem :label="$t('setting.preventDnsSpoofing')">
            <div class="flex items-center gap-2">
              <ElSelect v-model="setting.antipollution" size="small" style="width:200px">
                <ElOption value="closed" :label="$t('setting.options.closed')" />
                <ElOption value="none" :label="$t('setting.options.antiDnsHijack')" />
                <ElOption value="dnsforward" :label="$t('setting.options.forwardDnsRequest')" />
                <ElOption value="doh" :label="$t('setting.options.doh')" />
                <ElOption value="advanced" :label="$t('setting.options.advanced')" />
              </ElSelect>
            </div>
          </ElFormItem>

          <ElFormItem v-if="setting.showSpecialMode" :label="$t('setting.specialMode')">
            <ElSelect v-model="setting.specialMode" size="small" style="width:200px">
              <ElOption value="none" :label="$t('setting.options.closed')" />
              <ElOption value="supervisor" label="supervisor" />
              <ElOption v-show="setting.antipollution !== 'closed'" value="fakedns" label="fakedns" />
            </ElSelect>
          </ElFormItem>
        </ElForm>
      </ElCollapseItem>

      <!-- ==================== 高级设置 ==================== -->
      <ElCollapseItem :title="$t('setting.advanced') || '高级设置'" name="advanced">
        <ElForm label-width="160px" label-position="left" size="default">
          <!-- TCPFastOpen -->
          <ElFormItem label="TCPFastOpen">
            <ElSelect v-model="setting.tcpFastOpen" size="small" style="width:160px">
              <ElOption value="default" :label="$t('setting.options.default')" />
              <ElOption value="yes" :label="$t('setting.options.on')" />
              <ElOption value="no" :label="$t('setting.options.off')" />
            </ElSelect>
          </ElFormItem>

          <!-- Log Level -->
          <ElFormItem :label="$t('setting.logLevel')">
            <ElSelect v-model="setting.logLevel" size="small" style="width:160px">
              <ElOption value="trace" :label="$t('setting.options.trace')" />
              <ElOption value="debug" :label="$t('setting.options.debug')" />
              <ElOption value="info" :label="$t('setting.options.info')" />
              <ElOption value="warn" :label="$t('setting.options.warn')" />
              <ElOption value="error" :label="$t('setting.options.error')" />
            </ElSelect>
          </ElFormItem>

          <!-- MUX -->
          <ElFormItem :label="$t('setting.mux')">
            <div class="flex items-center gap-2">
              <ElSelect v-model="setting.muxOn" size="small" style="width:120px">
                <ElOption value="no" :label="$t('setting.options.off')" />
                <ElOption value="yes" :label="$t('setting.options.on')" />
              </ElSelect>
              <ElInput v-if="setting.muxOn === 'yes'" v-model="setting.mux" size="small" style="width:100px"
                :placeholder="$t('setting.concurrency')" type="number" min="1" max="1024" />
            </div>
          </ElFormItem>

          <!-- Inbound Sniffing -->
          <ElFormItem :label="$t('setting.inboundSniffing') || '嗅探'">
            <div class="flex items-center gap-2 flex-wrap">
              <ElSelect v-model="setting.inboundSniffing" size="small" style="width:200px">
                <ElOption value="disable" :label="$t('setting.options.off')" />
                <ElOption value="http,tls" label="Http + TLS" />
                <ElOption value="http,tls,quic" label="Http + TLS + Quic" />
              </ElSelect>
              <template v-if="setting.inboundSniffing !== 'disable'">
                <ElButton size="small" @click="loadSniffingDomains">排除域名</ElButton>
                <ElCheckboxButton v-model="setting.routeOnly" :label="'RouteOnly'" size="small" />
              </template>
            </div>
          </ElFormItem>
        </ElForm>
      </ElCollapseItem>

      <!-- ==================== 自动更新 ==================== -->
      <ElCollapseItem :title="$t('setting.autoUpdate') || '自动更新'" name="autoUpdate">
        <ElForm label-width="200px" label-position="left" size="default">
          <!-- Auto update GFWList -->
          <ElFormItem v-show="setting.pacMode === 'gfwlist' || setting.transparent === 'gfwlist'"
            :label="$t('setting.autoUpdateGfwlist')">
            <div class="flex items-center gap-2">
              <ElSelect v-model="setting.pacAutoUpdateMode" size="small" style="width:220px">
                <ElOption value="none" :label="$t('setting.options.off')" />
                <ElOption value="auto_update" :label="$t('setting.options.updateGfwlistWhenStart')" />
                <ElOption value="auto_update_at_intervals" :label="$t('setting.options.updateGfwlistAtIntervals')" />
              </ElSelect>
              <ElInput
                v-if="setting.pacAutoUpdateMode === 'auto_update_at_intervals'"
                v-model="setting.pacAutoUpdateIntervalHour" size="small" style="width:80px" type="number" min="1"
                placeholder="小时" />
            </div>
          </ElFormItem>

          <!-- Auto update subscription -->
          <ElFormItem :label="$t('setting.autoUpdateSub')">
            <div class="flex items-center gap-2">
              <ElSelect v-model="setting.subscriptionAutoUpdateMode" size="small" style="width:220px">
                <ElOption value="none" :label="$t('setting.options.off')" />
                <ElOption value="auto_update" :label="$t('setting.options.updateSubWhenStart')" />
                <ElOption value="auto_update_at_intervals" :label="$t('setting.options.updateSubAtIntervals')" />
              </ElSelect>
              <ElInput
                v-if="setting.subscriptionAutoUpdateMode === 'auto_update_at_intervals'"
                v-model="setting.subscriptionAutoUpdateIntervalHour" size="small" style="width:80px" type="number" min="1"
                placeholder="小时" />
            </div>
          </ElFormItem>

          <!-- Prefer mode when update -->
          <ElFormItem :label="$t('setting.preferModeWhenUpdate')">
            <ElSelect v-model="setting.proxyModeWhenSubscribe" size="small" style="width:220px">
              <ElOption value="direct"
                :label="setting.transparent === 'close' || lite
                  ? $t('setting.options.direct')
                  : $t('setting.options.dependTransparentMode')" />
              <ElOption value="proxy" :label="$t('setting.options.global')" />
              <ElOption value="pac" :label="$t('setting.options.pac')" />
            </ElSelect>
          </ElFormItem>
        </ElForm>
      </ElCollapseItem>
    </ElCollapse>

    <!-- ==================== 操作按钮行 ==================== -->
    <div class="flex items-center gap-3 mt-6 flex-wrap">
      <ElButton @click="loadPorts">
        {{ $t('customAddressPort.title') || '地址与端口' }}
      </ElButton>
      <ElButton @click="loadDnsRules">
        {{ $t('dns.title') || 'DNS 设置' }}
      </ElButton>
      <ElButton @click="loadInbounds">
        {{ $t('customInbound.title') || '自定义入站' }}
      </ElButton>
      <div class="flex-1" />
      <ElButton type="primary" @click="updateSetting">
        {{ $t('operations.saveApply') }}
      </ElButton>
    </div>

    <!-- ==================== Ports Dialog ==================== -->
    <ElDialog v-model="showPortsDialog" :title="$t('customAddressPort.title') || '地址与端口'" width="520px" :close-on-click-modal="false">
      <ElForm v-if="portsData" label-width="180px" label-position="left" size="default">
        <ElFormItem :label="$t('customAddressPort.serviceAddress') || '服务地址'">
          <ElInput v-model="portsData.backendAddress" size="small" placeholder="http://localhost:2017" />
        </ElFormItem>
        <ElFormItem :label="$t('customAddressPort.portSocks5') || 'SOCKS5 端口'">
          <ElInput v-model="portsData.socks5" size="small" type="number" min="0" />
        </ElFormItem>
        <ElFormItem :label="$t('customAddressPort.portHttp') || 'HTTP 端口'">
          <ElInput v-model="portsData.http" size="small" type="number" min="0" />
        </ElFormItem>
        <ElFormItem :label="$t('customAddressPort.portSocks5WithPac') || 'SOCKS5 (PAC) 端口'">
          <ElInput v-model="portsData.socks5WithPac" size="small" type="number" min="0" />
        </ElFormItem>
        <ElFormItem :label="$t('customAddressPort.portHttpWithPac') || 'HTTP (PAC) 端口'">
          <ElInput v-model="portsData.httpWithPac" size="small" type="number" min="0" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="showPortsDialog = false">{{ $t('operations.cancel') }}</ElButton>
        <ElButton type="primary" @click="savePorts">{{ $t('operations.confirm') }}</ElButton>
      </template>
    </ElDialog>

    <!-- ==================== DNS Rules Dialog ==================== -->
    <ElDialog v-model="showDnsDialog" :title="$t('dns.title') || 'DNS 设置'" width="700px" :close-on-click-modal="false">
      <div class="mb-3 flex items-center gap-2">
        <ElButton size="small" type="primary" @click="addDnsRule">+ {{ $t('dns.addRule') || '添加规则' }}</ElButton>
        <ElButton size="small" @click="resetDnsRules">{{ $t('dns.resetDefault') || '重置默认' }}</ElButton>
      </div>
      <div class="border rounded overflow-hidden">
        <div class="grid grid-cols-[220px_1fr_120px_50px] bg-gray-100 dark:bg-gray-700 text-xs font-semibold px-3 py-2 border-b">
          <div>{{ $t('dns.colServer') || 'DNS 服务器' }}</div>
          <div>{{ $t('dns.colDomains') || '域名' }}</div>
          <div>{{ $t('dns.colOutbound') || '出站' }}</div>
          <div></div>
        </div>
        <div v-for="(rule, i) in dnsRules" :key="i"
          class="grid grid-cols-[220px_1fr_120px_50px] items-start border-b last:border-b-0 px-3 py-2">
          <ElInput v-model="rule.server" size="small" :placeholder="$t('dns.serverPlaceholder') || '例如: 8.8.8.8'" />
          <ElInput v-model="rule.domains" size="small" type="textarea" :rows="2"
            :placeholder="$t('dns.domainsPlaceholder') || '例如: geosite:cn'" class="font-mono text-xs" />
          <ElSelect v-model="rule.outbound" size="small">
            <ElOption value="direct" label="direct" />
            <ElOption value="proxy" label="proxy" />
          </ElSelect>
          <div class="flex justify-center pt-1">
            <ElButton size="small" type="danger" :icon="Delete" circle @click="removeDnsRule(i)" />
          </div>
        </div>
      </div>
      <template #footer>
        <ElButton @click="showDnsDialog = false">{{ $t('operations.cancel') }}</ElButton>
        <ElButton type="primary" @click="saveDnsRules">{{ $t('operations.save') }}</ElButton>
      </template>
    </ElDialog>

    <!-- ==================== RoutingA Dialog ==================== -->
    <ElDialog v-model="showRoutingADialog" title="RoutingA" width="600px" :close-on-click-modal="false">
      <ElInput v-model="routingAText" type="textarea" :rows="16" class="font-mono text-sm"
        :placeholder="$t('routingA.messages.0') || '请输入 RoutingA 规则'" />
      <template #footer>
        <ElButton @click="showRoutingADialog = false">{{ $t('operations.cancel') }}</ElButton>
        <ElButton type="primary" @click="saveRoutingA">{{ $t('operations.save') }}</ElButton>
      </template>
    </ElDialog>

    <!-- ==================== Custom Inbound Dialog ==================== -->
    <ElDialog v-model="showInboundDialog" :title="$t('customInbound.title') || '自定义入站'" width="560px" :close-on-click-modal="false">
      <ElTable v-if="inbounds.length" :data="inbounds" size="small" class="mb-4" border>
        <ElTableColumn prop="tag" :label="$t('customInbound.tag') || '标签'" width="180" />
        <ElTableColumn prop="protocol" :label="$t('customInbound.protocol') || '协议'" width="80" />
        <ElTableColumn prop="port" :label="$t('customInbound.port') || '端口'" width="80" />
        <ElTableColumn :label="$t('operations.name')" width="80">
          <template #default="scope">
            <ElButton size="small" type="danger" @click="deleteInbound(scope.row.tag)">
              {{ $t('operations.delete') }}
            </ElButton>
          </template>
        </ElTableColumn>
      </ElTable>
      <div v-else class="text-center text-gray-400 py-4">{{ $t('customInbound.empty') || '暂无自定义入站' }}</div>
      <div class="border rounded p-3 bg-gray-50 dark:bg-gray-800">
        <p class="text-sm font-semibold mb-2">{{ $t('customInbound.addNew') || '新增入站' }}</p>
        <div class="flex items-center gap-2 flex-wrap">
          <ElInput v-model="inboundForm.tag" size="small" style="width:140px"
            :placeholder="$t('customInbound.tagPlaceholder') || '标签名'" />
          <ElSelect v-model="inboundForm.protocol" size="small" style="width:100px">
            <ElOption value="socks" label="SOCKS" />
            <ElOption value="http" label="HTTP" />
          </ElSelect>
          <ElInput v-model="inboundForm.port" size="small" style="width:100px" type="number" min="1" max="65535"
            :placeholder="$t('customInbound.portPlaceholder') || '端口'" />
          <ElButton size="small" type="primary" :loading="inboundAdding" @click="addInbound">
            {{ $t('operations.add') }}
          </ElButton>
        </div>
      </div>
      <template #footer>
        <ElButton @click="showInboundDialog = false">{{ $t('operations.close') }}</ElButton>
      </template>
    </ElDialog>

    <!-- ==================== Sniffing Domains Excluded Dialog ==================== -->
    <ElDialog v-model="showSniffingDialog" title="嗅探排除域名" width="500px" :close-on-click-modal="false">
      <ElInput v-model="sniffingDomains" type="textarea" :rows="8" class="font-mono text-sm"
        placeholder="每行一个域名，例如:&#10;example.com&#10;*.google.com" />
      <template #footer>
        <ElButton @click="showSniffingDialog = false">{{ $t('operations.cancel') }}</ElButton>
        <ElButton type="primary" @click="saveSniffingDomains">{{ $t('operations.save') }}</ElButton>
      </template>
    </ElDialog>

    <!-- ==================== TPROXY White IP Groups Dialog ==================== -->
    <ElDialog v-model="showTproxyDialog" :title="$t('operations.tproxyWhiteIpGroups') || 'TPROXY IP 白名单'" width="500px" :close-on-click-modal="false">
      <ElInput v-model="tproxyWhiteIps" type="textarea" :rows="10" class="font-mono text-sm"
        placeholder="每行一个 IP 段，例如:&#10;10.0.0.0/8&#10;172.16.0.0/12" />
      <template #footer>
        <ElButton @click="showTproxyDialog = false">{{ $t('operations.cancel') }}</ElButton>
        <ElButton type="primary" @click="saveTproxyWhiteIps">{{ $t('operations.save') }}</ElButton>
      </template>
    </ElDialog>

    <!-- ==================== TUN Route Script Dialog ==================== -->
    <ElDialog v-model="showTunRouteDialog" title="TUN 路由脚本" width="600px" :close-on-click-modal="false">
      <ElForm label-width="120px" label-position="left" size="default">
        <ElFormItem label="Shell 类型">
          <ElSelect v-model="tunRouteShellType" size="small" style="width:160px">
            <ElOption value="" label="自动检测" />
            <ElOption value="bash" label="bash" />
            <ElOption value="sh" label="sh" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="Shell 路径">
          <ElInput v-model="tunRouteShellPath" size="small" placeholder="/bin/bash" />
        </ElFormItem>
        <ElFormItem label="Setup 脚本">
          <ElInput v-model="tunSetupScript" type="textarea" :rows="5" class="font-mono text-sm" />
        </ElFormItem>
        <ElFormItem label="Teardown 脚本">
          <ElInput v-model="tunTeardownScript" type="textarea" :rows="5" class="font-mono text-sm" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="showTunRouteDialog = false">{{ $t('operations.cancel') }}</ElButton>
        <ElButton type="primary" @click="saveTunRouteScript">{{ $t('operations.save') }}</ElButton>
      </template>
    </ElDialog>

    <!-- ==================== TUN Exclude Processes Dialog ==================== -->
    <ElDialog v-model="showTunExcludeDialog" title="TUN 排除进程" width="500px" :close-on-click-modal="false">
      <ElInput v-model="tunExcludeProcessesInput" type="textarea" :rows="8" class="font-mono text-sm"
        placeholder="每行一个进程路径，例如:&#10;/usr/bin/example&#10;/opt/app/bin/*" />
      <template #footer>
        <ElButton @click="showTunExcludeDialog = false">{{ $t('operations.cancel') }}</ElButton>
        <ElButton type="primary" @click="saveTunExcludeProcesses">{{ $t('operations.save') }}</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<style scoped>
.font-mono {
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Menlo, Consolas, monospace;
}
</style>
