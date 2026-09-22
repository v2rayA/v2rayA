<script setup lang="ts">
// Update the GFWList (geosite.dat) from GitHub or a custom link, or
// delete the local copy. Resolves true when the local file changed.
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import { deleteGfwList, putGfwList } from "@/api";
import { errorText } from "@/api/errors";
import { useLoading, useNotify } from "@/composables";

defineOptions({ name: "GfwListDialog" });
const props = defineProps<{ localVersion: string }>();
const emit = defineEmits<{ close: [changed?: boolean] }>();
const { t } = useI18n();
const notify = useNotify();
const loading = useLoading();
const link = ref("");
const busy = ref(false);

async function update() {
  if (busy.value) return;
  if (link.value && !link.value.startsWith("http")) {
    notify.warning(t("gfwList.wrongCustomLink"));
    return;
  }
  busy.value = true;
  const overlay = loading.open();
  try {
    const res = await putGfwList({ downloadLink: link.value });
    // "already the latest" is a result, not a failure
    if (res.alreadyUpToDate)
      notify.info(
        t("gfwList.alreadyUpToDate", { version: res.localGFWListVersion }),
      );
    else notify.success(t("gfwList.updated"));
    emit("close", true);
  } catch (err) {
    notify.warning(t("gfwList.saveFailed", { message: errorText(err) }));
  } finally {
    overlay.close();
    busy.value = false;
  }
}

async function remove() {
  try {
    await deleteGfwList();
    emit("close", true);
  } catch (err) {
    notify.warning(t("delete.failed", { message: errorText(err) }));
  }
}
</script>

<template>
  <v-card>
    <v-card-item class="px-6 pt-6 pb-2">
      <v-card-title class="md3-headline-small pa-0">{{
        t("gfwList.title")
      }}</v-card-title>
    </v-card-item>
    <v-card-text class="px-6">
      <p class="md3-body-medium mb-2">{{ t("gfwList.messages.0") }}</p>
      <p class="md3-body-medium mb-4">{{ t("gfwList.messages.1") }}</p>
      <v-text-field
        v-model="link"
        :label="t('gfwList.formName')"
        placeholder="https://example.com/LoyalsoldierSite.dat"
        dir="ltr"
        @keydown.enter="update"
      />
      <v-alert
        type="warning"
        variant="tonal"
        density="compact"
        class="md3-body-small"
      >
        {{ t("gfwList.messages.2") }}
      </v-alert>
    </v-card-text>
    <v-card-actions class="px-6 pb-4">
      <v-btn
        variant="text"
        color="error"
        :disabled="!props.localVersion"
        @click="remove"
      >
        {{ t("operations.delete") }}
      </v-btn>
      <v-spacer />
      <v-btn variant="text" @click="emit('close')">{{
        t("operations.cancel")
      }}</v-btn>
      <v-btn variant="flat" color="primary" :loading="busy" @click="update">{{
        t("operations.update")
      }}</v-btn>
    </v-card-actions>
  </v-card>
</template>
