<script lang="ts" setup>
const { t } = useI18n()

const columns = [
  { key: 'id', label: 'ID', width: 70 },
  { key: 'host', label: t('subscription.host'), width: 220 },
  { key: 'remarks', label: t('subscription.remarks'), width: 120 },
  { key: 'status', label: t('subscription.timeLastUpdate'), width: 180 }
]

let selectRows = $ref<any[]>([])
const handleSelectionChange = (val: any) => { selectRows = val }

let isUpdating = $ref(false)
const updateSubscription = async(row: any) => {
  isUpdating = true
  const { data } = await useV2Fetch('subscription').put({ id: row.id, _type: row._type }).json()
  isUpdating = false

  if (data.value.code === 'SUCCESS')
    ElMessage.success({ message: t('common.success'), duration: 5000 })
}

const removeSubscription = async(row: any) => {
  const { data } = await useV2Fetch('touch').delete({
    touches: selectRows.map((x) => { return { id: x.id, _type: x._type } })
  }).json()

  if (data.value.code === 'SUCCESS')
    proxies.value.subs = data.value.data.touch.subscriptions
}
</script>

<template>
  <OperateImport />

  <ElButton
    v-if="!(Array.isArray(selectRows) && selectRows.length === 0)"
    class="ml-2"
    @click="removeSubscription(selectRows)"
  >
    {{ t('operations.delete') }}
  </ElButton>

  <ElTable :data="proxies.subs" @selection-change="handleSelectionChange">
    <ElTableColumn type="selection" width="55" />
    <ElTableColumn v-for="c in columns" :key="c.key" :prop="c.key" :label="c.label" :min-width="c.width" />
    <ElTableColumn property="servers.length" :label="t('subscription.numberServers')" min-width="70" />
    <ElTableColumn :label="t('operations.name')" min-width="240">
      <template #default="scope">
        <ElButton size="small" :loading="isUpdating" @click="updateSubscription(scope.row)">
          <UnoIcon v-if="!isUpdating" class="ri:refresh-line mr-1" /> {{ t('operations.update') }}
        </ElButton>
        <OperateRemark :data="scope.row" />
        <OperateShare :data="scope" />
      </template>
    </ElTableColumn>
  </ElTable>
</template>
