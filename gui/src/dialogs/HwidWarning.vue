<script setup lang="ts">
import { onMounted, useTemplateRef } from "vue";
import { useI18n } from "vue-i18n";
const safeButton = useTemplateRef<{ $el: HTMLButtonElement }>("safeButton");
onMounted(() => requestAnimationFrame(() => safeButton.value?.$el.focus()));
const emit = defineEmits<{ close: [allowed: boolean] }>();
const { t } = useI18n();
</script>

<template>
  <v-card>
    <v-card-title class="md3-title-large text-wrap">{{
      t("subscription.hwidAll")
    }}</v-card-title>
    <v-card-text>{{ t("subscription.hwidWarning") }}</v-card-text>
    <v-card-actions class="hwid-actions">
      <v-btn
        variant="text"
        size="small"
        height="auto"
        class="text-wrap my-1 px-1 md3-label-medium"
        @click="emit('close', true)"
        >{{ t("subscription.hwidAccept") }}</v-btn
      >
      <v-btn
        variant="flat"
        color="primary"
        size="large"
        min-height="56"
        ref="safeButton"
        height="auto"
        class="text-wrap align-self-stretch py-2"
        @click="emit('close', false)"
        >{{ t("subscription.hwidDecline") }}</v-btn
      >
    </v-card-actions>
  </v-card>
</template>

<style scoped>
.hwid-actions {
  display: grid;
  grid-template-columns: minmax(0, 2fr) minmax(0, 3fr);
  align-items: center;
}
.hwid-actions :deep(.v-btn__content) {
  white-space: normal;
}
</style>
