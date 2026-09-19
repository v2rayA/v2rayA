<script setup lang="ts">
// Renders the banners from useBanner at the top of the page, Material's
// banner anatomy: an icon, one or two lines of text, the action and the
// dismiss at the end.
import { useI18n } from "vue-i18n";
import {
  mdiAlertCircleOutline,
  mdiAlertOutline,
  mdiClose,
  mdiInformationOutline,
} from "@mdi/js";
import { bannerState, withdrawBanner } from "@/composables/useBanner";

const { t } = useI18n();
const icons = {
  info: mdiInformationOutline,
  warning: mdiAlertOutline,
  error: mdiAlertCircleOutline,
};
// Material's banner sits on a surface container; the kind shows in the
// icon's colour, so the banner follows the theme instead of painting the
// page red or green
const iconColors = {
  info: "primary",
  warning: "error",
  error: "error",
};
</script>

<template>
  <v-banner
    v-for="b in bannerState.list"
    :key="b.key"
    :icon="icons[b.kind]"
    :icon-color="iconColors[b.kind]"
    bg-color="surface-container-high"
    :text="b.text"
    lines="two"
    rounded="lg"
    class="mb-3"
    density="comfortable"
  >
    <template #actions>
      <v-btn v-if="b.action" variant="text" @click="b.action.onClick">{{
        b.action.label
      }}</v-btn>
      <v-btn
        v-if="b.dismissible"
        :icon="mdiClose"
        variant="text"
        size="small"
        :aria-label="t('operations.cancel')"
        @click="withdrawBanner(b.key)"
      />
    </template>
  </v-banner>
</template>
