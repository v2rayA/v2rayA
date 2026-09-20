<script setup lang="ts">
// A small "read the documentation" button that opens a section of the
// in-app docs; the shell turns the "#docs/<section>" hash into the page.
import { useI18n } from "vue-i18n";
import { mdiBookOpenPageVariantOutline } from "@mdi/js";
import type { Section } from "@/docs";

// from a modal dialog the docs open in a new tab, so the dialog stays
defineProps<{ section: Section; newTab?: boolean }>();
const { t } = useI18n();
</script>

<template>
  <v-tooltip :text="t('common.docs')">
    <template #activator="{ props }">
      <v-btn
        v-bind="props"
        :icon="mdiBookOpenPageVariantOutline"
        variant="text"
        size="small"
        :href="`#docs/${section}`"
        :target="newTab ? '_blank' : undefined"
        :aria-label="t('common.docs')"
      />
    </template>
  </v-tooltip>
</template>
