<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, ref, useId, watch } from "vue";
import { useI18n } from "vue-i18n";
import HighlightedCode from "./HighlightedCode.vue";
import { check, type LineError } from "./check";

defineOptions({ name: "RoutingEditor" });
const props = defineProps<{
  modelValue: string;
  disabled?: boolean;
  readonly?: boolean;
}>();
const emit = defineEmits<{ "update:modelValue": [value: string]; save: [] }>();
const { t } = useI18n();
const textarea = ref<HTMLTextAreaElement>();
const highlight = ref<InstanceType<typeof HighlightedCode>>();
const gutter = ref<HTMLElement>();
const currentLine = ref(1);
const errors = ref<LineError[]>([]);
const errorId = useId();
const lines = computed(() => props.modelValue.split("\n").length);
const messages = computed(
  () => new Map(errors.value.map((error) => [error.line, t(error.message)])),
);
let timer: ReturnType<typeof setTimeout>;
watch(
  () => props.modelValue,
  (value) => {
    clearTimeout(timer);
    timer = setTimeout(() => {
      errors.value = check(value);
    }, 150);
    nextTick(syncScroll);
  },
  { immediate: true },
);
onBeforeUnmount(() => clearTimeout(timer));

function syncScroll() {
  const input = textarea.value;
  const pre = highlight.value?.$el as HTMLElement | undefined;
  if (!input || !pre || !gutter.value) return;
  pre.scrollTop = input.scrollTop;
  pre.scrollLeft = input.scrollLeft;
  gutter.value.scrollTop = input.scrollTop;
}
function selection() {
  const input = textarea.value;
  if (input)
    currentLine.value = input.value
      .slice(0, input.selectionStart)
      .split("\n").length;
}
function input() {
  emit("update:modelValue", textarea.value!.value);
  selection();
}
async function insert(text: string, wholeLines = false) {
  const field = textarea.value;
  if (!field || props.disabled || props.readonly) return;
  if (wholeLines) {
    const before = field.value[field.selectionStart - 1];
    const after = field.value[field.selectionEnd];
    text =
      (before && before !== "\n" ? "\n" : "") +
      text +
      (after && after !== "\n" ? "\n" : "");
  }
  field.setRangeText(text, field.selectionStart, field.selectionEnd, "end");
  input();
  const caret = field.selectionStart;
  await nextTick();
  field.focus();
  field.setSelectionRange(caret, caret);
  selection();
  syncScroll();
}
function keydown(event: KeyboardEvent) {
  if (event.isComposing) return;
  if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === "s") {
    event.preventDefault();
    emit("save");
  } else if (
    event.key === "Tab" &&
    !event.shiftKey &&
    !event.ctrlKey &&
    !event.metaKey &&
    !event.altKey
  ) {
    event.preventDefault();
    void insert("  ");
  } else if (
    event.key === "Enter" &&
    !event.ctrlKey &&
    !event.metaKey &&
    !event.altKey
  ) {
    event.preventDefault();
    const field = textarea.value!;
    const previous = field.value
      .slice(0, field.selectionStart)
      .split("\n")
      .at(-1)!;
    void insert("\n" + previous.match(/^[\t ]*/)![0]);
  }
}
defineExpose({ insert });
</script>

<template>
  <v-sheet
    class="routing-editor"
    rounded="lg"
    color="surface-container-lowest"
    dir="ltr"
  >
    <div ref="gutter" class="routing-editor__gutter" aria-hidden="true">
      <div
        v-for="line in lines"
        :key="line"
        class="routing-editor__number"
        :class="{
          'routing-editor__number--current': currentLine === line,
          'routing-editor__number--error': messages.has(line),
        }"
        :title="messages.get(line)"
      >
        <span class="routing-editor__dot">{{
          messages.has(line) ? "•" : ""
        }}</span
        >{{ line }}
      </div>
    </div>
    <div class="routing-editor__content">
      <HighlightedCode
        ref="highlight"
        :text="modelValue"
        class="routing-editor__highlight"
        aria-hidden="true"
      />
      <textarea
        ref="textarea"
        :value="modelValue"
        :disabled="disabled"
        :readonly="readonly"
        :aria-label="t('routingA.title')"
        :aria-describedby="errors.length ? errorId : undefined"
        class="routing-editor__input"
        dir="ltr"
        spellcheck="false"
        autocapitalize="off"
        autocomplete="off"
        autocorrect="off"
        wrap="off"
        @input="input"
        @scroll="syncScroll"
        @keydown="keydown"
        @keyup="selection"
        @click="selection"
        @select="selection"
      />
    </div>
    <span :id="errorId" class="d-sr-only" aria-live="polite">{{
      errors
        .map((error) =>
          t("routingA.lineError", {
            line: error.line,
            message: t(error.message),
          }),
        )
        .join("\n")
    }}</span>
  </v-sheet>
</template>

<style scoped>
.routing-editor {
  display: grid;
  grid-template-columns: max-content minmax(0, 1fr);
  height: 60vh;
  min-height: 320px;
  max-height: 60vh;
  min-width: 0;
  overflow: hidden;
  border: 1px solid rgb(var(--v-theme-outline-variant));
}
.routing-editor:focus-within {
  border-color: rgb(var(--v-theme-primary));
  box-shadow: inset 0 0 0 1px rgb(var(--v-theme-primary));
}
.routing-editor__gutter {
  overflow: hidden;
  padding: 16px 8px 32px;
  font:
    13px / 20px ui-monospace,
    "Cascadia Mono",
    "Fira Mono",
    Menlo,
    monospace;
  font-variant-numeric: tabular-nums;
  text-align: right;
  color: rgb(var(--v-theme-outline));
}
.routing-editor__number {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  height: 20px;
}
.routing-editor__number--current {
  color: rgb(var(--v-theme-on-surface));
}
.routing-editor__dot {
  width: 8px;
  color: rgb(var(--v-theme-error));
}
.routing-editor__content {
  position: relative;
  min-width: 0;
  min-height: 0;
}
.routing-editor__highlight,
.routing-editor__input {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  margin: 0;
  border: 0;
  padding: 16px 8px;
  overflow: scroll;
  font:
    13px / 20px ui-monospace,
    "Cascadia Mono",
    "Fira Mono",
    Menlo,
    monospace;
  font-variant-ligatures: none;
  letter-spacing: normal;
  tab-size: 2;
  white-space: pre;
}
.routing-editor__highlight {
  pointer-events: none;
}
.routing-editor__input {
  resize: none;
  outline: none;
  color: transparent;
  background: transparent;
  caret-color: rgb(var(--v-theme-on-surface));
}
.routing-editor__input::selection {
  background: rgba(var(--v-theme-primary), 0.24);
}
@media (forced-colors: active) {
  .routing-editor__input {
    color: CanvasText;
  }
  .routing-editor__highlight {
    visibility: hidden;
  }
}
</style>
