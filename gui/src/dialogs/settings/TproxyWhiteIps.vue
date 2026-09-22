<script setup lang="ts">
// The IP groups that bypass the core under the transparent proxy: the
// built-in groups as filter chips, custom CIDRs one per line.
import { computed, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { getTproxyWhiteIpGroups, putTproxyWhiteIpGroups } from "@/api";
import { errorText } from "@/api/errors";
import { useNotify } from "@/composables";

defineOptions({ name: "TproxyWhiteIpsDialog" });
const emit = defineEmits<{ close: [saved?: boolean] }>();
const { t } = useI18n();
const notify = useNotify();
const groups = ["CN", "PRIVATE", "US", "CLOUDFLARE"] as const;
const labels: Record<(typeof groups)[number], string> = {
  CN: "tproxyWhiteIpGroups.cn",
  PRIVATE: "tproxyWhiteIpGroups.private",
  US: "tproxyWhiteIpGroups.us",
  CLOUDFLARE: "tproxyWhiteIpGroups.cloudflare",
};
const selected = ref<string[]>([]);
const customIps = ref("");
const saving = ref(false);

const ipv4 =
  /^(?:25[0-5]|2[0-4]\d|1\d{2}|[1-9]?\d)(?:\.(?:25[0-5]|2[0-4]\d|1\d{2}|[1-9]?\d)){3}\/(?:[0-9]|[12]\d|3[0-2])$/;
const ipv6 =
  /^(?:(?:[A-Fa-f0-9]{1,4}:){7}[A-Fa-f0-9]{1,4}|(?:[A-Fa-f0-9]{1,4}:){1,7}:|(?:[A-Fa-f0-9]{1,4}:){1,6}:[A-Fa-f0-9]{1,4}|(?:[A-Fa-f0-9]{1,4}:){1,5}(?::[A-Fa-f0-9]{1,4}){1,2}|(?:[A-Fa-f0-9]{1,4}:){1,4}(?::[A-Fa-f0-9]{1,4}){1,3}|(?:[A-Fa-f0-9]{1,4}:){1,3}(?::[A-Fa-f0-9]{1,4}){1,4}|(?:[A-Fa-f0-9]{1,4}:){1,2}(?::[A-Fa-f0-9]{1,4}){1,5}|[A-Fa-f0-9]{1,4}:(?:(?::[A-Fa-f0-9]{1,4}){1,6})|:(?:(?::[A-Fa-f0-9]{1,4}){1,7}|:))\/(?:12[0-8]|1[01]\d|[1-9]?\d)$/;
const lines = computed(() =>
  customIps.value
    .split("\n")
    .map((l) => l.trim())
    .filter((l) => l !== ""),
);
const invalid = computed(() =>
  lines.value.some((l) => !ipv4.test(l) && !ipv6.test(l)),
);

onMounted(async () => {
  try {
    const res = await getTproxyWhiteIpGroups();
    selected.value = (res.countryCodes ?? []).filter((c) => c !== "NONE");
    customIps.value = (res.customIps ?? []).join("\n");
  } catch (err) {
    notify.warning(errorText(err));
  }
});

async function save() {
  if (saving.value) return;
  if (invalid.value) {
    notify.warning(t("tproxyWhiteIpGroups.invalidCustomIps"));
    return;
  }
  saving.value = true;
  try {
    await putTproxyWhiteIpGroups({
      countryCodes: selected.value.length ? selected.value : ["NONE"],
      customIps: lines.value,
    });
    emit("close", true);
  } catch (err) {
    notify.warning(
      t("tproxyWhiteIpGroups.saveFailed", { message: errorText(err) }),
    );
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <v-card>
    <v-card-item class="px-6 pt-6 pb-2">
      <v-card-title class="md3-headline-small pa-0">
        {{ t("tproxyWhiteIpGroups.title") }}
      </v-card-title>
    </v-card-item>
    <v-card-text class="px-6">
      <p class="md3-body-medium mb-4">
        {{ t("tproxyWhiteIpGroups.messages.0") }}
      </p>
      <v-chip-group v-model="selected" multiple column filter class="mb-4">
        <v-chip v-for="g in groups" :key="g" :value="g" variant="outlined">
          {{ t(labels[g]) }}
        </v-chip>
      </v-chip-group>
      <v-textarea
        v-model="customIps"
        :label="t('tproxyWhiteIpGroups.formName2')"
        :placeholder="t('tproxyWhiteIpGroups.formPlaceholder2')"
        :error-messages="
          invalid ? [t('tproxyWhiteIpGroups.invalidCustomIps')] : []
        "
        rows="4"
        auto-grow
        dir="ltr"
        @keydown.enter="
          (e: KeyboardEvent) => (e.ctrlKey || e.metaKey) && save()
        "
      />
      <v-alert
        type="info"
        variant="tonal"
        density="compact"
        class="md3-body-small"
      >
        {{ t("tproxyWhiteIpGroups.messages.1") }}
      </v-alert>
    </v-card-text>
    <v-card-actions class="px-6 pb-4">
      <v-spacer />
      <v-btn variant="text" @click="emit('close')">{{
        t("operations.cancel")
      }}</v-btn>
      <v-btn
        variant="flat"
        color="primary"
        :loading="saving"
        :disabled="invalid"
        @click="save"
      >
        {{ t("operations.save") }}
      </v-btn>
    </v-card-actions>
  </v-card>
</template>
