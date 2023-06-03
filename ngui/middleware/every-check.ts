import type { MessageParams } from 'element-plus'

export default defineNuxtRouteMiddleware(async() => {
  const nuxtApp = useNuxtApp()
  const { t } = nuxtApp.$i18n

  const { data } = await useV2Fetch<any>('version').json()

  if (data.value.code === 'SUCCESS') {
    system.value.docker = data.value.data.dockerMode
    system.value.version = data.value.data.version
    system.value.lite = data.value.data.lite

    let messageConf: MessageParams = {
      message: t(system.value.docker ? 'welcome.docker' : 'welcome.default', {
        version: system.value.version
      }),
      duration: 3000
    }

    if (data.value.data.foundNew) {
      messageConf = {
        duration: 5000,
        type: 'success',
        message: `${messageConf.message}. ${t('welcome.newVersion', {
          version: data.value.data.remoteVersion
        })}`
      }
    }

    ElMessage(messageConf)

    if (data.value.data.serviceValid === false)
      ElMessage.error({ message: t('version.v2rayInvalid'), duration: 10000 })
    else if (!data.value.data.v5)
      ElMessage.error({ message: t('version.v2rayNotV5'), duration: 10000 })
  }
})
