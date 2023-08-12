export default defineNuxtRouteMiddleware(async() => {
  if (user.value.firstCheck) {
    // TODO: change backend api
    const { data } = await useV2Fetch<any>('account').post({ username: '', password: 'aaaaaa' }).json()
    user.value.firstCheck = false
    user.value.exist = data.value?.message === 'register closed'
  }
})
