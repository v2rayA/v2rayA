<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { openDialog } from "@/composables/useDialog";
import HwidWarning from "./HwidWarning.vue";

const model = defineModel<boolean>({ required: true });
const { t } = useI18n();
async function change(event: MouseEvent) {
  if (!(event.target instanceof HTMLInputElement)) return;
  // Defer the native checkbox change until the user's decision is known.
  event.preventDefault();
  const enabled = !model.value;
  if (
    !enabled ||
    (await openDialog<boolean>(HwidWarning, {}, { width: 480 }).result)
  )
    model.value = !!enabled;
}
</script>

<template>
  <v-switch
    :model-value="model"
    :label="t('subscription.hwidAll')"
    hide-details
    @click="change"
  />
</template>
