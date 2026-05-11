<script lang="ts" setup>
import { ElMessage } from 'element-plus'
import { ref } from 'vue'
import { user } from '~/composables/user'

// If no account exists, redirect to signup page.
if (!user.value.exist) {
  await navigateTo('/signup', { replace: true })
}

const { t } = useI18n()

const username = ref('')
const password = ref('')
const isSubmitting = ref(false)

async function login() {
  if (isSubmitting.value)
    return

  isSubmitting.value = true

  const { data } = await useV2Fetch<any>('login').post({ username: username.value, password: password.value }).json()

  if (data.value?.data?.first === true) {
    user.value.exist = false
    await navigateTo('/signup')
    isSubmitting.value = false
    return
  }

  if (data.value.code !== 'SUCCESS') {
    ElMessage.warning({ message: data.value.message, duration: 5000 })
  } else {
    user.value.token = data.value.data.token
    ElMessage.success(t('common.success'))
    await navigateTo('/')
  }

  isSubmitting.value = false
}
</script>

<template>
  <div class="mx-auto w-96">
    <h1 class="text-2xl mb-6">{{ `${t('login.title')} - v2rayA` }}</h1>

    <ElForm label-width="auto">
      <ElFormItem :label="t('login.username')">
        <ElInput v-model="username" autofocus />
      </ElFormItem>

      <ElFormItem :label="t('login.password')">
        <ElInput v-model="password" type="password" show-password />
      </ElFormItem>

      <ElFormItem>
        <ElButton type="primary" class="flex mx-auto" :loading="isSubmitting" :disabled="isSubmitting || username === '' || password === ''" @click="login">
          {{ t("operations.login") }}
        </ElButton>
      </ElFormItem>

      <ElAlert type="info" show-icon :closable="false">
        If you forget your password, you can reset it by exec <code>v2raya --reset-password</code> and restarting v2rayA.

        <a @click="user.exist = false">Already reset password</a>
      </ElAlert>
    </ElForm>
  </div>
</template>

<style>
.va-input-wrapper--labeled .va-input-wrapper__label {
  height: 14px;
}
</style>
