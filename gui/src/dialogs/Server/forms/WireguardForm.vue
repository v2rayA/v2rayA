<script setup lang="ts">
import { useI18n } from "vue-i18n";
import type { WireguardModel } from "../models";
import { required } from "./parts/rules";

const model = defineModel<WireguardModel>({ required: true });
defineProps<{ readonly?: boolean }>();
const { t } = useI18n();
</script>

<template>
  <v-row dense>
    <v-col cols="12">
      <v-text-field
        v-model="model.name"
        :label="t('configureServer.wireguardName')"
        :placeholder="t('configureServer.servername')"
        :readonly="readonly"
      />
    </v-col>
    <v-col cols="12" sm="8">
      <v-text-field
        v-model="model.address"
        :label="t('configureServer.wireguardAddress')"
        :placeholder="t('configureServer.host')"
        :rules="[required]"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
    <v-col cols="12" sm="4">
      <v-text-field
        v-model="model.port"
        type="number"
        :label="t('configureServer.wireguardPort')"
        :rules="[required]"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
    <v-col cols="12" sm="6">
      <v-text-field
        v-model="model.publicKey"
        :label="t('configureServer.wireguardPublicKey')"
        :rules="[required]"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
    <v-col cols="12" sm="6">
      <v-text-field
        v-model="model.privateKey"
        :label="t('configureServer.wireguardPrivateKey')"
        :rules="[required]"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
    <v-col cols="12" sm="6">
      <v-text-field
        v-model="model.localAddress"
        :label="t('configureServer.wireguardLocalAddress')"
        :placeholder="t('configureServer.wireguardLocalAddressPlaceholder')"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
    <v-col cols="12" sm="6">
      <v-text-field
        v-model="model.mtu"
        type="number"
        :label="t('configureServer.wireguardMtu')"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
    <v-col cols="12" sm="6">
      <v-text-field
        v-model="model.allowedIPs"
        :label="t('configureServer.wireguardAllowedIPs')"
        placeholder="0.0.0.0/0, ::/0"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
    <v-col cols="12" sm="6">
      <v-text-field
        v-model="model.persistentKeepalive"
        type="number"
        :label="t('configureServer.wireguardPersistentKeepalive')"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
    <v-col cols="12" sm="6">
      <v-text-field
        v-model="model.preSharedKey"
        :label="t('configureServer.wireguardPreSharedKey')"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
    <v-col cols="12" sm="6">
      <v-text-field
        v-model="model.reserved"
        :label="t('configureServer.wireguardReserved')"
        placeholder="0,0,0"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
    <v-col cols="12" sm="6">
      <v-text-field
        v-model="model.workers"
        type="number"
        :label="t('configureServer.wireguardWorkers')"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
    <v-col cols="12">
      <v-switch
        v-model="model.kernelMode"
        :label="t('configureServer.wireguardKernelMode')"
        :readonly="readonly"
        color="primary"
        hide-details
      />
    </v-col>
  </v-row>
</template>
