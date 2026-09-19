<script setup lang="ts">
// The node editor: a protocol choice, the form of the chosen protocol,
// save as a share link. `which` names the node to edit (its link is
// loaded first); null creates one. `readonly` shows a subscription's node.
// The dialog resolves true after a save.
import { computed, onMounted, ref } from "vue";
import { useI18n } from "vue-i18n";
import type { Which } from "@/api/types";
import { errorText } from "@/api/errors";
import { useNotify } from "@/composables/useNotify";
import {
  createEditor,
  modelKey,
  protocolLabels,
  protocols,
  type Protocol,
} from "./editor";
import VmessForm from "./forms/VmessForm.vue";
import VlessForm from "./forms/VlessForm.vue";
import WireguardForm from "./forms/WireguardForm.vue";
import SsForm from "./forms/SsForm.vue";
import TrojanForm from "./forms/TrojanForm.vue";
import JuicityForm from "./forms/JuicityForm.vue";
import TuicForm from "./forms/TuicForm.vue";
import Hysteria2Form from "./forms/Hysteria2Form.vue";
import HttpForm from "./forms/HttpForm.vue";
import Socks5Form from "./forms/Socks5Form.vue";
import AnytlsForm from "./forms/AnytlsForm.vue";

const props = withDefaults(
  defineProps<{ which?: Which | null; readonly?: boolean }>(),
  { which: null, readonly: false },
);
const emit = defineEmits<{ close: [saved?: boolean] }>();
defineOptions({ name: "ServerDialog" });
const { t } = useI18n();
const notify = useNotify();

// protocol → its form; a form file is one tab and nothing else
const forms: Record<Protocol, unknown> = {
  vmess: VmessForm,
  vless: VlessForm,
  wireguard: WireguardForm,
  ss: SsForm,
  trojan: TrojanForm,
  juicity: JuicityForm,
  tuic: TuicForm,
  hysteria2: Hysteria2Form,
  http: HttpForm,
  socks5: Socks5Form,
  anytls: AnytlsForm,
};

const editor = createEditor(props.which);
const { models, protocol } = editor;
const form = ref<{ validate(): Promise<{ valid: boolean }> } | null>(null);
const loading = ref(props.which !== null);
const saving = ref(false);
const activeForm = computed(() => forms[protocol.value]);
const activeKey = computed(() => modelKey(protocol.value));

onMounted(async () => {
  // a node that could not be loaded must not be saved over from the
  // defaults, so the dialog closes with the error
  try {
    if (!(await editor.load())) {
      notify.error(t("sharing.failed", { message: t("common.fail") }));
      emit("close");
    }
  } catch (err) {
    notify.error(t("sharing.failed", { message: errorText(err) }));
    emit("close");
  } finally {
    loading.value = false;
  }
});

async function save() {
  if (saving.value) return;
  const check = await form.value?.validate();
  if (check && !check.valid) return;
  saving.value = true;
  try {
    await editor.save();
    notify.success(t("server.saved"));
    emit("close", true);
  } catch (err) {
    notify.error(t("server.saveFailed", { message: errorText(err) }));
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <v-card class="server-editor">
    <v-card-item class="px-6 pt-6 pb-2">
      <v-card-title class="md3-headline-small pa-0">
        {{
          t(
            readonly
              ? "configureServer.titleReadonly"
              : "configureServer.title",
          )
        }}
      </v-card-title>
    </v-card-item>
    <v-card-text class="px-6 pb-0">
      <v-select
        v-model="protocol"
        :items="protocols.map((p) => ({ value: p, title: protocolLabels[p] }))"
        :label="t('server.protocol')"
        :readonly="readonly"
        hide-details
      />
    </v-card-text>
    <v-card-text class="px-6 server-editor__body">
      <v-skeleton-loader v-if="loading" type="list-item-two-line@3" />
      <v-form v-else ref="form" :readonly="readonly" @submit.prevent="save">
        <component
          :is="activeForm"
          :key="protocol"
          v-model="models[activeKey]"
          :readonly="readonly"
        />
      </v-form>
    </v-card-text>
    <v-card-actions class="px-6 pb-4">
      <v-spacer />
      <v-btn variant="text" @click="emit('close')">
        {{ t(readonly ? "operations.confirm" : "operations.cancel") }}
      </v-btn>
      <v-btn
        v-if="!readonly"
        variant="flat"
        color="primary"
        :loading="saving"
        :disabled="loading"
        @click="save"
      >
        {{ t("operations.saveApply") }}
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<style scoped>
.server-editor__body {
  max-height: min(70vh, 640px);
  overflow-y: auto;
}
</style>
