<script lang="ts" setup>
definePageMeta({ middleware: ['auth', 'every-check'] })

const currentTab = ref('SUBSCRIPTION')

const { data: { value: { data: { outbounds } } } } = await useV2Fetch<any>('outbounds').json()
const { data: { value: { data: { touch } } } } = await useV2Fetch<any>('touch').json()

proxies.value = {
  ...proxies.value,
  outbounds,
  subs: touch.subscriptions,
  servers: touch.servers
}
</script>

<template>
  <ElTabs v-model="currentTab" class="mb-10" tab-position="left">
    <ElTabPane name="SUBSCRIPTION" label="SUBSCRIPTION">
      <Subscription />
    </ElTabPane>
    <ElTabPane name="SERVER" label="SERVER">
      <Server />
    </ElTabPane>
    <ElTabPane
      v-for="s in proxies.subs"
      :key="`${s.host}:${s.id}`"
      :name="s.host"
      :label="s.remarks || s.host.toUpperCase()"
    >
      <SubServer :id="s.id" :data="s.servers" />
    </ElTabPane>
  </ElTabs>
</template>
