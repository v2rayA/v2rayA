<script lang="ts" setup>
const { data: row, subID } = defineProps<{ data: any, subID?: number }>()

const { t } = useI18n()

const connectServer = async() => {
  const { data } = await useV2Fetch('connection').post({
    sub: subID! - 1,
    id: row.id,
    _type: row._type,
    outbound: proxies.value.currentOutbound
  }).json()

  if (data.value.code === 'SUCCESS')
    ElMessage.success(t('common.success'))
}
</script>

<template>
  <ElButton size="small" @click="connectServer">
    <UnoIcon class="ri:links-fill mr-1" />
    {{ row.connected ? $t("operations.cancel") : $t("operations.select") }}
  </ElButton>
</template>
