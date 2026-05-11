<script lang="ts" setup>
definePageMeta({ middleware: ['auth'] })

const { t } = useI18n()

let versionInfo = $ref<any>({})
let loading = $ref(true)

onMounted(async () => {
  try {
    const { data } = await useV2Fetch<any>('version').json()
    if (data.value?.code === 'SUCCESS') {
      versionInfo = data.value.data
    }
  } catch (e) {
    console.error('Failed to load version info', e)
  } finally {
    loading = false
  }
})
</script>

<template>
  <div class="mx-auto max-w-2xl px-4 py-6 space-y-6">
    <div class="text-center space-y-2">
      <h1 class="text-2xl font-bold">mzz2017 / v2rayA</h1>
      <div class="flex justify-center space-x-2">
        <a href="https://github.com/v2rayA/v2rayA" target="_blank" class="flex space-x-2">
          <img src="https://img.shields.io/github/stars/mzz2017/v2rayA.svg?style=social" alt="stars">
          <img src="https://img.shields.io/github/forks/mzz2017/v2rayA.svg?style=social" alt="forks">
          <img src="https://img.shields.io/github/watchers/mzz2017/v2rayA.svg?style=social" alt="watchers">
        </a>
      </div>
    </div>

    <!-- Version Info -->
    <ElCard v-if="!loading">
      <template #header>
        <span class="font-semibold">{{ $t('common.v2rayCoreStatus') }}</span>
      </template>
      <div class="space-y-2 text-sm">
        <div class="flex justify-between">
          <span>v2rayA</span>
          <span class="font-mono">{{ versionInfo.version || '-' }}</span>
        </div>
        <div class="flex justify-between">
          <span>v2ray-core / xray-core</span>
          <span class="font-mono">{{ versionInfo.coreVersion || '-' }}</span>
        </div>
        <div class="flex justify-between">
          <span>{{ $t('common.latest') }}</span>
          <span class="font-mono">{{ versionInfo.remoteVersion || '-' }}</span>
        </div>
        <div class="flex justify-between">
          <span>Docker</span>
          <span>{{ versionInfo.dockerMode ? $t('common.yes') : $t('common.no') }}</span>
        </div>
        <div class="flex justify-between">
          <span>OS</span>
          <span class="font-mono">{{ versionInfo.os || '-' }}</span>
        </div>
        <div class="flex justify-between">
          <span>Lite</span>
          <span>{{ versionInfo.lite ? $t('common.yes') : $t('common.no') }}</span>
        </div>
      </div>
    </ElCard>

    <!-- About Content -->
    <ElCard>
      <div class="prose max-w-none" v-html="$t('about')" />
    </ElCard>

    <!-- Core Status Warnings -->
    <ElAlert v-if="versionInfo.coreVersionValid === false" type="warning" show-icon :closable="false">
      <template #title>
        {{ $t('version.coreVersionMismatch', { err: versionInfo.coreVersionErr || '' }) }}
      </template>
    </ElAlert>
    <ElAlert v-if="versionInfo.serviceValid === false" type="warning" show-icon :closable="false">
      <template #title>
        {{ $t('version.v2rayInvalid') }}
      </template>
    </ElAlert>
    <ElAlert v-if="versionInfo.v5 === false" type="warning" show-icon :closable="false">
      <template #title>
        {{ $t('version.v2rayNotV5') }}
      </template>
    </ElAlert>
  </div>
</template>

<style scoped>
.prose :deep(.about-small) {
  font-size: 0.85em;
}
</style>
