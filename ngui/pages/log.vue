<script lang="ts" setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

definePageMeta({ middleware: ['auth'] })

type LogItem = {
  id: number
  text: string
  level: string
  source: string
}

const items = $ref<LogItem[]>([])
const endOfLine = $ref(true)
const currentSkip = $ref(0)
const intervalTime = $ref(5)
const intervalCandidate = [2, 5, 10, 15]
const autoScroll = $ref(true)
const autoShowNew = $ref(true)
const levelFilter = $ref('all')
const sourceFilter = $ref('all')
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

const detectSource = (text: string): string => {
  // Try to extract source from log format like [source] message
  const match = text.match(/^\[(\w+)\]/)
  if (match) return match[1].toLowerCase()
  return 'other'
}

const filteredItems = computed(() => {
  let result = items
  if (levelFilter !== 'all') {
    result = result.filter(item => item.level === levelFilter)
  }
  if (sourceFilter !== 'all') {
    result = result.filter(item => item.source === sourceFilter)
  }
  return result
})

// Collect available sources from items
const availableSources = computed(() => {
  const sources = new Set<string>()
  sources.add('all')
  items.forEach(item => {
    if (item.source) sources.add(item.source)
  })
  return Array.from(sources)
})

const appendLogs = (payload: string) => {
  if (!payload)
    return
  const baseIndex = items.length
  const rows = payload.split('\n')
  const nextItems = rows.map((text, index) => ({
    id: baseIndex + index,
    text,
    level: detectLevel(text),
    source: detectSource(text)
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
      <!-- Level Filter -->
      <div class="flex items-center gap-2">
        <span class="text-sm">{{ $t('log.category') }}</span>
        <ElSelect v-model="levelFilter" size="small" style="width:100px">
          <ElOption value="all" :label="$t('log.categories.all')" />
          <ElOption value="error" :label="$t('log.categories.error')" />
          <ElOption value="warn" :label="$t('log.categories.warn')" />
          <ElOption value="info" :label="$t('log.categories.info')" />
          <ElOption value="debug" :label="$t('log.categories.debug')" />
          <ElOption value="trace" :label="$t('log.categories.trace')" />
          <ElOption value="other" :label="$t('log.categories.other')" />
        </ElSelect>
      </div>

      <!-- Source Filter -->
      <div class="flex items-center gap-2">
        <span class="text-sm">{{ $t('log.source') }}</span>
        <ElSelect v-model="sourceFilter" size="small" style="width:100px">
          <ElOption value="all" :label="$t('log.sources.all')" />
          <ElOption v-for="src in availableSources.filter(s => s !== 'all')" :key="src" :value="src" :label="src" />
        </ElSelect>
      </div>

      <!-- Refresh Interval -->
      <div class="flex items-center gap-2">
        <span class="text-sm">{{ $t('log.refreshInterval') }}</span>
        <ElSelect v-model="intervalTime" size="small" style="width:120px" @change="startPolling">
          <ElOption v-for="candidate in intervalCandidate" :key="candidate" :label="`${candidate} ${$t('log.seconds')}`" :value="candidate" />
        </ElSelect>
      </div>

      <!-- Auto Scroll -->
      <ElCheckbox v-model="autoScroll">{{ $t('log.autoScroll') }}</ElCheckbox>

      <!-- Auto Show New -->
      <ElCheckbox v-model="autoShowNew">{{ $t('log.autoShowNew') }}</ElCheckbox>
    </div>

    <!-- Log Panel -->
    <div ref="logContainer" class="log-panel">
      <div v-if="filteredItems.length === 0" class="text-sm text-gray-500">
        Empty
      </div>
      <div v-for="(item, index) in filteredItems" :key="item.id" class="log-line" :class="`log-level-${item.level}`">
        <span class="log-line-number">{{ index + 1 }}</span>
        <span class="log-line-source">{{ item.source }}</span>
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
  gap: 8px;
  white-space: pre;
}

.log-line-number {
  min-width: 3.5rem;
  text-align: right;
  color: #9aa4b2;
  user-select: none;
}

.log-line-source {
  min-width: 5rem;
  color: #66d9ef;
  user-select: none;
  font-size: 0.9em;
}

.log-line-text {
  white-space: pre;
  flex: 1;
}

.log-level-error .log-line-text {
  color: #f92672;
}

.log-level-warn .log-line-text {
  color: #e6db74;
}

.log-level-info .log-line-text {
  color: #a6e22e;
}

.log-level-debug .log-line-text {
  color: #66d9ef;
}

.log-level-trace .log-line-text {
  color: #75715e;
}
</style>
