<script setup lang="ts">
import { computed } from "vue";
import { mdiChevronLeft, mdiChevronRight } from "@mdi/js";
import { useRtl } from "vuetify";
import SettingRow from "./SettingRow.vue";

const model = defineModel<string>({ required: true });
const props = defineProps<{
  title: string;
  hint?: string;
  items: { value: string; title: string; props?: { disabled: boolean } }[];
}>();
const { isRtl } = useRtl();
const current = computed(
  () =>
    props.items.find((item) => item.value === model.value)?.title ??
    model.value,
);
</script>

<template>
  <SettingRow :title="title" :hint="hint">
    <v-menu max-width="360">
      <template #activator="{ props: menu }">
        <v-btn
          v-bind="menu"
          class="setting-choice"
          variant="text"
          color="on-surface-variant"
          :aria-label="`${title}: ${current}`"
          :append-icon="isRtl ? mdiChevronLeft : mdiChevronRight"
        >
          <span class="setting-choice__value">{{ current }}</span>
        </v-btn>
      </template>
      <v-list :aria-label="title" role="menu">
        <v-list-item
          v-for="item in items"
          :key="item.value"
          :title="item.title"
          :disabled="item.props?.disabled"
          :active="model === item.value"
          :aria-checked="model === item.value"
          role="menuitemradio"
          @click="model = item.value"
        />
      </v-list>
    </v-menu>
  </SettingRow>
</template>

<style scoped>
.setting-choice {
  max-width: 240px;
  height: auto;
  min-height: 48px;
}
.setting-choice__value {
  white-space: normal;
  text-align: end;
  overflow-wrap: anywhere;
}
@media (max-width: 599px) {
  .setting-choice {
    max-width: 112px;
    padding-inline: 8px;
  }
}
</style>
