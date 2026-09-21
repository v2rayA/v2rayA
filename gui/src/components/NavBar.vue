<script setup lang="ts">
// The navigation bar for compact windows (Material 3: 80 dp at the
// bottom on surface-container, one item per destination with the 64×32
// indicator pill behind the active icon and the label under it).
import { useI18n } from "vue-i18n";
import { barDestinations } from "./destinations";
import { useAppStore, type View } from "@/stores/app";

const { t } = useI18n();
const store = useAppStore();
</script>

<template>
  <v-bottom-navigation
    :model-value="store.view"
    :height="80"
    bg-color="surface-container"
    class="bar"
    tag="nav"
    @update:model-value="
      (v: unknown) => typeof v === 'string' && (store.view = v as View)
    "
  >
    <v-btn
      v-for="d in barDestinations()"
      :key="d.view"
      :value="d.view"
      variant="text"
      height="80"
      class="bar__item"
      :aria-current="store.view === d.view ? 'page' : undefined"
    >
      <span class="bar__indicator">
        <v-icon
          :icon="store.view === d.view ? d.activeIcon : d.icon"
          size="24"
        />
      </span>
      <span class="md3-label-medium bar__label">{{ t(d.label) }}</span>
    </v-btn>
  </v-bottom-navigation>
</template>

<style scoped>
/* Material's navigation bar item: a 64×32 pill indicator over the label,
   drawn inside Vuetify's own button so the state layers and focus come
   from the component */
.bar :deep(.v-bottom-navigation__content) {
  display: flex;
  justify-content: space-evenly;
  align-items: stretch;
}
.bar__item {
  flex: 1;
  min-width: 0;
  color: rgb(var(--v-theme-on-surface-variant));
}
.bar__item :deep(.v-btn__content) {
  flex-direction: column;
  gap: 4px;
}
.bar__item.v-btn--active {
  color: rgb(var(--v-theme-on-surface));
}
.bar__item.v-btn--active :deep(.v-btn__overlay) {
  opacity: 0;
}
.bar__indicator {
  width: 64px;
  height: 32px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 200ms cubic-bezier(0.2, 0, 0, 1);
}
.bar__item.v-btn--active .bar__indicator {
  background: rgb(var(--v-theme-secondary-container));
  color: rgb(var(--v-theme-on-secondary-container));
}
.bar__label {
  text-transform: none;
  letter-spacing: 0.5px;
}
</style>
