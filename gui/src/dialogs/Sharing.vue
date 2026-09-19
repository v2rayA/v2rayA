<script setup lang="ts">
// A node's or a subscription's share link, as a QR code and as text with
// a copy button. A subscription's code carries sub://<base64 address>.
import { onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import { Base64 } from "js-base64";
import QRCode from "qrcode";
import { mdiContentCopy } from "@mdi/js";
import type { TouchType } from "@/api/types";
import { useNotify } from "@/composables/useNotify";
import { copyText } from "@/lib/clipboard";

defineOptions({ name: "SharingDialog" });
const props = defineProps<{
  title: string;
  link: string;
  /** what the code is for: the name of the node or the subscription's host */
  name: string;
  type: TouchType;
}>();
const emit = defineEmits<{ close: [] }>();
const { t } = useI18n();
const notify = useNotify();
const canvas = ref<HTMLCanvasElement | null>(null);

onMounted(() => {
  const data =
    props.type === "subscription"
      ? "sub://" + Base64.encode(props.link)
      : props.link;
  if (canvas.value)
    QRCode.toCanvas(
      canvas.value,
      data,
      { errorCorrectionLevel: "H", width: 240 },
      () => {},
    );
});

async function copy() {
  try {
    await copyText(props.link);
    notify.success(t("sharing.copied"));
  } catch {
    notify.warning(t("sharing.copyFailed"));
  }
}
</script>

<template>
  <v-card>
    <v-card-item class="px-6 pt-6 pb-2">
      <v-card-title class="md3-headline-small pa-0">{{ title }}</v-card-title>
      <v-card-subtitle class="pa-0">{{ name }}</v-card-subtitle>
    </v-card-item>
    <v-card-text class="px-6">
      <v-sheet
        color="surface-container-lowest"
        rounded="lg"
        class="qr mx-auto mb-4"
      >
        <canvas ref="canvas" />
      </v-sheet>
      <v-text-field
        :model-value="link"
        readonly
        dir="ltr"
        :label="t('operations.copyLink')"
        :append-inner-icon="mdiContentCopy"
        @click:append-inner="copy"
      />
    </v-card-text>
    <v-card-actions class="px-6 pb-4">
      <v-spacer />
      <v-btn variant="text" @click="emit('close')">{{
        t("operations.confirm")
      }}</v-btn>
    </v-card-actions>
  </v-card>
</template>

<style scoped>
.qr {
  width: 256px;
  padding: 8px;
  display: flex;
}
.qr canvas {
  width: 240px !important;
  height: 240px !important;
}
</style>
