<script setup lang="ts">
import { useI18n } from "vue-i18n";
import type { SsrModel } from "../models";
import { required } from "./parts/rules";

const model = defineModel<SsrModel>({ required: true });
defineProps<{ readonly?: boolean }>();
const { t } = useI18n();
const methods = [
  "aes-128-cfb",
  "aes-192-cfb",
  "aes-256-cfb",
  "aes-128-ctr",
  "aes-192-ctr",
  "aes-256-ctr",
  "aes-128-ofb",
  "aes-192-ofb",
  "aes-256-ofb",
  "des-cfb",
  "bf-cfb",
  "cast5-cfb",
  "rc4-md5",
  "chacha20",
  "chacha20-ietf",
  "salsa20",
  "camellia-128-cfb",
  "camellia-192-cfb",
  "camellia-256-cfb",
  "idea-cfb",
  "rc2-cfb",
  "seed-cfb",
  "none",
];
const protocols = [
  "origin",
  "verify_sha1",
  "auth_sha1_v4",
  "auth_aes128_md5",
  "auth_aes128_sha1",
  "auth_chain_a",
  "auth_chain_b",
];
const obfuscations = [
  "plain",
  "http_simple",
  "http_post",
  "random_head",
  "tls1.2_ticket_auth",
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
    <v-col cols="12" sm="8">
      <v-text-field
        v-model="model.server"
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
        v-model="model.password"
        :label="t('configureServer.password')"
        :rules="[required]"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
    <v-col cols="12" sm="6">
      <v-select
        v-model="model.method"
        :items="methods"
        :label="t('configureServer.method')"
        :rules="[required]"
        :readonly="readonly"
      />
    </v-col>
    <v-col cols="12" sm="6">
      <v-select
        v-model="model.proto"
        :items="protocols"
        :label="t('server.protocol')"
        :rules="[required]"
        :readonly="readonly"
      />
    </v-col>
    <v-expand-transition>
      <v-col v-if="model.proto !== 'origin'" cols="12" sm="6">
        <v-text-field
          v-model="model.protoParam"
          :label="t('configureServer.protocolParam')"
          :placeholder="`(${t('common.optional')})`"
          :readonly="readonly"
          dir="ltr"
        />
      </v-col>
    </v-expand-transition>
    <v-col cols="12" sm="6">
      <v-select
        v-model="model.obfs"
        :items="obfuscations"
        :label="t('configureServer.obfs')"
        :rules="[required]"
        :readonly="readonly"
      />
    </v-col>
    <v-expand-transition>
      <v-col v-if="model.obfs !== 'plain'" cols="12" sm="6">
        <v-text-field
          v-model="model.obfsParam"
          :label="t('configureServer.obfsParam')"
          :placeholder="`(${t('common.optional')})`"
          :readonly="readonly"
          dir="ltr"
        />
      </v-col>
    </v-expand-transition>
  </v-row>
</template>
