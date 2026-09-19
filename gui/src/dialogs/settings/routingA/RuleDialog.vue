<script setup lang="ts">
import { computed, ref } from "vue";
import { useI18n } from "vue-i18n";
import { mdiDeleteOutline } from "@mdi/js";
import { splitOutside, type Entry } from "./rules";

defineOptions({ name: "RoutingRuleDialog" });
const props = defineProps<{
  entry?: Extract<Entry, { kind: "rule" }>;
  outbounds: string[];
}>();
const emit = defineEmits<{ close: [entry?: Entry] }>();
const { t } = useI18n();
const conditions = ref(
  props.entry?.conditions.map((call) => ({
    fn: call.fn,
    args: call.args.join(", "),
  })) ?? [{ fn: "domain", args: "" }],
);
const outbound = ref(props.entry?.outbound ?? "proxy");
const form = ref<{ validate(): Promise<{ valid: boolean }> }>();
const functions = [
  "domain",
  "ip",
  "port",
  "network",
  "protocol",
  "source",
  "extern",
];
const hints: Record<string, string> = {
  domain: "geosite: cn, domain: example.com",
  ip: "geoip: private, 10.0.0.0/8",
  port: "80, 443, 1000-2000",
  network: "tcp, udp",
  protocol: "http, tls, bittorrent",
  source: "192.168.1.0/24",
  extern: "domain, geosite, cn",
};
const choices = computed(() => [
  ...new Set([...functions, ...conditions.value.map((call) => call.fn)]),
]);
const required = (value: string) =>
  !!value.trim() || t("routingA.form.required");
const validArgs = (value: string) => {
  const args = splitOutside(value, ",");
  return (
    (!!args?.length && args.every((arg) => !!arg)) ||
    t("routingA.form.invalidArguments")
  );
};
async function save() {
  if (!(await form.value?.validate())?.valid) return;
  emit("close", {
    ...props.entry,
    kind: "rule",
    conditions: conditions.value.map((call) => ({
      fn: call.fn,
      args: splitOutside(call.args, ",")!,
    })),
    outbound: outbound.value,
  });
}
</script>

<template>
  <v-card>
    <v-card-title class="md3-headline-small px-6 pt-6">{{
      t(entry ? "routingA.form.editRule" : "routingA.form.addRule")
    }}</v-card-title>
    <v-card-text class="px-6">
      <v-form ref="form" @submit.prevent="save">
        <div
          v-for="(condition, index) in conditions"
          :key="index"
          class="rule-condition mb-4"
        >
          <v-select
            v-model="condition.fn"
            :items="choices"
            :label="t('routingA.form.condition')"
            hide-details
          />
          <v-text-field
            v-model="condition.args"
            :label="t('routingA.form.arguments')"
            :placeholder="hints[condition.fn]"
            :rules="[validArgs]"
            dir="ltr"
            class="rule-condition__arguments"
          />
          <v-btn
            :icon="mdiDeleteOutline"
            size="40"
            variant="text"
            :aria-label="t('operations.delete')"
            :disabled="conditions.length === 1"
            @click="conditions.splice(index, 1)"
          />
        </div>
        <v-btn
          variant="text"
          class="mb-4"
          @click="conditions.push({ fn: 'domain', args: '' })"
          >{{ t("routingA.form.addCondition") }}</v-btn
        >
        <v-select
          v-model="outbound"
          :items="[...new Set([...outbounds, outbound])]"
          :label="t('routingA.form.outbound')"
          :rules="[required]"
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

<style scoped>
.rule-condition {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 40px;
  gap: 8px;
}
.rule-condition__arguments {
  grid-column: 1;
  grid-row: 2;
}
</style>
