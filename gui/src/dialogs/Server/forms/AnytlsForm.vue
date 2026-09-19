<script setup lang="ts">
import { useI18n } from "vue-i18n";
import type { AnytlsModel } from "../models";
import { required } from "./parts/rules";

const model = defineModel<AnytlsModel>({ required: true });
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
        v-model="model.auth"
        :label="t('configureServer.auth')"
        :placeholder="t('configureServer.authKey')"
        :rules="[required]"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
    <v-col cols="12" sm="6">
      <v-text-field
        v-model="model.sni"
        label="SNI(Peer)"
        placeholder="SNI / Peer"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
    <v-col cols="12" sm="6">
      <v-text-field
        v-model="model.pinnedPeerCertSha256"
        :label="t('pinnedPeerCertSha256')"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
    <v-col cols="12" sm="6">
      <v-text-field
        v-model="model.verifyPeerCertByName"
        :label="t('verifyPeerCertByName')"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
    <v-col cols="12" sm="6">
      <v-switch
        v-model="model.allowInsecure"
        :label="t('configureServer.allowInsecure')"
        hide-details
        :readonly="readonly"
      />
    </v-col>
    <v-col cols="12" sm="6">
      <v-text-field
        v-model="model.minIdleSession"
        type="number"
        :label="t('configureServer.minIdleSession')"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
  </v-row>
</template>
