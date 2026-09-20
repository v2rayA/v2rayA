<script setup lang="ts">
// Import a node link, several (one per line), or a subscription address;
// a QR code image can supply the link. Resolves true after an import.
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import { Decoder } from "@nuintun/qrcode";
import { mdiQrcodeScan } from "@mdi/js";
import { postImport } from "@/api";
import { ApiError } from "@/api/client";
import { errorText } from "@/api/errors";
import { useNotify } from "@/composables/useNotify";

defineOptions({ name: "ImportDialog" });
const props = defineProps<{
  /** what the dialog opens on; server links unless told otherwise */
  kind?: "server" | "subscription";
}>();
const emit = defineEmits<{ close: [imported?: boolean] }>();
const { t } = useI18n();
const notify = useNotify();

const kind = ref<"server" | "subscription">(props.kind ?? "server");
const text = ref("");
const importing = ref(false);
const picker = ref<HTMLInputElement | null>(null);

async function submit(url = text.value.trim()) {
  if (!url || importing.value) return;
  importing.value = true;
  try {
    await postImport({ url, kind: kind.value });
    notify.success(t("import.success"));
    emit("close", true);
  } catch (err) {
    if (err instanceof ApiError && err.kind === "timeout")
      notify.warning(t("import.timeout"));
    else notify.warning(t("import.failed", { message: errorText(err) }));
  } finally {
    importing.value = false;
  }
}

function onImage(event: Event) {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  input.value = "";
  if (!file) return;
  if (!file.type.startsWith("image/")) {
    notify.warning(t("import.notImage"));
    return;
  }
  const reader = new FileReader();
  reader.onload = async () => {
    try {
      const result = await new Decoder().scan(reader.result as string);
      text.value = result.data;
      await submit(result.data);
    } catch {
      notify.warning(t("import.qrcodeError"));
    }
  };
  reader.readAsDataURL(file);
}
</script>

<template>
  <v-card>
    <v-card-item class="px-6 pt-6 pb-2">
      <v-card-title class="md3-headline-small pa-0">
        {{ t("operations.import") }}
      </v-card-title>
    </v-card-item>
    <v-card-text class="px-6">
      <v-btn-toggle
        v-model="kind"
        mandatory
        divided
        variant="outlined"
        rounded="xl"
        density="comfortable"
        selected-class="bg-secondary-container text-on-secondary-container"
        class="w-100 mb-5"
      >
        <v-btn value="server" class="flex-grow-1 text-none">{{
          t("import.server")
        }}</v-btn>
        <v-btn value="subscription" class="flex-grow-1 text-none">
          {{ t("import.subscription") }}
        </v-btn>
      </v-btn-toggle>
      <v-textarea
        v-model="text"
        :label="
          kind === 'server'
            ? t('import.batchMessage')
            : t('import.subscriptionMessage')
        "
        rows="3"
        auto-grow
        autofocus
        dir="ltr"
        :append-inner-icon="mdiQrcodeScan"
        @click:append-inner="picker?.click()"
        @keydown.ctrl.enter.prevent="submit()"
      />
      <input
        ref="picker"
        type="file"
        accept="image/*"
        class="d-none"
        :aria-label="t('operations.import')"
        @change="onImage"
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
        :loading="importing"
        :disabled="!text.trim()"
        @click="submit()"
      >
        {{ t("operations.import") }}
      </v-btn>
    </v-card-actions>
  </v-card>
</template>
