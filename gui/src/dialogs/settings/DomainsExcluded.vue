<script setup lang="ts">
// The domains sniffing must not rewrite the target of, one per line.
import { onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { getDomainsExcluded, putDomainsExcluded } from "@/api";
import { errorText } from "@/api/errors";
import { useNotify } from "@/composables";

defineOptions({ name: "DomainsExcludedDialog" });
const emit = defineEmits<{ close: [saved?: boolean] }>();
const { t } = useI18n();
const notify = useNotify();
const domains = ref("");
const saving = ref(false);

onMounted(async () => {
  try {
    domains.value = (await getDomainsExcluded()).domains ?? "";
  } catch (err) {
    notify.warning(errorText(err));
  }
});

async function save() {
  saving.value = true;
  try {
    await putDomainsExcluded({ domains: domains.value });
    emit("close", true);
  } catch (err) {
    notify.warning(
      t("domainsExcluded.saveFailed", { message: errorText(err) }),
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
        {{ t("domainsExcluded.title") }}
      </v-card-title>
    </v-card-item>
    <v-card-text class="px-6">
      <p class="md3-body-medium mb-4">{{ t("domainsExcluded.messages.0") }}</p>
      <v-textarea
        v-model="domains"
        :label="t('domainsExcluded.formName')"
        :placeholder="t('domainsExcluded.formPlaceholder')"
        rows="6"
        auto-grow
        dir="ltr"
        class="code"
      />
    </v-card-text>
    <v-card-actions class="px-6 pb-4">
      <v-spacer />
      <v-btn variant="text" @click="emit('close')">{{
        t("operations.cancel")
      }}</v-btn>
      <v-btn variant="flat" color="primary" :loading="saving" @click="save">{{
        t("operations.save")
      }}</v-btn>
    </v-card-actions>
  </v-card>
</template>

<style scoped>
.code :deep(textarea) {
  font-family: ui-monospace, "Cascadia Mono", "Fira Mono", Menlo, monospace;
}
</style>
