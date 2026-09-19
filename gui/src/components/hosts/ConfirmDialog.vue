<script setup lang="ts">
// The confirm and prompt dialogs behind useConfirm and usePrompt: a title,
// a message, an optional text field with a validator, and two buttons.
// Resolved through the dialog handle: true / the entered text on confirm,
// false / null on cancel.
import { computed, ref } from "vue";
import { useI18n } from "vue-i18n";

const props = withDefaults(
  defineProps<{
    title?: string;
    message: string;
    confirmText?: string;
    cancelText?: string;
    /** confirm becomes the danger colour */
    destructive?: boolean;
    /** prompt mode: show a text field */
    input?: {
      label?: string;
      placeholder?: string;
      value?: string;
      maxlength?: number;
      validate?: (v: string) => string | true;
    };
  }>(),
  {
    title: "",
    confirmText: "",
    cancelText: "",
    destructive: false,
    input: undefined,
  },
);
const emit = defineEmits<{ close: [result: boolean | string | null] }>();
const { t } = useI18n();

const value = ref(props.input?.value ?? "");
const error = computed(() => {
  if (!props.input?.validate) return "";
  const r = props.input.validate(value.value);
  return r === true ? "" : r;
});

function confirm() {
  if (props.input) {
    if (error.value) return;
    emit("close", value.value);
  } else {
    emit("close", true);
  }
}
function cancel() {
  emit("close", props.input ? null : false);
}
</script>

<template>
  <v-card>
    <v-card-title v-if="title" class="md3-title-large">{{
      title
    }}</v-card-title>
    <v-card-text class="md3-body-medium">
      <div>{{ message }}</div>
      <v-text-field
        v-if="input"
        v-model="value"
        class="mt-4"
        :label="input.label"
        :placeholder="input.placeholder"
        :maxlength="input.maxlength"
        :error-messages="error ? [error] : []"
        autofocus
        @keydown.enter.prevent="confirm"
      />
    </v-card-text>
    <v-card-actions>
      <v-spacer />
      <v-btn variant="text" @click="cancel">{{
        cancelText || t("operations.cancel")
      }}</v-btn>
      <v-btn
        variant="flat"
        :color="destructive ? 'error' : 'primary'"
        @click="confirm"
        >{{ confirmText || t("operations.confirm") }}</v-btn
      >
    </v-card-actions>
  </v-card>
</template>
