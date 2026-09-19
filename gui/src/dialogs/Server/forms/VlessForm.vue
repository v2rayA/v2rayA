<script setup lang="ts">
import { useI18n } from "vue-i18n";
import type { V2rayModel } from "../models";
import { required } from "./parts/rules";
import TlsFields from "./parts/TlsFields.vue";
import TransportFields from "./parts/TransportFields.vue";

const model = defineModel<V2rayModel>({ required: true });
defineProps<{ readonly?: boolean }>();
const { t } = useI18n();
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
    <v-col cols="12">
      <v-text-field
        v-model="model.id"
        label="ID"
        placeholder="UserID"
        :rules="[required]"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
  </v-row>
  <TransportFields v-model="model" :readonly="readonly" />
  <TlsFields v-model="model" :readonly="readonly" vless />
</template>
