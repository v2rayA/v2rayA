export default defineNuxtRouteMiddleware(async() => {
  if (user.value.firstCheck) {
    try {
      // 使用 /api/version 接口的 hasAccounts 字段判断账户是否存在
      // 避免发送伪造的注册请求
      const { data } = await useV2Fetch<any>('version').get().json()
      user.value.exist = data.value?.data?.hasAccounts === true
    } catch {
      // API 失败时保持默认值 false，不阻塞导航
      user.value.exist = false
    }
    user.value.firstCheck = false
  }
})
