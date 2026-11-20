<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { getSetting } from "@/api";
import { useI18n } from "vue-i18n";
import { openDialog } from "@/composables/useDialog";
import HwidWarning from "./HwidWarning.vue";

const props = defineProps<{ address?: string; disabled?: boolean }>();
const model = defineModel<boolean>({ required: true });
const { t } = useI18n();
const inherited = ref(false);
const ready = ref(props.address === undefined);
onMounted(async () => {
  if (ready.value) return;
  try {
    inherited.value = (await getSetting()).setting.sendHWID === true;
    ready.value = true;
  } catch {
    // Keep the switch unavailable until the effective permission is known.
  }
});
const host = computed(() => {
  if (!props.address) return "";
  try {
    const u = new URL(props.address);
    return u.protocol === "https:" ? u.host : "";
  } catch {
    return "";
  }
});
async function change(event: MouseEvent) {
  if (!(event.target instanceof HTMLInputElement)) return;
  // Defer the native checkbox change until the user's decision is known.
  event.preventDefault();
  const enabled = !model.value;
  const address = props.address;
  if (
    !enabled ||
    ((await openDialog<boolean>(
      HwidWarning,
      { host: host.value },
      { width: 480 },
    ).result) &&
      address === props.address)
  )
    model.value = !!enabled;
}
</script>

<template>
  <v-switch
    :disabled="
      disabled || !ready || (address !== undefined && (!host || inherited))
    "
    :model-value="
      address !== undefined ? !!host && (inherited || model) : model
    "
    :label="
      address === undefined
        ? t('subscription.hwidAll')
        : inherited && host
          ? t('subscription.hwidInherited')
          : t('subscription.hwidOne')
    "
    hide-details
    @click="change"
  />
</template>
