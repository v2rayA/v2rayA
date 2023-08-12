export default defineNuxtRouteMiddleware(async() => {
  if (!user.value.token) {
    if (user.value.exist)
      return navigateTo('/login')
    else
      return navigateTo('/signup')
  }
})
