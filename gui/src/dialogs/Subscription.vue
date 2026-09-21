<script setup lang="ts">
// Edit a subscription: its address, remarks and whether new nodes are
// connected automatically after an update. Saves with PATCH
// /subscription and resolves true.
import { reactive, ref } from "vue";
import { useI18n } from "vue-i18n";
import { patchSubscription } from "@/api";
import { errorText } from "@/api/errors";
import type { TouchSubscription } from "@/api/types";
import { useNotify } from "@/composables/useNotify";

defineOptions({ name: "SubscriptionDialog" });
const props = defineProps<{ subscription: TouchSubscription }>();
const emit = defineEmits<{ close: [saved?: boolean] }>();
const { t } = useI18n();
const notify = useNotify();

// the old page sent the row with its servers emptied
const form = reactive({ ...props.subscription, servers: [] });
const saving = ref(false);

async function save() {
  saving.value = true;
  try {
    await patchSubscription({ subscription: form });
    notify.success(t("subscription.saved"));
    emit("close", true);
  } catch (err) {
    notify.warning(t("subscription.saveFailed", { message: errorText(err) }));
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <v-card>
    <v-card-item class="px-6 pt-6 pb-2">
      <v-card-title class="md3-headline-small pa-0">
        {{ t("configureSubscription.title") }}
      </v-card-title>
    </v-card-item>
    <v-card-text class="px-6">
      <v-textarea
        v-model="form.address"
        :label="t('subscription.subscription')"
        rows="2"
        auto-grow
        dir="ltr"
        @keydown.enter="
          (e: KeyboardEvent) => (e.ctrlKey || e.metaKey) && save()
        "
      />
      <v-text-field v-model="form.remarks" :label="t('subscription.remarks')" />
      <v-switch
        v-model="form.autoSelect"
        :label="t('subscription.autoSelect')"
        hide-details
      />
    </v-card-text>
    <v-card-actions class="px-6 pb-4">
      <v-spacer />
      <v-btn variant="text" @click="emit('close')">{{
        t("operations.cancel")
      }}</v-btn>
      <v-btn variant="flat" color="primary" :loading="saving" @click="save">
        {{ t("operations.saveApply") }}
      </v-btn>
    </v-card-actions>
  </v-card>
</template>
