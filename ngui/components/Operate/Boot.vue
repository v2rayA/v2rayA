<script lang="ts" setup>
import { system } from '~/composables/system'

const { t } = useI18n()

const bootV2rayA = async() => {
  const { data } = await useV2Fetch('v2ray').post().json()
  if (data.value.code === 'SUCCESS') {
    system.value.running = !!data.value.data?.running && !data.value.data?.networkPaused
    system.value.networkPaused = !!data.value.data?.networkPaused
    system.value.connect = data.value.data.touch.connectedServer
  }
}

const stopV2rayA = async() => {
  const { data } = await useV2Fetch('v2ray').delete().json()
  if (data.value.code === 'SUCCESS') {
    system.value.running = false
    system.value.networkPaused = !!data.value.data?.networkPaused
    system.value.connect = data.value.data.touch.connectedServer
  }
}

const handleClick = () => {
  if (system.value.networkPaused)
    return

  if (system.value.running) {
    stopV2rayA()
  } else {
    bootV2rayA()
  }
}

const statusLabel = computed(() => {
  if (system.value.networkPaused)
    return t('common.waitingNetwork')

  return system.value.running ? t('common.isRunning') : t('common.notRunning')
})
</script>

<template>
  <ElButton
    size="small"
    :type="system.networkPaused ? 'info' : (system.running ? 'warning' : 'primary')"
    :disabled="system.networkPaused"
    @click="handleClick"
  >
    {{ statusLabel }}
  </ElButton>
</template>
