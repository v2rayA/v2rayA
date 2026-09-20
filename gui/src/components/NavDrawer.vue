<script setup lang="ts">
// The standard navigation drawer for expanded windows (≥ 840 dp): the
// brand, the destinations, the version at the bottom. There is no top
// app bar at this width: the page titles itself and carries the theme,
// language and account menus; the core's state lives on the dashboard.
// The menu button at the bottom folds the drawer to a rail (Vuetify's
// rail mode) and the choice is remembered.
import { ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { mdiMenu, mdiMenuOpen } from "@mdi/js";
import { destinations } from "./destinations";
import { useDialog } from "@/composables";
import AboutDialog from "@/views/settings/AboutDialog.vue";
import { useAppStore } from "@/stores/app";
import logo from "@/assets/img/v2raya-icon.svg";

const { t } = useI18n();
const store = useAppStore();
const { open } = useDialog();
const openAbout = () => open(AboutDialog, {}, { width: 640 });
const rail = ref(localStorage.getItem("drawer") === "rail");
watch(rail, (v) => localStorage.setItem("drawer", v ? "rail" : "open"));
</script>

<template>
  <v-navigation-drawer
    permanent
    :rail="rail"
    :width="256"
    :rail-width="80"
    color="surface"
    class="drawer"
  >
    <div class="drawer__brand">
      <img :src="logo" alt="" class="drawer__logo" />
      <span v-if="!rail" class="md3-title-large drawer__wordmark">v2rayA</span>
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
        :icon="rail ? mdiMenu : mdiMenuOpen"
        variant="text"
        class="drawer__toggle"
        :aria-label="t('common.menu')"
        :aria-expanded="!rail"
        @click="rail = !rail"
      />
      <v-btn
        v-if="!rail"
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
.drawer__toggle {
  margin: 12px 16px 0;
}
.drawer.v-navigation-drawer--rail .drawer__toggle {
  margin-inline: auto;
  margin-bottom: 16px;
  display: flex;
}
/* folded: the logo and the icons sit centred in the 80 dp column */
.drawer.v-navigation-drawer--rail .drawer__brand {
  justify-content: center;
  padding-inline: 0;
}
.drawer.v-navigation-drawer--rail :deep(.v-list-item) {
  justify-content: center;
  padding-inline: 0;
}
.drawer.v-navigation-drawer--rail :deep(.v-list-item__prepend) {
  margin-inline-end: 0;
}
.drawer.v-navigation-drawer--rail
  :deep(.v-list-item__prepend .v-list-item__spacer),
.drawer.v-navigation-drawer--rail :deep(.v-list-item__content) {
  display: none;
}
.drawer__brand {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 28px 28px 20px;
  white-space: nowrap;
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
