<script lang="ts" setup>
const { t } = useI18n()

const columns = [
  { key: 'id', label: 'ID' },
  { key: 'name', label: t('server.name') },
  { key: 'address', label: t('server.address') },
  { key: 'net', label: t('server.protocol') },
  { key: 'pingLatency', label: t('server.latency') }
]
</script>

<template>
  <ElTable :data="proxies.servers">
    <ElTableColumn type="selection" width="55" />
    <ElTableColumn v-for="c in columns" :key="c.key" :prop="c.key" :label="c.label" />
  </ElTable>
</template>
