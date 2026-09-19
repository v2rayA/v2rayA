<script setup lang="ts">
// Two signs that something runs, mounted once inside v-app by the shell:
// a linear progress bar along the top while any request is in flight
// (shown only once it has taken a moment, so quick calls do not flash),
// and the full-page overlay behind useLoading for actions that block.
import { computed, onUnmounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { requestActivity } from "@/api/client";
import { loadingState } from "@/composables/useLoading";

const { t } = useI18n();
const active = computed(() => loadingState.open.size > 0);
const busy = computed(() => requestActivity.inFlight > 0);
const shown = ref(false);
let timer: ReturnType<typeof setTimeout> | undefined;
watch(
  busy,
  (value) => {
    clearTimeout(timer);
    if (value) timer = setTimeout(() => (shown.value = true), 200);
    else shown.value = false;
  },
  { immediate: true },
);
onUnmounted(() => clearTimeout(timer));
</script>

<template>
  <v-fade-transition>
    <v-progress-linear
      v-if="shown"
      indeterminate
      color="primary"
      height="3"
      class="loading-bar"
      :aria-label="t('common.checkRunning')"
    />
  </v-fade-transition>
  <v-overlay
    :model-value="active"
    persistent
    class="align-center justify-center"
    scrim="surface"
  >
    <v-progress-circular indeterminate color="primary" size="48" />
  </v-overlay>
</template>

<style scoped>
.loading-bar {
  position: fixed;
  inset-block-start: 0;
  inset-inline: 0;
  z-index: 2010;
}
</style>
