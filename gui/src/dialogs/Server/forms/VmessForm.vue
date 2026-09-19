<script setup lang="ts">
// The vmess tab of the node editor: the node's identity, then its
// transport and TLS (parts/ shared with vless). The model comes in by
// v-model and is edited in place; `readonly` shows a subscription node's
// values; a field hidden by its condition is not validated.
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import type { V2rayModel } from "../models";
import { required } from "./parts/rules";
import TlsFields from "./parts/TlsFields.vue";
import TransportFields from "./parts/TransportFields.vue";

defineProps<{ readonly?: boolean }>();
const model = defineModel<V2rayModel>({ required: true });
const { t } = useI18n();

const securities = computed(() => [
  { value: "auto", title: t("configureServer.auto") },
  ...[
    "aes-256-gcm",
    "aes-128-gcm",
    "chacha20-poly1305",
    "xchacha20-poly1305",
    "none",
    "zero",
  ].map((v) => ({ value: v, title: v })),
]);
</script>

<template>
  <v-row dense>
    <v-col cols="12">
      <v-text-field
        v-model="model.ps"
        :label="t('configureServer.servername')"
        :readonly="readonly"
      />
    </v-col>
    <v-col cols="12" sm="8">
      <v-text-field
        v-model="model.add"
        :label="t('configureServer.host')"
        :rules="[required]"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
    <v-col cols="12" sm="4">
      <v-text-field
        v-model="model.port"
        type="number"
        :label="t('configureServer.port')"
        :rules="[required]"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
    <v-col cols="12" sm="8">
      <v-text-field
        v-model="model.id"
        label="ID"
        placeholder="UserID"
        :rules="[required]"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
    <v-col cols="12" sm="4">
      <v-text-field
        v-model="model.aid"
        type="number"
        min="0"
        max="65535"
        label="AlterID"
        :readonly="readonly"
      />
    </v-col>
    <v-col cols="12" sm="6">
      <v-select
        v-model="model.scy"
        :items="securities"
        :label="t('configureServer.security')"
        :readonly="readonly"
      />
    </v-col>
  </v-row>
  <TransportFields v-model="model" :readonly="readonly" />
  <TlsFields v-model="model" :readonly="readonly" />
</template>
