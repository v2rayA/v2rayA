<script setup lang="ts">
import { useI18n } from "vue-i18n";
import type { Hysteria2Model } from "../models";
import { required } from "./parts/rules";

const model = defineModel<Hysteria2Model>({ required: true });
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
        v-model="model.obfs"
        :items="['none', 'salamander']"
        :label="t('configureServer.obfs')"
        :rules="[required]"
        :readonly="readonly"
      />
    </v-col>
    <v-expand-transition>
      <v-col v-if="model.obfs !== 'none'" cols="12" sm="6">
        <v-text-field
          v-model="model.obfsPassword"
          :label="t('configureServer.obfsPassword')"
          :readonly="readonly"
          dir="ltr"
        />
      </v-col>
    </v-expand-transition>
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
  </v-row>
</template>
