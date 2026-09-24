<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import type { SsModel } from "../models";
import { required } from "./parts/rules";

const model = defineModel<SsModel>({ required: true });
defineProps<{ readonly?: boolean }>();
const { t } = useI18n();
const methods = [
  "2022-blake3-aes-128-gcm",
  "2022-blake3-aes-256-gcm",
  "2022-blake3-chacha20-poly1305",
  "aes-128-gcm",
  "aes-256-gcm",
  "chacha20-poly1305",
  "chacha20-ietf-poly1305",
  "xchacha20-ietf-poly1305",
];
const plugins = computed(() => [
  { value: "", title: t("setting.options.off") },
  { value: "simple-obfs", title: "simple-obfs" },
  { value: "v2ray-plugin", title: "v2ray-plugin" },
]);
const implementations = computed(() => [
  { value: "", title: t("setting.options.default") },
  { value: "chained", title: "chained" },
  { value: "transport", title: "transport" },
]);
const tlsModes = computed(() => [
  { value: "", title: t("setting.options.off") },
  { value: "tls", title: "tls" },
]);
const hasPlugin = computed(() =>
  ["simple-obfs", "v2ray-plugin"].includes(model.value.plugin),
);
const showsHost = computed(
  () =>
    (model.value.plugin === "simple-obfs" &&
      ["http", "tls"].includes(model.value.obfs)) ||
    model.value.plugin === "v2ray-plugin",
);
const showsPath = computed(
  () =>
    (model.value.plugin === "simple-obfs" && model.value.obfs === "http") ||
    model.value.plugin === "v2ray-plugin",
);
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
        v-model="model.plugin"
        :items="plugins"
        :label="t('configureServer.plugin')"
        :readonly="readonly"
      />
    </v-col>
    <v-expand-transition>
      <v-col v-if="hasPlugin" cols="12" sm="6">
        <v-select
          v-model="model.impl"
          :items="implementations"
          :label="t('configureServer.pluginImpl')"
          :readonly="readonly"
        />
        <v-alert type="info" variant="tonal" density="compact" class="mb-4">
          {{ t("setting.messages.ssPluginImpl") }}
        </v-alert>
      </v-col>
    </v-expand-transition>
    <v-expand-transition>
      <v-col v-if="model.plugin === 'simple-obfs'" cols="12" sm="6">
        <v-select
          v-model="model.obfs"
          :items="['http', 'tls']"
          :label="t('configureServer.obfs')"
          :readonly="readonly"
        />
      </v-col>
    </v-expand-transition>
    <v-expand-transition>
      <v-col v-if="model.plugin === 'v2ray-plugin'" cols="12" sm="6">
        <v-select
          v-model="model.mode"
          :items="['websocket']"
          :label="t('configureServer.mode')"
          :readonly="readonly"
        />
      </v-col>
    </v-expand-transition>
    <v-expand-transition>
      <v-col v-if="model.plugin === 'v2ray-plugin'" cols="12" sm="6">
        <v-select
          v-model="model.tls"
          :items="tlsModes"
          label="TLS"
          :readonly="readonly"
        />
      </v-col>
    </v-expand-transition>
    <v-expand-transition>
      <v-col v-if="showsHost" cols="12" sm="6">
        <v-text-field
          v-model="model.host"
          :label="t('configureServer.host')"
          :placeholder="`(${t('common.optional')})`"
          :readonly="readonly"
          dir="ltr"
        />
      </v-col>
    </v-expand-transition>
    <v-expand-transition>
      <v-col v-if="showsPath" cols="12" sm="6">
        <v-text-field
          v-model="model.path"
          :label="t('configureServer.pathObfuscation')"
          placeholder="/"
          :readonly="readonly"
          dir="ltr"
        />
      </v-col>
    </v-expand-transition>
  </v-row>
</template>
