<script lang="ts" setup>
const { t } = useI18n()

const { data, id } = defineProps<{ data: any[], id: number }>()

const emit = defineEmits<{
  refresh: []
}>()

const columns = [
  { key: 'id', label: 'ID', width: 70 },
  { key: 'name', label: t('server.name') },
  { key: 'address', label: t('server.address') },
  { key: 'net', label: t('server.protocol') },
  { key: 'pingLatency', label: t('server.latency') }
]

let selectRows = $ref<any[]>([])
const handleSelectionChange = (val: any) => { selectRows = val }

// ---- ServerEditor integration ----
let showServerEditor = $ref(false)
let editingServer = $ref<any>(null)

const openCreateServer = () => {
  editingServer = null
  showServerEditor = true
}

const openEditServer = (row: any) => {
  editingServer = { id: row.id, _type: row._type, sub: id }
  showServerEditor = true
}

const onServerSaved = () => {
  emit('refresh')
}
</script>

<template>
  <OperateLatency :data="selectRows" type="ping" />
  <OperateLatency :data="selectRows" type="http" />

  <div class="flex items-center gap-2 mb-2">
    <ElButton type="primary" size="small" @click="openCreateServer">
      {{ t('operations.add') || '创建服务器' }}
    </ElButton>
  </div>

  <ElTable :data="data" @selection-change="handleSelectionChange">
    <ElTableColumn type="selection" width="55" />
    <ElTableColumn v-for="c in columns" :key="c.key" :prop="c.key" :label="c.label" :width="c.width" />
    <ElTableColumn :label="t('operations.name')" min-width="300">
      <template #default="scope">
        <OperateConnect :data="scope.row" :sub-i-d="id" />
        <OperateView :data="scope.row" :sub-i-d="id" />
        <OperateShare :data="scope" :sub-i-d="id" />
        <ElButton size="small" text type="primary" @click="openEditServer(scope.row)">
          编辑
        </ElButton>
      </template>
    </ElTableColumn>
  </ElTable>

  <ServerEditor
    v-model="showServerEditor"
    :server="editingServer"
    @save="onServerSaved"
  />
</template>
