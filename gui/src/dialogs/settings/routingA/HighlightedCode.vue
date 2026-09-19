<script setup lang="ts">
import { computed } from "vue";
import { tokenize } from "./highlight";

defineOptions({ name: "RoutingHighlightedCode" });
const props = defineProps<{ text: string }>();
const lines = computed(() => props.text.split("\n").map(tokenize));
</script>

<template>
  <pre class="routing-code" dir="ltr"><span
    v-for="(line, index) in lines"
    :key="index"
    class="routing-code__line"
  ><span v-for="(token, part) in line" :key="part" :class="`tok-${token.type}`">{{ token.text }}</span>{{ line.length ? "" : "\n" }}</span></pre>
</template>

<style scoped>
.routing-code {
  margin: 0;
  font:
    13px / 20px ui-monospace,
    "Cascadia Mono",
    "Fira Mono",
    Menlo,
    monospace;
  font-variant-ligatures: none;
  tab-size: 2;
  white-space: pre;
  color: rgb(var(--v-theme-on-surface));
}
.routing-code__line {
  display: block;
  min-height: 20px;
}
.tok-comment {
  color: rgb(var(--v-theme-outline));
}
.tok-keyword,
.tok-outbound {
  color: rgb(var(--v-theme-primary));
}
.tok-keyword {
  font-weight: 500;
}
.tok-function {
  color: rgb(var(--v-theme-tertiary));
}
.tok-argument,
.tok-string {
  color: rgb(var(--v-theme-secondary));
}
.tok-arrow,
.tok-operator {
  color: rgb(var(--v-theme-on-surface-variant));
}
.tok-number {
  color: rgb(var(--v-theme-on-surface));
}
</style>
