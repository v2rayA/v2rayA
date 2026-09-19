<script setup lang="ts">
// The standard navigation drawer for expanded windows (≥ 840 dp): the
// brand, the destinations, the version at the bottom. There is no top
// app bar at this width: the page titles itself and carries the theme,
// language and account menus; the core's state lives on the dashboard.
import { useI18n } from "vue-i18n";
import { destinations } from "./destinations";
import { useDialog } from "@/composables";
import AboutDialog from "@/views/settings/AboutDialog.vue";
import { useAppStore } from "@/stores/app";
import logo from "@/assets/img/v2raya-icon.svg";

const { t } = useI18n();
const store = useAppStore();
const { open } = useDialog();
const openAbout = () => open(AboutDialog, {}, { width: 640 });
</script>

<template>
  <v-navigation-drawer permanent :width="256" color="surface" class="drawer">
    <div class="drawer__brand">
      <img :src="logo" alt="" class="drawer__logo" />
      <span class="md3-title-large drawer__wordmark">v2rayA</span>
    </div>
    <v-list nav density="default" class="px-3 pt-2 pb-0">
      <v-list-item
        v-for="d in destinations"
        :key="d.view"
        :active="store.view === d.view"
        :prepend-icon="store.view === d.view ? d.activeIcon : d.icon"
        :title="t(d.label)"
        rounded="xl"
        class="drawer__item"
        @click="store.view = d.view"
      />
    </v-list>
    <template #append>
      <v-divider class="mx-7" />
      <v-btn
        variant="plain"
        class="drawer__version text-none"
        @click="openAbout"
      >
        <span class="md3-label-large"
          >v2rayA {{ store.version?.version ?? "" }}</span
        >
        <v-tooltip activator="parent" location="end" :offset="12">
          {{ t("common.about") }}
        </v-tooltip>
      </v-btn>
    </template>
  </v-navigation-drawer>
</template>

<style scoped>
.drawer__brand {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 28px 28px 20px;
}
.drawer__wordmark {
  font-weight: 500;
}
/* the version is a footnote aligned with the brand: outline text, no container; it opens About */
.drawer__version {
  height: auto;
  margin: 12px 16px 20px;
  padding: 4px 12px;
  color: rgb(var(--v-theme-outline));
  opacity: 1;
}
.drawer__logo {
  width: 28px;
  height: 28px;
}
/* Material's drawer item: 56 dp tall, label-large */
.drawer :deep(.v-list-item) {
  min-height: 48px;
  padding-inline: 16px;
}
.drawer :deep(.v-list-item__spacer) {
  width: 16px;
}
.drawer :deep(.v-list-item-title) {
  font-size: 14px;
  font-weight: 500;
  letter-spacing: 0.1px;
}
.drawer :deep(.v-list-item__prepend > .v-icon) {
  opacity: 1;
}
/* the active destination: Material's secondary-container pill */
.drawer :deep(.v-list-item--active) {
  background: rgb(var(--v-theme-secondary-container));
  color: rgb(var(--v-theme-on-secondary-container));
}
.drawer :deep(.v-list-item--active .v-list-item__overlay) {
  opacity: 0;
}
</style>
