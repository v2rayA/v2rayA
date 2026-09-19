<script setup lang="ts">
// Login, or the first account's registration when the backend has none.
// Opened persistent by the session starter; it closes itself after the
// token arrives and starts the session on it.
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import { mdiEye, mdiEyeOff } from "@mdi/js";
import { postAccount, postLogin } from "@/api";
import { errorText } from "@/api/errors";
import { useNotify } from "@/composables/useNotify";
import { useDialog } from "@/composables/useDialog";
import { resetSession } from "@/session";
import PortsDialog from "@/dialogs/settings/Ports.vue";
import logo from "@/assets/img/v2raya-icon.svg";

const props = defineProps<{ first: boolean }>();
const emit = defineEmits<{ close: [] }>();
defineOptions({ name: "LoginDialog" });
const { t } = useI18n();
const notify = useNotify();

const username = ref("");
const password = ref("");
const showPassword = ref(false);
const submitting = ref(false);

async function submit() {
  if (submitting.value) return;
  submitting.value = true;
  try {
    const body = { username: username.value, password: password.value };
    const { token } = props.first
      ? await postAccount(body)
      : await postLogin(body);
    emit("close");
    await resetSession({ token });
  } catch (err) {
    notify.error(
      t(props.first ? "register.failed" : "login.failed", {
        message: errorText(err),
      }),
    );
  } finally {
    submitting.value = false;
  }
}

function openAddress() {
  useDialog().open(PortsDialog, {}, { width: 520 });
}
</script>

<template>
  <v-card class="login">
    <v-card-item class="pt-6 px-6">
      <div class="d-flex flex-column align-center ga-3">
        <img :src="logo" alt="" class="login__logo" />
        <v-card-title class="md3-headline-small pa-0 text-center text-wrap">
          {{ first ? t("register.title") : t("login.title") }}
        </v-card-title>
      </div>
    </v-card-item>
    <v-card-text class="px-6 pt-4">
      <v-text-field
        v-model="username"
        :label="t('login.username')"
        name="username"
        autocomplete="username"
        autofocus
        @keydown.enter.prevent="submit"
      />
      <v-text-field
        v-model="password"
        :label="t('login.password')"
        :type="showPassword ? 'text' : 'password'"
        :maxlength="first ? 32 : undefined"
        :append-inner-icon="showPassword ? mdiEyeOff : mdiEye"
        name="password"
        :autocomplete="first ? 'new-password' : 'current-password'"
        @click:append-inner="showPassword = !showPassword"
        @keydown.enter.prevent="submit"
      />
      <v-alert
        v-if="first"
        type="info"
        variant="tonal"
        density="compact"
        class="md3-body-small"
      >
        <p>{{ t("register.messages.0") }}</p>
        <p>{{ t("register.messages.1") }}</p>
        <p>{{ t("register.messages.2") }}</p>
      </v-alert>
    </v-card-text>
    <v-card-actions class="px-6 pb-4">
      <v-btn variant="text" @click="openAddress">
        {{ t("customAddressPort.serviceAddress") }}
      </v-btn>
      <v-spacer />
      <v-btn
        variant="flat"
        color="primary"
        :loading="submitting"
        :disabled="submitting"
        @click="submit"
      >
        {{ first ? t("operations.create") : t("operations.login") }}
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<style scoped>
.login__logo {
  width: 64px;
  height: 64px;
}
</style>
