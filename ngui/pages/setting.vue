<script lang="ts" setup>
definePageMeta({ middleware: ['auth'] })

const { t } = useI18n()

let setting = $ref<any>()
const { data } = await useV2Fetch<any>('setting').json()

system.value.gfwlist = data.value.data.localGFWListVersion

setting = data.value.data.setting

const { data: { value: { data: { remoteGFWListVersion } } } }
  = await useV2Fetch<any>('remoteGFWListVersion').json()

const updateGFWList = async() => {
  const { data } = await useV2Fetch<any>('gfwList').put().json()
  if (data.value.code === 'SUCCESS') {
    system.value.gfwlist = data.value.data.localGFWListVersion

    ElMessage.success(t('common.success'))
  }
}

const updateSetting = async() => {
  const { data } = await useV2Fetch<any>('setting').put(setting).json()
  if (data.value.code === 'SUCCESS')
    ElMessage.success(t('common.success'))
}
</script>

<template>
  <div class="flex-col mx-auto w-108">
    <div>
      <div>GFWList</div>
      <div>{{ $t('common.latest') }}:</div>
      <ElLink href="https://github.com/v2ray-a/dist-v2ray-rules-dat/releases">
        {{ remoteGFWListVersion }}
      </ElLink>

      <div>{{ $t('common.local') }}:</div>
      <div>{{ system.gfwlist }}</div>
      <ElButton size="small" @click="updateGFWList">{{ $t('operations.update') }}</ElButton>
    </div>

    <div>
      <div>{{ $t('setting.transparentProxy') }}</div>
      <ElSelect v-model="setting.transparent" size="small">
        <ElOption value="close" :label="$t('setting.options.off')" />
        <ElOption value="proxy" :label="`${$t('setting.options.on')}:${$t('setting.options.global')}`" />
        <ElOption value="whitelist" :label="`${$t('setting.options.on')}:${$t('setting.options.whitelistCn')}`" />
        <ElOption value="gfwlist" :label="`${$t('setting.options.on')}:${$t('setting.options.gfwlist')}`" />
        <ElOption value="pac" :label="`${$t('setting.options.on')}:${$t('setting.options.sameAsPacMode')}`" />
      </ElSelect>

      <ElCheckboxButton v-show="!system.lite" v-model="setting.ipforward">
        {{ $t('setting.ipForwardOn') }}
      </ElCheckboxButton>

      <ElCheckboxButton v-model="setting.portSharing">
        {{ $t('setting.portSharingOn') }}
      </ElCheckboxButton>
    </div>

    <div v-if="setting.transparent !== 'close'">
      <div>{{ $t('setting.transparentType') }}</div>
      <ElSelect v-model="setting.transparentType" size="small">
        <ElOption v-show="!system.lite" value="redirect">redirect</ElOption>
        <ElOption v-show="!system.lite" value="tproxy">tproxy</ElOption>
        <ElOption value="system_proxy">system proxy</ElOption>
      </ElSelect>
    </div>

    <div>
      <div>{{ $t('setting.pacMode') }}</div>
      <ElSelect v-model="setting.pacMode" size="small">
        <ElOption value="whitelist" :label="$t('setting.options.whitelistCn')" />
        <ElOption value="gfwlist" :label="$t('setting.options.gfwlist')" />
        <!-- <ElOption value="custom" :label="$t('setting.options.customRouting')" /> -->
        <ElOption value="routingA" label="RoutingA" />
      </ElSelect>

      <ElButton v-if="setting.pacMode === 'custom'">{{ $t('operations.configure') }}</ElButton>
      <ElButton v-if="setting.pacMode === 'routingA'">{{ $t('operations.configure') }}</ElButton>
    </div>

    <div>
      <div>{{ $t('setting.preventDnsSpoofing') }}</div>
      <ElSelect v-model="setting.antipollution" size="small">
        <ElOption value="closed" :label="$t('setting.options.closed')" />
        <ElOption value="none" :label=" $t('setting.options.antiDnsHijack')" />
        <ElOption value="dnsforward" :label="$t('setting.options.forwardDnsRequest')" />
        <ElOption value="doh" :label="$t('setting.options.doh')" />
        <ElOption value="advanced" :label="$t('setting.options.advanced')" />
      </ElSelect>

      <ElButton v-if="setting.antipollution === 'advanced'">{{ $t('operations.configure') }}</ElButton>
    </div>

    <div v-if="setting.showSpecialMode">
      <div>{{ $t('setting.specialMode') }}</div>
      <ElSelect v-model="setting.specialMode" size="small">
        <ElOption value="none" :label="$t('setting.options.closed')" />
        <ElOption value="supervisor">supervisor</ElOption>
        <ElOption v-show="setting.antipollution !== 'closed'" value="fakedns">
          fakedns
        </ElOption>
      </ElSelect>
    </div>

    <div>
      <div>TCPFastOpen</div>
      <ElSelect v-model="setting.tcpFastOpen" size="small">
        <ElOption value="default" :label="$t('setting.options.default')" />
        <ElOption value="yes" :label="$t('setting.options.on')" />
        <ElOption value="no" :label="$t('setting.options.off')" />
      </ElSelect>
    </div>

    <div>
      <div>{{ $t('setting.mux') }}</div>
      <ElSelect v-model="setting.muxOn" size="small">
        <ElOption value="no" :label="$t('setting.options.off')" />
        <ElOption value="yes" :label="$t('setting.options.on')" />
      </ElSelect>
      <ElInput
        v-if="setting.muxOn === 'yes'"
        ref="muxinput" v-model="setting.mux"
        :placeholder="$t('setting.concurrency')"
        type="number" min="1" max="1024"
      />
    </div>

    <div v-show="setting.pacMode === 'gfwlist' || setting.transparent === 'gfwlist'">
      <div>{{ $t('setting.options.off') }}</div>
      <ElSelect v-model="setting.pacAutoUpdateMode" size="small">
        <ElOption value="none" :label="$t('setting.options.off')" />
        <ElOption value="auto_update" :label="$t('setting.options.updateGfwlistWhenStart')" />
        <ElOption value="auto_update_at_intervals" :label="$t('setting.options.updateGfwlistAtIntervals')" />

        <ElInput
          v-if="setting.pacAutoUpdateMode === 'auto_update_at_intervals'"
          ref="autoUpdatePacInput"
          v-model="setting.pacAutoUpdateIntervalHour"
          type="number" min="1"
        />
      </ElSelect>
    </div>

    <div>
      <div>{{ $t('setting.autoUpdateSub') }}</div>
      <ElSelect v-model="setting.subscriptionAutoUpdateMode" size="small">
        <ElOption value="none" :label="$t('setting.options.off')" />
        <ElOption value="auto_update" :label="$t('setting.options.updateSubWhenStart')" />
        <ElOption value="auto_update_at_intervals" :label="$t('setting.options.updateSubAtIntervals')" />
      </ElSelect>

      <ElInput
        v-if="setting.subscriptionAutoUpdateMode === 'auto_update_at_intervals'"
        ref="autoUpdateSubInput"
        v-model="setting.subscriptionAutoUpdateIntervalHour"
        type="number" min="1"
      />
    </div>

    <div>
      <div>{{ $t('setting.preferModeWhenUpdate') }}</div>
      <ElSelect v-model="setting.proxyModeWhenSubscribe" size="small">
        <ElOption
          value="direct"
          :label="setting.transparent === 'close' || setting.lite
            ? $t('setting.options.direct')
            : $t('setting.options.dependTransparentMode')
          "
        />
        <ElOption value="proxy" :label="$t('setting.options.global')" />
        <ElOption value="pac" :label="$t('setting.options.pac')" />
      </ElSelect>
    </div>

    <ElButton @click="updateSetting">
      {{ $t('operations.saveApply') }}
    </ElButton>
  </div>
</template>

<style scoped>
div {
  @apply flex space-x-2 my-1 justify-items-baseline;
}
</style>

<style>
.el-checkbox-button__inner {
  padding: 4px 12px;
  border-left-style: solid;
  border-left-width: 1px;
  border-left-color: rgb(229, 231, 235);
}
</style>
