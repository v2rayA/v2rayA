<script lang="ts" setup>
if (user.value.exist) navigateTo('/login')

const { t } = useI18n()

const username = $ref('')
const password = $ref('')

async function signup() {
  const { data } = await useV2Fetch<any>('account').post({ username, password }).json()

  if (data.value.code !== 'SUCCESS') {
    ElMessage.warning({ message: data.value.message, duration: 5000 })
  } else {
    user.value.exist = true
    user.value.token = data.value.data.token
    ElMessage.success(t('common.success'))
    navigateTo('/')
  }
}
</script>

<template>
  <div class="mx-auto w-96">
    <h1 class="text-2xl mb-6">{{ t('register.title') }}</h1>

    <ElForm label-width="auto">
      <ElFormItem :label="t('login.username')">
        <ElInput v-model="username" autofocus />
      </ElFormItem>

      <ElFormItem :label="t('login.password')">
        <ElInput v-model="password" type="password" max-length="36" show-password />
      </ElFormItem>

      <ElFormItem>
        <ElButton type="primary" class="flex mx-auto" @click="signup">{{ t("operations.create") }}</ElButton>
      </ElFormItem>
    </ElForm>

    <div class="mt-4 bg-gray-200 p-4 rounded-sm" />

    <ElAlert type="info" show-icon :closable="false">
      <p>{{ t("register.messages.0") }}</p>
      <p>{{ t("register.messages.1") }}</p>
      <p>{{ t("register.messages.2") }}</p>
    </ElAlert>
  </div>
</template>

<style>
.va-input-wrapper--labeled .va-input-wrapper__label {
  height: 14px;
}
</style>
