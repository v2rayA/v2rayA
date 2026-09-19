<script setup lang="ts">
// The log page: a toolbar of filters (level chips, source, refresh
// period, follow), the lines in a virtual scroller with the accesslog
// highlighting, and export of what is shown.
import { computed, nextTick, onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { useDisplay } from "vuetify";
import { mdiMagnify, mdiTrayArrowDown } from "@mdi/js";
import hljs from "highlight.js/lib/core";
import accesslog from "highlight.js/lib/languages/accesslog";
import { useLogStream } from "@/composables/useLogStream";

defineOptions({ name: "LogsView" });
hljs.registerLanguage("accesslog", accesslog);
const { t } = useI18n();
const { width } = useDisplay();
const compact = computed(() => width.value < 600);
const stream = useLogStream(5);
const { lines, sources, interval } = stream;

const levels = [
  "all",
  "error",
  "warn",
  "info",
  "debug",
  "trace",
  "other",
] as const;
const level = ref<string>("all");
const source = ref("all");
const query = ref("");
const follow = ref(localStorage.getItem("log.follow") !== "false");
const intervals = [2, 5, 10, 15];
const scroller = ref<{
  scrollToIndex(i: number): void;
  $el: HTMLElement;
} | null>(null);

const shown = computed(() => {
  let list = lines.value;
  if (level.value !== "all") list = list.filter((l) => l.level === level.value);
  if (source.value !== "all")
    list = list.filter((l) => l.source === source.value);
  const search = query.value.trim().toLowerCase();
  if (search) list = list.filter((l) => l.text.toLowerCase().includes(search));
  return list;
});
const sourceItems = computed(() => [
  { value: "all", title: t("log.sources.all") },
  ...sources.value.map((s) => ({ value: s, title: s })),
]);
const intervalItems = intervals.map((n) => ({
  value: n,
  title: `${n} ${t("log.seconds")}`,
}));

const highlight = (text: string) =>
  hljs.highlight(text, { language: "accesslog", ignoreIllegals: true }).value;

watch(follow, (v) => localStorage.setItem("log.follow", String(v)));
watch(
  () => shown.value.length,
  async () => {
    if (!follow.value) return;
    await nextTick();
    scroller.value?.scrollToIndex(shown.value.length - 1);
  },
);

function exportShown() {
  const text = shown.value.map((l) => l.text).join("\n") + "\n";
  const stamp = new Date().toISOString().replace(/[:T]/g, "-").slice(0, 19);
  const url = URL.createObjectURL(
    new Blob([text], { type: "text/plain;charset=utf-8" }),
  );
  const a = document.createElement("a");
  a.href = url;
  a.download = `v2raya-log-${stamp}.txt`;
  a.click();
  setTimeout(() => URL.revokeObjectURL(url), 1000);
}

function onKey(event: KeyboardEvent) {
  const el = scroller.value?.$el;
  if (!el) return;
  const step = Math.max(1, Math.floor(el.clientHeight * 0.9));
  const moves: Record<string, () => void> = {
    Home: () => (el.scrollTop = 0),
    End: () => (el.scrollTop = el.scrollHeight),
    PageUp: () => (el.scrollTop = Math.max(0, el.scrollTop - step)),
    PageDown: () =>
      (el.scrollTop = Math.min(el.scrollHeight, el.scrollTop + step)),
  };
  if (moves[event.key]) {
    moves[event.key]();
    event.preventDefault();
  }
}

onMounted(() => {
  void stream.fetch();
  stream.start();
});
defineExpose({ sync: () => stream.fetch() });
</script>

<template>
  <div class="logs">
    <v-sheet color="surface-container-low" rounded="xl" class="pa-4 mb-4">
      <div class="d-flex flex-wrap align-center ga-2">
        <v-chip-group v-model="level" mandatory>
          <v-chip
            v-for="l in levels"
            :key="l"
            :value="l"
            variant="outlined"
            filter
          >
            {{ t(`log.categories.${l}`) }}
          </v-chip>
        </v-chip-group>
      </div>
      <div class="d-flex flex-wrap align-center ga-3 mt-3">
        <v-text-field
          v-model="query"
          :placeholder="t('log.search')"
          :prepend-inner-icon="mdiMagnify"
          variant="solo-filled"
          bg-color="surface-container-high"
          rounded="pill"
          density="compact"
          flat
          hide-details
          clearable
          class="logs__search"
          @click:clear="query = ''"
        />
        <v-select
          v-model="source"
          :items="sourceItems"
          :label="t('log.source')"
          variant="outlined"
          density="compact"
          hide-details
          class="logs__select"
        />
        <v-select
          v-model="interval"
          :items="intervalItems"
          :label="t('log.refreshInterval')"
          variant="outlined"
          density="compact"
          hide-details
          class="logs__select"
        />
        <v-switch
          v-model="follow"
          :label="t('log.autoShowNew')"
          hide-details
          density="compact"
          class="ms-1"
        />
        <v-spacer />
        <v-btn
          variant="tonal"
          :prepend-icon="mdiTrayArrowDown"
          @click="exportShown"
        >
          {{ t("log.export") }}
        </v-btn>
      </div>
    </v-sheet>

    <v-sheet
      color="surface-container-lowest"
      rounded="xl"
      class="logs__pane pa-2"
      tabindex="0"
      @keydown="onKey"
    >
      <v-empty-state
        v-if="!shown.length"
        :title="t('log.logsLabel')"
        class="py-8"
      />
      <v-virtual-scroll
        v-else
        ref="scroller"
        :items="shown"
        :item-height="compact ? undefined : 28"
        class="logs__scroll"
      >
        <template #default="{ item, index }">
          <div class="logs__row" :class="{ 'logs__row--wrap': compact }">
            <span class="logs__number">{{ index + 1 }}</span>
            <span
              class="logs__text language-accesslog"
              v-html="highlight(item.text)"
            />
          </div>
        </template>
      </v-virtual-scroll>
    </v-sheet>
  </div>
</template>

<style scoped>
.logs__select {
  max-width: 220px;
  min-width: 160px;
}
.logs__search {
  flex: 1 1 240px;
  max-width: 420px;
}
/* the pane fills what the toolbar leaves of the viewport */
.logs__scroll {
  height: max(320px, calc(100dvh - 300px));
}
.logs__pane:focus-visible {
  outline: 2px solid rgb(var(--v-theme-primary));
}
.logs__row {
  display: flex;
  align-items: center;
  gap: 12px;
  height: 28px;
  padding: 0 8px;
  white-space: nowrap;
}
.logs__row--wrap {
  height: auto;
  align-items: flex-start;
  white-space: pre-wrap;
  word-break: break-word;
  padding: 4px 8px;
}
.logs__number {
  min-width: 3em;
  text-align: end;
  color: rgb(var(--v-theme-outline));
  font-variant-numeric: tabular-nums;
  font-size: 12px;
}
.logs__text {
  font-family: ui-monospace, "Cascadia Mono", "Fira Mono", Menlo, monospace;
  font-size: 13px;
  direction: ltr;
  unicode-bidi: isolate;
}
/* the accesslog grammar's tokens, in the theme's colours */
.logs__text :deep(.hljs-number) {
  color: rgb(var(--v-theme-tertiary));
}
.logs__text :deep(.hljs-string) {
  color: rgb(var(--v-theme-primary));
}
</style>
