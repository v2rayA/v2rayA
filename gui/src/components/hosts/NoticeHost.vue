<script setup lang="ts">
// Renders the notice queue: one snackbar, the next when it ends. Mounted
// once inside v-app by the shell.
import { computed } from "vue";
import {
  mdiAlertCircleOutline,
  mdiAlertOutline,
  mdiCheckCircleOutline,
  mdiInformationOutline,
} from "@mdi/js";
import { dismissNotice, noticeState } from "@/composables/useNotify";

const current = computed(() => noticeState.current);
// The snackbar sits on the highest surface container, in the theme's own
// tint; the kind shows in a leading icon and the action in primary
const icons = {
  info: mdiInformationOutline,
  success: mdiCheckCircleOutline,
  warning: mdiAlertOutline,
  error: mdiAlertCircleOutline,
};
const iconColors = {
  info: "primary",
  success: "primary",
  warning: "error",
  error: "error",
};
const shown = computed({
  get: () => current.value !== null,
  set: (v: boolean) => {
    if (!v && current.value) dismissNotice(current.value.id);
  },
});
</script>

<template>
  <v-snackbar
    v-if="current"
    :key="current.id"
    v-model="shown"
    :timeout="current.timeout || -1"
    color="surface-container-highest"
    location="bottom"
    variant="flat"
    rounded="lg"
  >
    <div class="d-flex align-center ga-3">
      <v-icon
        :icon="icons[current.kind]"
        :color="iconColors[current.kind]"
        size="20"
      />
      <span class="md3-body-medium">{{ current.text }}</span>
    </div>
    <template v-if="current.action" #actions>
      <v-btn
        variant="text"
        color="primary"
        @click="
          current.action.onClick();
          dismissNotice(current.id);
        "
        >{{ current.action.label }}</v-btn
      >
    </template>
  </v-snackbar>
</template>
