<script setup lang="ts">
// The documentation: a section list beside the article from the medium
// window class, a select above it on compact. Sections are Markdown
// compiled at build time; the parameters section appends the flag table
// the service reports, so it never lags behind the binary.
import { computed, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { useDisplay } from "vuetify";
import { getParams } from "@/api";
import { errorText } from "@/api/errors";
import type { Param } from "@/api/types";
import {
  docsLocale,
  isSection,
  loadSection,
  sections,
  type Section,
} from "@/docs";
import { useAppStore } from "@/stores/app";

defineOptions({ name: "DocsView" });
const { t, locale } = useI18n();
const store = useAppStore();
const { width } = useDisplay();
const compact = computed(() => width.value < 600);

const section = ref<Section>(
  isSection(store.docsSection) ? store.docsSection : sections[0],
);
const html = ref("");
const loading = ref(false);
const params = ref<Param[]>([]);
const paramsError = ref("");
const items = computed(() =>
  sections.map((key) => ({ value: key, title: t(`docs.sections.${key}`) })),
);
// the content is written in English and Chinese; the other locales read English
const fallback = computed(() => docsLocale(locale.value) !== locale.value);

async function show() {
  loading.value = true;
  try {
    html.value = await loadSection(locale.value, section.value);
    if (section.value === "parameters" && !params.value.length) {
      try {
        params.value = (await getParams()).params;
        paramsError.value = "";
      } catch (err) {
        paramsError.value = errorText(err);
      }
    }
  } finally {
    loading.value = false;
  }
  document.querySelector(".docs__article")?.scrollTo?.(0, 0);
}
function select([value]: unknown[]) {
  if (typeof value === "string" && isSection(value)) section.value = value;
}
watch([section, locale], show, { immediate: true });
watch(section, (value) => (store.docsSection = value));
</script>

<template>
  <div class="docs" :class="{ 'docs--compact': compact }">
    <nav class="docs__nav" :aria-label="t('common.docs')">
      <v-select
        v-if="compact"
        v-model="section"
        :items="items"
        :label="t('common.docs')"
        variant="outlined"
        density="comfortable"
        hide-details
      />
      <v-list
        v-else
        :selected="[section]"
        mandatory
        density="compact"
        rounded="lg"
        bg-color="transparent"
        class="docs__list"
        @update:selected="select"
      >
        <v-list-item
          v-for="item in items"
          :key="item.value"
          :value="item.value"
          :title="item.title"
          rounded="xl"
        />
      </v-list>
    </nav>
    <v-sheet
      color="surface-container-low"
      rounded="xl"
      class="docs__article pa-5 pa-sm-8"
    >
      <v-alert
        v-if="fallback"
        type="info"
        variant="tonal"
        density="compact"
        class="mb-4"
        :text="t('docs.fallback')"
      />
      <v-progress-linear v-if="loading && !html" indeterminate />
      <!-- eslint-disable-next-line vue/no-v-html -- our own Markdown, compiled at build time -->
      <article class="docs__body md3-body-large" v-html="html" />
      <template v-if="section === 'parameters'">
        <v-alert
          v-if="paramsError"
          type="warning"
          variant="tonal"
          density="compact"
          :text="paramsError"
        />
        <v-table
          v-else-if="params.length"
          class="docs__params"
          density="comfortable"
        >
          <thead>
            <tr>
              <th>{{ t("docs.params.flag") }}</th>
              <th>{{ t("docs.params.env") }}</th>
              <th>{{ t("docs.params.default") }}</th>
              <th>{{ t("docs.params.desc") }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="p in params" :key="p.flag">
              <td>
                <code>{{ p.flag }}</code>
                <code v-if="p.short" class="ms-1">-{{ p.short }}</code>
              </td>
              <td>
                <code>{{ p.env }}</code>
              </td>
              <td>
                <code v-if="p.default">{{ p.default }}</code>
              </td>
              <td>{{ p.desc }}</td>
            </tr>
          </tbody>
        </v-table>
      </template>
    </v-sheet>
  </div>
</template>

<style scoped>
.docs {
  display: grid;
  grid-template-columns: 220px minmax(0, 1fr);
  gap: 24px;
  align-items: start;
}
.docs--compact {
  grid-template-columns: minmax(0, 1fr);
  gap: 16px;
}
.docs__nav {
  position: sticky;
  top: 80px;
}
.docs--compact .docs__nav {
  position: static;
}
/* the current section: Material's secondary-container pill, as in the drawer */
.docs__list :deep(.v-list-item--active) {
  background: rgb(var(--v-theme-secondary-container));
  color: rgb(var(--v-theme-on-secondary-container));
}
.docs__list :deep(.v-list-item--active .v-list-item__overlay) {
  opacity: 0;
}
.docs__article {
  min-width: 0;
  overflow-wrap: anywhere;
}
/* the Material type scale: headline-medium, title-large, title-medium */
.docs__body :deep(h1) {
  font-size: 28px;
  line-height: 36px;
  font-weight: 400;
  margin: 0 0 16px;
}
.docs__body :deep(h2) {
  font-size: 22px;
  line-height: 28px;
  font-weight: 400;
  margin: 32px 0 12px;
}
.docs__body :deep(h3) {
  font-size: 16px;
  line-height: 24px;
  font-weight: 500;
  letter-spacing: 0.15px;
  margin: 24px 0 8px;
}
.docs__body :deep(p),
.docs__body :deep(ul),
.docs__body :deep(ol) {
  margin: 0 0 12px;
}
.docs__body :deep(li) {
  margin-bottom: 4px;
}
.docs__body :deep(a) {
  color: rgb(var(--v-theme-primary));
}
.docs__body :deep(code),
.docs__params code {
  font-family: ui-monospace, "Cascadia Mono", "Fira Mono", Menlo, monospace;
  font-size: 0.9em;
  padding: 1px 5px;
  border-radius: 6px;
  background: rgb(var(--v-theme-surface-container-high));
}
.docs__body :deep(pre) {
  margin: 0 0 16px;
  padding: 12px 16px;
  border-radius: 12px;
  overflow-x: auto;
  background: rgb(var(--v-theme-surface-container-high));
}
.docs__body :deep(pre code) {
  padding: 0;
  background: none;
  font-size: 0.875rem;
}
.docs__body :deep(table) {
  border-collapse: collapse;
  margin: 0 0 16px;
  width: 100%;
}
.docs__body :deep(th),
.docs__body :deep(td) {
  text-align: start;
  padding: 6px 10px;
  border-bottom: 1px solid rgb(var(--v-theme-outline-variant));
  vertical-align: top;
}
/* the first column names a condition or a path; keep it on one line */
.docs__body :deep(td:first-child) {
  white-space: nowrap;
}
.docs__body :deep(blockquote) {
  margin: 0 0 12px;
  padding: 4px 16px;
  border-inline-start: 3px solid rgb(var(--v-theme-primary));
  color: rgb(var(--v-theme-on-surface-variant));
}
.docs__params {
  background: transparent;
  margin-top: 8px;
}
</style>
