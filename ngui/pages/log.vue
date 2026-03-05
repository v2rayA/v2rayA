<script lang="ts" setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

definePageMeta({ middleware: ['auth'] })

type LogItem = {
  id: number
  text: string
  level: string
}

const items = $ref<LogItem[]>([])
const endOfLine = $ref(true)
const currentSkip = $ref(0)
const intervalTime = $ref(5)
const intervalCandidate = [2, 5, 10, 15]
const autoScroll = $ref(true)
const levelFilter = $ref('all')
const logContainer = ref<HTMLElement | null>(null)

let intervalId: ReturnType<typeof setInterval> | null = null

const detectLevel = (text: string) => {
  const lower = text.toLowerCase()
  if (lower.includes('[e]') || lower.includes(' error '))
    return 'error'
  if (lower.includes('[w]') || lower.includes(' warn'))
    return 'warn'
  if (lower.includes('[d]') || lower.includes(' debug'))
    return 'debug'
  if (lower.includes('[t]') || lower.includes(' trace'))
    return 'trace'
  if (lower.includes('[i]') || lower.includes(' info'))
    return 'info'
  return 'other'
}

const filteredItems = computed(() => {
  if (levelFilter === 'all')
    return items
  return items.filter(item => item.level === levelFilter)
})

const appendLogs = (payload: string) => {
  if (!payload)
    return
  const baseIndex = items.length
  const rows = payload.split('\n')
  const nextItems = rows.map((text, index) => ({
    id: baseIndex + index,
    text,
    level: detectLevel(text)
  }))
  if (endOfLine) {
    items.push(...nextItems)
  }
  else if (nextItems.length > 0) {
    items[items.length - 1].text += nextItems[0].text
    items.push(...nextItems.slice(1))
  }
  endOfLine = rows[rows.length - 1] === ''
  currentSkip += new Blob([payload]).size
  if (autoScroll && logContainer.value) {
    requestAnimationFrame(() => {
      if (logContainer.value)
        logContainer.value.scrollTop = logContainer.value.scrollHeight
    })
  }
}

const fetchLogs = async() => {
  const { data } = await useV2Fetch<string>('logger', {
    params: { skip: currentSkip }
  }).get().text()
  appendLogs(data.value || '')
}

const startPolling = () => {
  if (intervalId)
    clearInterval(intervalId)
  intervalId = setInterval(fetchLogs, intervalTime * 1000)
}

onMounted(async() => {
  await fetchLogs()
  startPolling()
})

onBeforeUnmount(() => {
  if (intervalId)
    clearInterval(intervalId)
})
</script>

<template>
  <div class="space-y-4">
    <div class="flex flex-wrap items-center gap-3">
      <div class="flex items-center gap-2">
        <span class="text-sm">{{ $t('log.category') }}</span>
        <ElSelect v-model="levelFilter" size="small">
          <ElOption value="all" :label="$t('log.categories.all')" />
          <ElOption value="error" :label="$t('log.categories.error')" />
          <ElOption value="warn" :label="$t('log.categories.warn')" />
          <ElOption value="info" :label="$t('log.categories.info')" />
          <ElOption value="debug" :label="$t('log.categories.debug')" />
          <ElOption value="trace" :label="$t('log.categories.trace')" />
          <ElOption value="other" :label="$t('log.categories.other')" />
        </ElSelect>
      </div>

      <div class="flex items-center gap-2">
        <span class="text-sm">{{ $t('log.refreshInterval') }}</span>
        <ElSelect v-model="intervalTime" size="small" @change="startPolling">
          <ElOption v-for="candidate in intervalCandidate" :key="candidate" :label="`${candidate} ${$t('log.seconds')}`" :value="candidate" />
        </ElSelect>
      </div>

      <ElCheckbox v-model="autoScroll">{{ $t('log.autoScroll') }}</ElCheckbox>
    </div>

    <div ref="logContainer" class="log-panel">
      <div v-if="filteredItems.length === 0" class="text-sm text-gray-500">
        Empty
      </div>
      <div v-for="(item, index) in filteredItems" :key="item.id" class="log-line">
        <span class="log-line-number">{{ index + 1 }}</span>
        <span class="log-line-text">{{ item.text }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.log-panel {
  max-height: 70vh;
  overflow: auto;
  border-radius: 6px;
  padding: 12px;
  background: #0f0f12;
  color: #f8f8f2;
  font-family: Consolas, Monaco, Menlo, "Courier New", monospace;
  font-size: 13px;
  line-height: 1.5;
  white-space: pre;
}

.log-line {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  white-space: pre;
}

.log-line-number {
  min-width: 3.5rem;
  text-align: right;
  color: #9aa4b2;
  user-select: none;
}

.log-line-text {
  white-space: pre;
}
</style>
