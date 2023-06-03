<script lang="ts" setup>
const { t } = useI18n()

const { data, id } = defineProps<{ data: any[], id: number }>()

const columns = [
  { key: 'id', label: 'ID', width: 70 },
  { key: 'name', label: t('server.name') },
  { key: 'address', label: t('server.address') },
  { key: 'net', label: t('server.protocol') },
  { key: 'pingLatency', label: t('server.latency') }
]

let selectRows = $ref<any[]>([])
const handleSelectionChange = (val: any) => { selectRows = val }
</script>

<template>
  <OperateLatency :data="selectRows" type="ping" />
  <OperateLatency :data="selectRows" type="http" />

  <ElTable :data="data" @selection-change="handleSelectionChange">
    <ElTableColumn type="selection" width="55" />
    <ElTableColumn v-for="c in columns" :key="c.key" :prop="c.key" :label="c.label" :width="c.width" />
    <ElTableColumn :label="t('operations.name')" min-width="240">
      <template #default="scope">
        <OperateConnect :data="scope.row" :sub-i-d="id" />
        <OperateView :data="scope.row" :sub-i-d="id" />
        <OperateShare :data="scope" :sub-i-d="id" />
      </template>
    </ElTableColumn>
  </ElTable>
</template>
