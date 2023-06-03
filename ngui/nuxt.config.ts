// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  ssr: false,
  modules: [
    '@vueuse/nuxt',
    '@unocss/nuxt',
    '@nuxtjs/i18n',
    '@element-plus/nuxt'
  ],
  i18n: {
    strategy: 'no_prefix',
    langDir: 'locales',
    locales: [
      {
        code: 'zh',
        iso: 'zh-hans',
        file: 'zh-hans.yaml',
        name: '简体中文'
      }
    ]
  },
  unocss: {
    preflight: true
  },
  experimental: {
    reactivityTransform: true
  }
})
