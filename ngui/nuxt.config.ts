// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  ssr: false,

  // Output directory control: OUTPUT_DIR env var overrides default
  // Keeps the same convention as gui/vite.config.js
  nitro: {
    output: {
      publicDir: process.env.OUTPUT_DIR || '../service/server/router/web',
    },
    prerender: {
      failOnError: false,
    },
  },

  // Resource path: Nuxt uses /_nuxt/ by default (not /static/)
  app: {
    baseURL: '/',
    buildAssetsDir: '/_nuxt',
  },

  // Dev proxy for API calls during development
  devServer: {
    port: 3000,
  },
  devProxy: {
    '/api': {
      target: 'http://127.0.0.1:2017',
      changeOrigin: true,
    },
  },

  // Pre-generate known routes for SPA fallback
  generate: {
    routes: ['/', '/login', '/signup', '/setting', '/log', '/about'],
  },

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
      },
      {
        code: 'en',
        iso: 'en-US',
        file: 'en.yaml',
        name: 'English-US'
      },
      {
        code: 'fa',
        iso: 'fa-IR',
        file: 'fa.yaml',
        name: 'فارسی'
      },
      {
        code: 'ru',
        iso: 'ru-RU',
        file: 'ru.yaml',
        name: 'Русский'
      },
      {
        code: 'pt-br',
        iso: 'pt-BR',
        file: 'pt-br.yaml',
        name: 'Português-Brasil'
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
