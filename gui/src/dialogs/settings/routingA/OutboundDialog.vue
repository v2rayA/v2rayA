<script setup lang="ts">
import { reactive, ref } from "vue";
import { useI18n } from "vue-i18n";
import type { Entry } from "./rules";

defineOptions({ name: "RoutingOutboundDialog" });
const props = defineProps<{
  entry?: Extract<Entry, { kind: "outbound" }>;
  names: string[];
}>();
const emit = defineEmits<{ close: [entry?: Entry] }>();
const { t } = useI18n();
const name = ref(props.entry?.name ?? "");
const fn = ref(props.entry?.fn ?? "socks");
const keys = ["address", "port", "user", "pass"] as const;
const fields = reactive(
  Object.fromEntries(
    keys.map((key) => [
      key,
      props.entry?.args.find((arg) => key in arg)?.[key] ?? "",
    ]),
  ),
) as Record<(typeof keys)[number], string>;
const form = ref<{ validate(): Promise<{ valid: boolean }> }>();
const required = (value: string) =>
  !!value.trim() || t("routingA.form.required");
const validName = (value: string) =>
  (/^[^\s=()#,]+$/.test(value) &&
    (!props.names.includes(value) || value === props.entry?.name)) ||
  t("routingA.form.invalidName");
function argument(value: string) {
  const trimmed = value.trim();
  if (/^"(?:\\.|[^"\\])*"$|^'(?:\\.|[^'\\])*'$/.test(trimmed)) return trimmed;
  return /[,()#\s]/.test(trimmed) ? JSON.stringify(trimmed) : trimmed;
}
async function save() {
  if (!(await form.value?.validate())?.valid) return;
  const args =
    props.entry?.args.filter((arg) => !keys.some((key) => key in arg)) ?? [];
  for (const key of keys)
    if (fields[key].trim()) args.push({ [key]: argument(fields[key]) });
  emit("close", {
    ...props.entry,
    kind: "outbound",
    name: name.value,
    fn: fn.value,
    args,
  });
}
</script>

<template>
  <v-card>
    <v-card-title class="md3-headline-small px-6 pt-6">{{
      t(entry ? "routingA.form.editOutbound" : "routingA.form.addOutbound")
    }}</v-card-title>
    <v-card-text class="px-6">
      <v-form ref="form" @submit.prevent="save">
        <v-text-field
          v-model="name"
          :label="t('routingA.form.name')"
          :rules="[required, validName]"
          dir="ltr"
        />
        <v-select
          v-model="fn"
          :items="[...new Set(['socks', 'http', fn])]"
          :label="t('routingA.form.protocol')"
        />
        <v-text-field
          v-model="fields.address"
          :label="t('routingA.form.address')"
          :rules="[required]"
          dir="ltr"
        />
        <v-text-field
          v-model="fields.port"
          :label="t('routingA.form.port')"
          :rules="[required]"
          dir="ltr"
        />
        <v-text-field
          v-model="fields.user"
          :label="t('routingA.form.user')"
          dir="ltr"
          autocomplete="off"
        />
        <v-text-field
          v-model="fields.pass"
          :label="t('routingA.form.pass')"
          type="password"
          dir="ltr"
          autocomplete="new-password"
        />
      </v-form>
    </v-card-text>
    <v-card-actions class="px-6 pb-4">
      <v-spacer />
      <v-btn variant="text" @click="emit('close')">{{
        t("operations.cancel")
      }}</v-btn>
      <v-btn variant="flat" color="primary" @click="save">{{
        t("operations.save")
      }}</v-btn>
    </v-card-actions>
  </v-card>
</template>
