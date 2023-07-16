<script lang="ts" setup>
const { data: row, type } = defineProps<{ data: any, type: 'ping' | 'http' }>()

const testServer = async(row: any) => {
  const alias = `${type}Latency`

  const { data } = await useV2Fetch(`${alias}?whiches=${JSON.stringify(row)}`).get().json()
  for (const i of data.value.data.whiches)
    proxies.value.subs[i.sub].servers[i.id - 1][alias] = i[alias]
}
</script>

<template>
  <ElButton @click="testServer(row)">{{ type.toUpperCase() }}</ElButton>
</template>
