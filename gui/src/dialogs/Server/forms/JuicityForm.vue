<script setup lang="ts">
import { useI18n } from "vue-i18n";
import type { JuicityModel } from "../models";
import { required } from "./parts/rules";

const model = defineModel<JuicityModel>({ required: true });
defineProps<{ readonly?: boolean }>();
const { t } = useI18n();
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
      <v-text-field
        v-model="model.server"
        :label="t('configureServer.host')"
        :rules="[required]"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
    <v-col cols="12" sm="6">
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
        v-model="model.uuid"
        label="UUID"
        :rules="[required]"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
    <v-col cols="12" sm="6">
      <v-text-field
        v-model="model.password"
        :label="t('configureServer.password')"
        :rules="[required]"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
  </v-row>
  <v-row dense>
    <v-col cols="12" sm="6">
      <v-select
        v-model="model.cc"
        :items="['bbr']"
        :label="t('configureServer.congestionControl')"
        :rules="[required]"
        :readonly="readonly"
      />
    </v-col>
  </v-row>
  <v-row dense>
    <v-col cols="12" sm="6">
      <v-text-field
        v-model="model.sni"
        label="SNI"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
    <v-col cols="12" sm="6">
      <v-text-field
        v-model="model.pinnedCertchainSha256"
        :label="t('configureServer.pinnedCertchainSha256')"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
    <v-col cols="12" sm="6">
      <v-switch
        v-model="model.allowInsecure"
        :label="t('configureServer.allowInsecure')"
        :readonly="readonly"
      />
    </v-col>
  </v-row>
</template>
