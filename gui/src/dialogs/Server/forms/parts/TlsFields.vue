<script setup lang="ts">
// The TLS of a vmess or vless node: off, tls or — for vless on Xray —
// reality; SNI, the uTLS fingerprint, ALPN, the certificate pin and
// name check, and vless's flow and reality keys. Edits the v2ray model
// in place. The obfuscation "dtls" forces TLS, so the choice hides then.
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { useAppStore } from "@/stores/app";
import type { V2rayModel } from "../../models";

const props = defineProps<{
  readonly?: boolean;
  /** vless: flow, and reality when the core is Xray */
  vless?: boolean;
}>();
const model = defineModel<V2rayModel>({ required: true });
const { t } = useI18n();
const store = useAppStore();

// v2raya_core is the merged Xray-based core
const xray = computed(() =>
  ["xray", "v2rayacore"].includes(store.variant.toLowerCase()),
);
const modes = computed(() => [
  { value: "none", title: t("setting.options.off") },
  { value: "tls", title: "tls" },
  ...(props.vless && xray.value
    ? [{ value: "reality", title: "reality" }]
    : []),
]);
const fingerprints = computed(() => [
  { value: "", title: t("common.none") },
  ...[
    "chrome",
    "firefox",
    "safari",
    "ios",
    "android",
    "edge",
    "random",
    "randomized",
  ].map((v) => ({ value: v, title: v })),
]);
const on = computed(() => model.value.tls !== "none");
const secured = computed(
  () =>
    model.value.tls === "tls" || (props.vless && model.value.tls === "reality"),
);
</script>

<template>
  <v-row dense>
    <v-col v-if="model.type !== 'dtls'" cols="12" sm="6">
      <v-select
        v-model="model.tls"
        :items="modes"
        label="TLS"
        :readonly="readonly"
      />
    </v-col>
    <v-col v-if="on" cols="12" sm="6">
      <v-text-field
        v-model="model.sni"
        label="SNI"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
    <v-col v-if="secured" cols="12" sm="6">
      <v-select
        v-model="model.fp"
        :items="fingerprints"
        :label="t('configureServer.utlsFingerprint')"
        :readonly="readonly"
      />
    </v-col>
    <v-col v-if="model.tls === 'tls'" cols="12" sm="6">
      <v-text-field
        v-model="model.alpn"
        label="ALPN"
        placeholder="h3,h2,http/1.1"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
    <v-col v-if="vless && on" cols="12" sm="6">
      <v-text-field
        v-model="model.flow"
        label="Flow"
        :readonly="readonly"
        dir="ltr"
      />
    </v-col>
    <template v-if="vless && model.tls === 'reality'">
      <v-col cols="12" sm="6">
        <v-text-field
          v-model="model.pbk"
          :label="t('configureServer.realityPublicKey')"
          :readonly="readonly"
          dir="ltr"
        />
      </v-col>
      <v-col cols="12" sm="6">
        <v-text-field
          v-model="model.sid"
          :label="t('configureServer.realityShortId')"
          :readonly="readonly"
          dir="ltr"
        />
      </v-col>
      <v-col cols="12" sm="6">
        <v-text-field
          v-model="model.spx"
          :label="t('configureServer.realitySpiderX')"
          :readonly="readonly"
          dir="ltr"
        />
      </v-col>
    </template>
    <template v-if="on">
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
    </template>
  </v-row>
</template>
