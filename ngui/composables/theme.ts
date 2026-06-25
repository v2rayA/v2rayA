// 主题管理 composable，与 /gui 的 theme 逻辑保持一致
// 支持 auto / light / dark 三种模式，通过 localStorage 持久化

export type ThemeMode = 'auto' | 'light' | 'dark'

export const useTheme = () => {
  const theme = useLocalStorage<ThemeMode>('theme', 'auto')

  const isDarkTheme = computed(() => {
    if (theme.value === 'dark') return true
    if (theme.value === 'light') return false
    // auto: 跟随系统
    return window.matchMedia('(prefers-color-scheme: dark)').matches
  })

  const themeSwitchLabel = computed(() => {
    if (theme.value === 'auto') return '自动'
    if (theme.value === 'dark') return '深色'
    return '浅色'
  })

  const applyTheme = () => {
    document.documentElement.classList.toggle('dark', isDarkTheme.value)
  }

  const toggleTheme = () => {
    const order: ThemeMode[] = ['auto', 'light', 'dark']
    const idx = order.indexOf(theme.value)
    theme.value = order[(idx + 1) % order.length]
    applyTheme()
  }

  // 监听系统主题变化
  let mediaQuery: MediaQueryList | null = null
  let onSystemChange: (() => void) | null = null

  onMounted(() => {
    applyTheme()
    mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    onSystemChange = () => {
      if (theme.value === 'auto') applyTheme()
    }
    mediaQuery.addEventListener('change', onSystemChange)
  })

  onUnmounted(() => {
    if (mediaQuery && onSystemChange) {
      mediaQuery.removeEventListener('change', onSystemChange)
    }
  })

  watch(theme, applyTheme)

  return {
    theme,
    isDarkTheme,
    themeSwitchLabel,
    toggleTheme,
    applyTheme
  }
}
