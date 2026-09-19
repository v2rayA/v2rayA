<script setup lang="ts">
import { mdiChevronLeft, mdiChevronRight, mdiHelpCircleOutline } from "@mdi/js";
import { useI18n } from "vue-i18n";
import { useRtl } from "vuetify";

const { isRtl } = useRtl();
const { t } = useI18n();
defineProps<{
  title: string;
  hint?: string;
  subtitle?: string;
  action?: boolean;
}>();
</script>

<template>
  <v-list-item class="setting-row" :title="title" lines="two">
    <template v-if="hint || subtitle" #subtitle>
      <span class="setting-row__hint">{{ subtitle || hint }}</span>
    </template>
    <template #append>
      <v-tooltip v-if="hint" :text="hint" max-width="360" open-on-click>
        <template #activator="{ props: tip }">
          <v-btn
            v-bind="tip"
            :icon="mdiHelpCircleOutline"
            variant="text"
            :aria-label="`${t('operations.helpManual')}: ${title}`"
            size="48"
            color="on-surface-variant"
            @click.stop
            @keydown.stop
          />
        </template>
      </v-tooltip>
      <slot />
      <v-icon
        v-if="action"
        :icon="isRtl ? mdiChevronLeft : mdiChevronRight"
        color="on-surface-variant"
      />
    </template>
  </v-list-item>
</template>

<style scoped>
.setting-row :deep(.v-list-item-title) {
  white-space: normal;
}
.setting-row__hint {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.setting-row :deep(.v-list-item__append) {
  margin-inline-start: 8px;
}
</style>
