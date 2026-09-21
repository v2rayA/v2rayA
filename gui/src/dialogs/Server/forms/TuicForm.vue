<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import type { TuicModel } from "../models";
import { required } from "./parts/rules";

const model = defineModel<TuicModel>({ required: true });
defineProps<{ readonly?: boolean }>();
const { t } = useI18n();

const disableSniOptions = computed(() => [
  { value: false, title: t("operations.no") },
  { value: true, title: t("operations.yes") },
]);
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
    <v-col cols="12" sm="6">
      <v-select
        v-model="model.udpRelayMode"
        :items="['native', 'quic']"
        :label="t('configureServer.udpRelayMode')"
        :rules="[required]"
        :readonly="readonly"
      />
    </v-col>
  </v-row>
  <v-row dense>
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
      <v-select
        v-model="model.disableSni"
        :items="disableSniOptions"
        :label="t('configureServer.disableSni')"
        :rules="[required]"
        :readonly="readonly"
      />
    </v-col>
    <v-expand-transition>
      <v-col v-if="model.disableSni === false" cols="12" sm="6">
        <v-text-field
          v-model="model.sni"
          label="SNI"
          :readonly="readonly"
          dir="ltr"
        />
      </v-col>
    </v-expand-transition>
    <v-col cols="12" sm="6">
      <v-text-field
        v-model="model.alpn"
        label="ALPN"
        placeholder="h3"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
  </v-row>
</template>
