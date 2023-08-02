<script lang="ts" setup>
import { useQRCode } from '@vueuse/integrations/useQRCode'
const { data, subID } = defineProps<{ data: any, subID?: number }>()

const { $index, row } = data

let isVisible = $ref(false)
let qrcode = ref('')

const shareSubscription = async() => {
  const params = JSON.stringify({
    id: row.id,
    _type: row._type,
    sub: row._type === 'subscription' ? $index : subID! - 1
  })

  const { data } = await useV2Fetch(`sharingAddress?touch=${params}`).get().json()

  qrcode = useQRCode(data.value.data.sharingAddress)
  isVisible = true
}
</script>

<template>
  <ElButton size="small" @click="shareSubscription">
    <UnoIcon class="ri:share-fill mr-1" /> {{ $t('operations.share') }}
  </ElButton>
  <ElDialog v-model="isVisible" :title="$t('operations.import')" width="250">
    <template #header="{ titleClass }">
      <div :class="titleClass">{{ row._type.toUpperCase() }}</div>
    </template>

    <img :src="qrcode">
  </ElDialog>
</template>
