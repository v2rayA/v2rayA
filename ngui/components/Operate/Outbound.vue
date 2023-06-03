<script lang="ts" setup>
const { t } = useI18n()

let isVisible = $ref(false)
let currentOutbound = $ref('')
let setting = $ref<{
  probeInterval: string
  probeURL: string
  type: string
}>()

const viewOutbound = async(outbound: string) => {
  isVisible = true
  currentOutbound = outbound
  const { data } = await useV2Fetch(`outbound?outbound=${outbound}`).json()

  setting = data.value.data.setting
}

const deleteOutbound = async(outbound: string) => {
  const { data } = await useV2Fetch('outbound').delete({ outbound }).json()
  proxies.value.outbounds = data.value.data.outbounds
  isVisible = false
}

const editOutbound = async(outbound: string) => {
  const { data } = await useV2Fetch('outbound').put({ outbound, setting }).json()
  if (data.value.code === 'SUCCESS')
    ElMessage.success(t('common.success'))
}
</script>

<template>
  <ElDropdown class="ml-2">
    <ElButton size="small">{{ proxies.currentOutbound.toUpperCase() }}</ElButton>
    <template #dropdown>
      <ElDropdownMenu class="w-28">
        <ElDropdownItem v-for="i in proxies.outbounds" :key="i" class="flex justify-between">
          <div class="w-full" @click="proxies.currentOutbound = i">{{ i }}</div>
          <UnoIcon class="ri:settings-fill ml-2" @click="viewOutbound(i)" />
        </ElDropdownItem>
      </ElDropdownMenu>
    </template>
  </ElDropdown>

  <ElDialog v-model="isVisible" :title="`${currentOutbound}- ${$t('common.outboundSetting')}`">
    <ElForm>
      <ElFormItem v-for="(v, k) in setting" :key="k" v-model="setting" :label="k.toString()">
        <ElInput v-model="setting![k]" />
      </ElFormItem>
    </ElForm>
    <template #footer>
      <span class="dialog-footer">
        <ElButton @click="deleteOutbound(currentOutbound)">{{ $t('operations.delete') }}</ElButton>
        <ElButton @click="isVisible = false">{{ $t('operations.cancel') }}</ElButton>
        <ElButton type="primary" @click="editOutbound(currentOutbound)">
          {{ $t("operations.confirm") }}
        </ElButton>
      </span>
    </template>
  </ElDialog>
</template>
