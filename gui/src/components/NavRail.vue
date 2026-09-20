<script setup lang="ts">
// The navigation rail (Material 3: 80 dp wide at the start edge, the
// brand at the top, one item per destination with the 56×32 indicator
// pill behind the active icon and the label under it). Shown from
// 600 dp up to 840, where the drawer takes over; the bottom bar below.
// Above that it stands in for a folded drawer, with the menu button
// that unfolds it at the bottom, where the drawer's is.
import { useI18n } from "vue-i18n";
import { mdiBackburger, mdiForwardburger } from "@mdi/js";
import { useRtl } from "vuetify";
import { destinations } from "./destinations";
import { useAppStore } from "@/stores/app";
import BrandShape from "./BrandShape.vue";

defineProps<{ foldable?: boolean }>();
const emit = defineEmits<{ unfold: [] }>();
const { t } = useI18n();
const { isRtl } = useRtl();
const store = useAppStore();
</script>

<template>
  <v-navigation-drawer permanent :width="80" color="surface" class="rail">
    <div class="rail__brand">
      <BrandShape :size="56" />
    </div>
    <nav class="rail__items" :aria-label="t('common.menu')">
      <v-btn
        v-for="d in destinations"
        :key="d.view"
        variant="text"
        stacked
        height="56"
        class="rail__item"
        :class="{ 'rail__item--active': store.view === d.view }"
        :aria-current="store.view === d.view ? 'page' : undefined"
        @click="store.view = d.view"
      >
        <span class="rail__indicator">
          <v-icon
            :icon="store.view === d.view ? d.activeIcon : d.icon"
            size="24"
          />
        </span>
        <span class="md3-label-medium rail__label">{{ t(d.label) }}</span>
      </v-btn>
    </nav>
    <template v-if="foldable" #append>
      <v-btn
        :icon="isRtl ? mdiBackburger : mdiForwardburger"
        variant="text"
        class="rail__menu"
        :aria-label="t('common.menu')"
        aria-expanded="false"
        @click="emit('unfold')"
      />
    </template>
  </v-navigation-drawer>
</template>

<style scoped>
.rail :deep(.v-navigation-drawer__content) {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding-top: 12px;
}
.rail__menu {
  display: flex;
  margin: 12px auto 20px;
  color: rgb(var(--v-theme-outline));
}
.rail__brand {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 20px;
}
.rail__items {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
/* Material's rail item: a 56×32 pill indicator over the label, inside
   Vuetify's own button so the state layers and focus come from it */
.rail__item {
  width: 80px;
  min-width: 0;
  padding: 0;
  color: rgb(var(--v-theme-on-surface-variant));
}
.rail__item :deep(.v-btn__content) {
  gap: 4px;
}
.rail__item--active {
  color: rgb(var(--v-theme-on-surface));
}
.rail__item--active :deep(.v-btn__overlay) {
  opacity: 0;
}
.rail__indicator {
  width: 56px;
  height: 32px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 200ms cubic-bezier(0.2, 0, 0, 1);
}
.rail__item--active .rail__indicator {
  background: rgb(var(--v-theme-secondary-container));
  color: rgb(var(--v-theme-on-secondary-container));
}
.rail__label {
  text-transform: none;
  letter-spacing: 0.5px;
}
</style>
