<script setup lang="ts">
// The process names whose traffic bypasses the TUN, edited as a list and
// returned to the settings form as the comma-joined value it stores.
import { ref } from "vue";
import { useI18n } from "vue-i18n";

defineOptions({ name: "TunProcessesDialog" });
const props = defineProps<{ value: string }>();
const emit = defineEmits<{ close: [value?: string] }>();
const { t } = useI18n();

const parse = (raw: string) => [
  ...new Set(
    raw
      .split(/[\n,;\t]/g)
      .map((p) => p.trim())
      .filter((p) => p),
  ),
];
const text = ref(parse(props.value).join("\n"));
</script>

<template>
  <v-card>
    <v-card-item class="px-6 pt-6 pb-2">
      <v-card-title class="md3-headline-small pa-0">
        {{ t("tun.processExclude.title") }}
      </v-card-title>
    </v-card-item>
    <v-card-text class="px-6">
      <v-alert
        type="warning"
        variant="tonal"
        density="compact"
        class="md3-body-small mb-4"
      >
        {{ t("tun.processExclude.warning") }}
      </v-alert>
      <v-textarea
        v-model="text"
        :label="t('tun.processExclude.listLabel')"
        :placeholder="t('tun.processExclude.placeholder')"
        :hint="t('tun.processExclude.hint')"
        persistent-hint
        rows="6"
        auto-grow
        dir="ltr"
        autocomplete="off"
        spellcheck="false"
        @keydown.enter="
          (e: KeyboardEvent) =>
            (e.ctrlKey || e.metaKey) && emit('close', parse(text).join(','))
        "
      />
    </v-card-text>
    <v-card-actions class="px-6 pb-4">
      <v-spacer />
      <v-btn variant="text" @click="emit('close')">{{
        t("operations.cancel")
      }}</v-btn>
      <v-btn
        variant="flat"
        color="primary"
        @click="emit('close', parse(text).join(','))"
      >
        {{ t("operations.save") }}
      </v-btn>
    </v-card-actions>
  </v-card>
</template>
