<script lang="ts" setup>
const bootV2rayA = async() => {
  const { data } = await useV2Fetch('v2ray').post().json()
  if (data.value.code === 'SUCCESS') {
    system.value.running = true
    system.value.connect = data.value.data.touch.connectedServer
  }
}
</script>

<template>
  <ElButton size="small" @click="bootV2rayA">
    {{ system.running ? $t('common.isRunning') : $t('common.notRunning') }}
  </ElButton>
</template>
