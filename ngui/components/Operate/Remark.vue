<script lang="ts" setup>
const { data: row } = defineProps<{ data: any }>()

const input = $ref('')
let isVisible = $ref(false)

const remarkSubscription = async(remark: string) => {
  const { data } = await useV2Fetch('subscription').patch({
    subscription: {
      ...row,
      remarks: remark
    }
  }).json()

  proxies.value.subs = data.value.data.touch.subscriptions
  isVisible = false
}
</script>

<template>
  <ElButton size="small" class="mr-3" @click="isVisible = true">
    <UnoIcon class="ri:edit-2-line mr-1" />{{ $t('operations.modify') }}
  </ElButton>

  <ElDialog v-model="isVisible" :title="$t('operations.import')">
    {{ $t("configureSubscription.title") }}
    <ElInput v-model="input" :placeholder="$t('subscription.remarks')" />
    <template #footer>
      <span class="dialog-footer">
        <ElButton @click="isVisible = false">{{ $t('operations.cancel') }}</ElButton>
        <ElButton type="primary" @click="remarkSubscription(input)">
          {{ $t("operations.confirm") }}
        </ElButton>
      </span>
    </template>
  </ElDialog>
</template>
