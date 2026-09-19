<script setup lang="ts">
// The theme, language and account menus of the shell: icon buttons in
// the top app bar (medium and compact windows), list items at the bottom
// of the navigation drawer (expanded windows).
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import {
  mdiAccountCircleOutline,
  mdiLogout,
  mdiPalette,
  mdiTranslate,
} from "@mdi/js";
import { resetSession } from "@/session";
import { useAppStore } from "@/stores/app";
import { languages } from "./languages";
import ThemePanel from "./ThemePanel.vue";

defineProps<{ variant: "icons" | "list" }>();
const { t, locale } = useI18n();
const store = useAppStore();

const currentLang = computed(
  () => languages.find((l) => l.flag === locale.value)?.label ?? locale.value,
);
function setLanguage(flag: string) {
  store.setLanguage(flag);
  locale.value = flag;
}
function logout() {
  void resetSession({ token: "" });
}
</script>

<template>
  <v-menu :close-on-content-click="false">
    <template #activator="{ props: menu }">
      <v-btn
        v-if="variant === 'icons'"
        v-bind="menu"
        :icon="mdiPalette"
        variant="text"
        class="ms-1"
        :aria-label="t('theme.title')"
      />
      <v-list-item
        v-else
        v-bind="menu"
        :prepend-icon="mdiPalette"
        :title="t('theme.title')"
        rounded="xl"
      />
    </template>
    <v-card class="theme-panel">
      <ThemePanel />
    </v-card>
  </v-menu>

  <v-menu>
    <template #activator="{ props: menu }">
      <v-btn
        v-if="variant === 'icons'"
        v-bind="menu"
        :icon="mdiTranslate"
        variant="text"
        class="ms-1"
        :aria-label="currentLang"
      />
      <v-list-item
        v-else
        v-bind="menu"
        :prepend-icon="mdiTranslate"
        :title="currentLang"
        rounded="xl"
      />
    </template>
    <v-list density="compact" min-width="220">
      <v-list-item
        v-for="lang in languages"
        :key="lang.code"
        :active="lang.flag === locale"
        :title="lang.label"
        :subtitle="lang.code"
        @click="setLanguage(lang.flag)"
      />
    </v-list>
  </v-menu>

  <v-menu>
    <template #activator="{ props: menu }">
      <v-btn
        v-if="variant === 'icons'"
        v-bind="menu"
        :icon="mdiAccountCircleOutline"
        variant="text"
        class="ms-1 me-1"
        :aria-label="store.username || t('common.notLogin')"
      />
      <v-list-item
        v-else
        v-bind="menu"
        :prepend-icon="mdiAccountCircleOutline"
        :title="store.username || t('common.notLogin')"
        rounded="xl"
      />
    </template>
    <v-list density="compact" min-width="240">
      <v-list-item disabled>
        <v-list-item-title class="md3-body-medium">
          <i18n-t keypath="common.loggedAs" tag="span" scope="global">
            <template #username>
              <b>{{ store.username || t("common.notLogin") }}</b>
            </template>
          </i18n-t>
        </v-list-item-title>
      </v-list-item>
      <v-divider />
      <v-list-item
        :prepend-icon="mdiLogout"
        :title="t('operations.logout')"
        @click="logout"
      />
    </v-list>
  </v-menu>
</template>

<style scoped>
.theme-panel {
  width: min(392px, calc(100vw - 32px));
}
</style>
