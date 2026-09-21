<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import type { TrojanModel } from "../models";
import { required } from "./parts/rules";

const model = defineModel<TrojanModel>({ required: true });
defineProps<{ readonly?: boolean }>();
const { t } = useI18n();

const methods = computed(() => [
  { value: "origin", title: t("configureServer.origin") },
  { value: "shadowsocks", title: "shadowsocks" },
]);
const ciphers = [
  "aes-128-gcm",
  "aes-256-gcm",
  "chacha20-poly1305",
  "chacha20-ietf-poly1305",
];
const networks = [
  { value: "tcp", title: "TCP" },
  { value: "kcp", title: "mKCP" },
  { value: "ws", title: "WebSocket" },
  { value: "h2", title: "HTTP/2" },
  { value: "grpc", title: "gRPC" },
];
const obfuscations = computed(() => [
  { value: "none", title: t("configureServer.noObfuscation") },
  { value: "websocket", title: "websocket" },
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
        v-model="model.method"
        :items="methods"
        :label="t('server.protocol')"
        :rules="[required]"
        :readonly="readonly"
      />
    </v-col>
    <v-expand-transition>
      <v-col v-if="model.method === 'shadowsocks'" cols="12" sm="6">
        <v-select
          v-model="model.ssCipher"
          :items="ciphers"
          :label="t('configureServer.ssCipher')"
          :rules="[required]"
          :readonly="readonly"
        />
      </v-col>
    </v-expand-transition>
    <v-expand-transition>
      <v-col v-if="model.method === 'shadowsocks'" cols="12" sm="6">
        <v-text-field
          v-model="model.ssPassword"
          :label="t('configureServer.ssPassword')"
          :rules="[required]"
          :readonly="readonly"
          dir="ltr"
        />
      </v-col>
    </v-expand-transition>
    <v-col cols="12" sm="6">
      <v-select
        v-model="model.net"
        :items="networks"
        :label="t('configureServer.network')"
        :rules="[required]"
        :readonly="readonly"
      />
    </v-col>
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
      <v-col v-if="model.obfs === 'websocket'" cols="12" sm="6">
        <v-text-field
          v-model="model.host"
          :label="t('configureServer.websocketHost')"
          :readonly="readonly"
          dir="ltr"
        />
      </v-col>
    </v-expand-transition>
    <v-expand-transition>
      <v-col v-if="model.obfs === 'websocket'" cols="12" sm="6">
        <v-text-field
          v-model="model.path"
          :label="t('configureServer.websocketPath')"
          placeholder="/"
          :readonly="readonly"
          dir="ltr"
        />
      </v-col>
    </v-expand-transition>
    <v-expand-transition>
      <v-col v-if="model.net === 'ws' || model.net === 'h2'" cols="12" sm="6">
        <v-text-field
          v-model="model.host"
          :label="t('configureServer.hostObfuscation')"
          :readonly="readonly"
          dir="ltr"
        />
      </v-col>
    </v-expand-transition>
    <v-expand-transition>
      <v-col v-if="model.net === 'ws' || model.net === 'h2'" cols="12" sm="6">
        <v-text-field
          v-model="model.path"
          :label="t('configureServer.pathObfuscation')"
          :readonly="readonly"
          dir="ltr"
        />
      </v-col>
    </v-expand-transition>
    <v-expand-transition>
      <v-col
        v-if="model.net === 'mkcp' || model.net === 'kcp'"
        cols="12"
        sm="6"
      >
        <v-text-field
          v-model="model.path"
          :label="t('configureServer.seedObfuscation')"
          :readonly="readonly"
          dir="ltr"
        />
      </v-col>
    </v-expand-transition>
    <v-expand-transition>
      <v-col v-if="model.net === 'grpc'" cols="12" sm="6">
        <v-text-field
          v-model="model.path"
          :label="t('configureServer.serviceName')"
          :readonly="readonly"
          dir="ltr"
        />
      </v-col>
    </v-expand-transition>
  </v-row>
  <v-row dense>
    <v-col cols="12" sm="6">
      <v-text-field
        v-model="model.peer"
        label="SNI(Peer)"
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
