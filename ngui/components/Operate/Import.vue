<script lang="ts" setup>
const { t } = useI18n()

const input = $ref('')

let isVisible = $ref(false)
let isImporting = $ref(false)
const ifMulite = $ref(false)

const handleClickImportConfirm = async() => {
  isImporting = true
  const { data } = await useV2Fetch('import').post({ url: input }).json()
  isImporting = false

  if (data.value.code === 'SUCCESS') {
    proxies.value.subs = data.value.data.touch.subscriptions
    ElMessage.success(t('common.success'))
    isVisible = false
  }
}
</script>

<template>
  <ElButton @click="isVisible = true">
    {{ t("operations.import") }}
  </ElButton>

  <ElDialog v-model="isVisible" :title="$t('operations.import')">
    {{ $t('import.message') }}
    <ElInput v-model="input" :type="ifMulite ? 'textarea' : 'url'" />
    <template #footer>
      <span class="dialog-footer">
        <ElButton @click="ifMulite = true">{{ $t('operations.inBatch') }}</ElButton>
        <ElButton @click="isVisible = false">{{ $t('operations.cancel') }}</ElButton>
        <ElButton type="primary" :loading="isImporting" @click="handleClickImportConfirm">
          {{ $t("operations.confirm") }}
        </ElButton>
      </span>
    </template>
  </ElDialog>
</template>
