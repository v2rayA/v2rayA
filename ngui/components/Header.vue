<script lang="ts" setup>
import { system } from '~/composables/system'
import { user } from '~/composables/user'

const { t, locale } = useI18n()
const route = useRoute()

const { theme, themeSwitchLabel, toggleTheme } = useTheme()
const ws = useWebSocket()

// 语言配置（与 /gui 保持一致）
const langs = [
  { code: 'zh', label: '中文-中国', flag: 'zh' },
  { code: 'en', label: 'English-US', flag: 'en' },
  { code: 'fa', label: 'فارسی', flag: 'fa' },
  { code: 'ru', label: 'Русский', flag: 'ru' },
  { code: 'pt-br', label: 'Português-Brasil', flag: 'pt' },
]

const currentLangLabel = computed(() => {
  const lang = langs.find(l => l.code === locale.value)
  return lang ? lang.label : '中文-中国'
})

const handleClickLang = (langCode: string) => {
  locale.value = langCode
  localStorage.setItem('_lang', langCode)
  window.location.reload()
}

// 用户名（从 JWT token 解析）
const username = computed(() => {
  const token = user.value.token
  if (!token) return t('common.notLogin')
  try {
    // JWT payload 是 base64url 编码
    const payload = JSON.parse(atob(token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/')))
    return payload['uname'] || t('common.notLogin')
  } catch {
    return t('common.notLogin')
  }
})

const handleLogout = () => {
  user.value.token = ''
  user.value.exist = false
  navigateTo('/login', { replace: true })
}

const applyRunningState = (body: any) => {
  if (!body) return
  system.value.networkPaused = !!body.networkPaused
  system.value.running = !!body.running && !body.networkPaused
}

const handleRunningStateMessage = (msg: any) => {
  applyRunningState(msg.body)
}

const syncTouchState = async() => {
  try {
    const { data } = await useV2Fetch<any>('touch').json()
    if (data.value?.code === 'SUCCESS') {
      applyRunningState(data.value.data)
      if (data.value.data?.touch?.connectedServer !== undefined) {
        system.value.connect = data.value.data.touch.connectedServer
      }
    }
  } catch {
    system.value.running = false
    system.value.networkPaused = false
  }
}

watch(() => user.value.token, async(token) => {
  ws.offMessage('running_state', handleRunningStateMessage)
  ws.disconnect()
  system.value.running = false
  system.value.networkPaused = false

  if (!token)
    return

  await syncTouchState()
  ws.onMessage('running_state', handleRunningStateMessage)
  ws.connect()
}, { immediate: true })

// 当前路由高亮
const isActive = (path: string) => route.path === path
</script>

<template>
  <ElMenu
    mode="horizontal"
    class="v2raya-header"
    :ellipsis="false"
  >
    <!-- 左侧：品牌 + 状态 + 出站 -->
    <ElMenuItem class="brand-item">
      <NuxtLink to="/" class="flex items-center gap-2 no-underline">
        <span class="text-lg font-bold">V2RayA</span>
      </NuxtLink>
    </ElMenuItem>

    <ElMenuItem>
      <OperateBoot />
    </ElMenuItem>

    <ElMenuItem>
      <OperateOutbound />
    </ElMenuItem>

    <!-- 中间：导航链接 -->
    <div class="flex-1" />

    <ElMenuItem :class="{ 'is-active': isActive('/setting') }">
      <NuxtLink to="/setting">
        <span class="i-ri:settings-3-line mr-1" />
        {{ t('common.setting') }}
      </NuxtLink>
    </ElMenuItem>

    <ElMenuItem :class="{ 'is-active': isActive('/log') }">
      <NuxtLink to="/log">
        <span class="i-ri:file-list-3-line mr-1" />
        {{ t('common.log') }}
      </NuxtLink>
    </ElMenuItem>

    <ElMenuItem :class="{ 'is-active': isActive('/about') }">
      <NuxtLink to="/about">
        <span class="i-ri:heart-3-line mr-1" />
        {{ t('common.about') }}
      </NuxtLink>
    </ElMenuItem>

    <!-- 主题切换 -->
    <ElSubMenu index="theme">
      <template #title>
        <span :class="theme === 'auto' ? 'i-ri:contrast-2-line' : (theme === 'dark' ? 'i-ri:moon-line' : 'i-ri:sun-line')" />
        <span class="ml-1">{{ themeSwitchLabel }}</span>
      </template>
      <ElMenuItem @click="theme = 'auto'">
        <span class="i-ri:contrast-2-line mr-2" />自动
      </ElMenuItem>
      <ElMenuItem @click="theme = 'light'">
        <span class="i-ri:sun-line mr-2" />浅色
      </ElMenuItem>
      <ElMenuItem @click="theme = 'dark'">
        <span class="i-ri:moon-line mr-2" />深色
      </ElMenuItem>
    </ElSubMenu>

    <!-- 语言切换 -->
    <ElSubMenu index="language">
      <template #title>
        <span class="i-ri:earth-line" />
        <span class="ml-1">{{ currentLangLabel }}</span>
      </template>
      <ElMenuItem
        v-for="lang in langs"
        :key="lang.code"
        @click="handleClickLang(lang.code)"
      >
        <span class="font-medium min-w-30 inline-block">{{ lang.label }}</span>
        <span class="ml-2 text-gray-400">{{ lang.code }}</span>
      </ElMenuItem>
    </ElSubMenu>

    <!-- 用户菜单 -->
    <ElSubMenu v-if="user.token" index="user">
      <template #title>
        <span class="i-ri:user-3-line" />
        <span class="ml-1">{{ username }}</span>
      </template>
      <ElMenuItem disabled>
        <span v-html="t('common.loggedAs', { username })" />
      </ElMenuItem>
      <ElMenuItem divided @click="handleLogout">
        <span class="i-ri:logout-box-r-line mr-2" />
        {{ t('operations.logout') }}
      </ElMenuItem>
    </ElSubMenu>
  </ElMenu>
</template>

<style scoped>
.v2raya-header {
  padding: 0 16px;
  border-bottom: 1px solid var(--el-border-color-light);
}

.brand-item {
  padding: 0 8px !important;
}

.v2raya-header :deep(.el-menu-item),
.v2raya-header :deep(.el-sub-menu__title) {
  height: 48px;
  line-height: 48px;
}

.v2raya-header :deep(a) {
  text-decoration: none;
  color: inherit;
}
</style>
