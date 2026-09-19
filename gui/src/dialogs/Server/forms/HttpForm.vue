<script setup lang="ts">
import { useI18n } from "vue-i18n";
import type { HttpModel } from "../models";
import { required } from "./parts/rules";

const model = defineModel<HttpModel>({ required: true });
defineProps<{ readonly?: boolean }>();
const { t } = useI18n();
const protocols = [
  { value: "http", title: "HTTP" },
  { value: "https", title: "HTTPS" },
];
</script>

<template>
  <v-row dense>
    <v-col cols="12">
      <v-text-field
        v-model="model.name"
        :label="t('configureServer.servername')"
        :readonly="readonly"
      />
    </v-col>
    <v-col cols="12" sm="6">
      <v-select
        v-model="model.protocol"
        :items="protocols"
        :label="t('server.protocol')"
        :readonly="readonly"
      />
    </v-col>
    <v-col cols="12" sm="8">
      <v-text-field
        v-model="model.host"
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
    <v-col cols="12" sm="6">
      <v-text-field
        v-model="model.username"
        :label="t('configureServer.username')"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
    <v-col cols="12" sm="6">
      <v-text-field
        v-model="model.password"
        :label="t('configureServer.password')"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
  </v-row>
</template>
