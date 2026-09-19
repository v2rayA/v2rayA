<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import {
  mdiArrowUp,
  mdiArrowDown,
  mdiArrowRight,
  mdiDeleteOutline,
  mdiDotsVertical,
  mdiPencilOutline,
} from "@mdi/js";
import { useConfirm, useDialog } from "@/composables";
import { parse, serialize, type Entry } from "./rules";
import RuleDialog from "./RuleDialog.vue";
import OutboundDialog from "./OutboundDialog.vue";

defineOptions({ name: "RoutingForm" });
const props = defineProps<{ modelValue: string; disabled?: boolean }>();
const emit = defineEmits<{ "update:modelValue": [text: string] }>();
const { t } = useI18n();
const { open } = useDialog();
const confirm = useConfirm();
const entries = ref<Entry[]>([]);
watch(
  () => props.modelValue,
  (text) => {
    entries.value = parse(text);
  },
  { immediate: true },
);
const names = computed(() => [
  "proxy",
  "direct",
  "block",
  ...entries.value.flatMap((entry) =>
    entry.kind === "outbound" ? [entry.name] : [],
  ),
]);
const defaults = computed(() =>
  entries.value.find((entry) => entry.kind === "default"),
);
const fallback = computed(() =>
  defaults.value?.kind === "default" ? defaults.value.outbound : "proxy",
);
const rows = computed(() =>
  entries.value
    .map((entry, index) => ({ entry, index }))
    .filter(
      ({ entry }) =>
        entry.kind !== "blank" &&
        entry.kind !== "default" &&
        entry.kind !== "outbound",
    ),
);
const outbounds = computed(() =>
  entries.value
    .map((entry, index) => ({ entry, index }))
    .filter(({ entry }) => entry.kind === "outbound"),
);
function update() {
  emit("update:modelValue", serialize(entries.value));
}
function setDefault(value: string) {
  if (defaults.value?.kind === "default") defaults.value.outbound = value;
  else entries.value.unshift({ kind: "default", outbound: value });
  update();
}
function move(index: number, direction: number) {
  const target =
    rows.value[rows.value.findIndex((row) => row.index === index) + direction]
      ?.index;
  if (target === undefined) return;
  const [entry] = entries.value.splice(index, 1);
  entries.value.splice(target, 0, entry);
  update();
}
async function remove(index: number) {
  if (
    await confirm({
      message: t("routingA.form.deleteConfirm"),
      destructive: true,
      confirmText: t("operations.delete"),
      cancelText: t("operations.cancel"),
    })
  ) {
    entries.value.splice(index, 1);
    update();
  }
}
async function edit(index?: number, outbound = false) {
  const entry = index === undefined ? undefined : entries.value[index];
  if (entry && entry.kind !== "rule" && entry.kind !== "outbound") return;
  const result = await open<Entry>(
    outbound ? OutboundDialog : RuleDialog,
    outbound
      ? { entry, names: names.value }
      : { entry, outbounds: names.value },
    { width: 560 },
  ).result;
  if (!result) return;
  if (index === undefined) entries.value.push(result);
  else entries.value[index] = result;
  update();
}

/** the outbound's colour: block in error, direct in tertiary, anything else (proxy, custom) in primary */
function outboundColor(outbound: string) {
  return outbound === "block"
    ? "error"
    : outbound === "direct"
      ? "tertiary"
      : "primary";
}
</script>

<template>
  <div class="routing-form">
    <v-select
      :model-value="fallback"
      :items="[...new Set([...names, fallback])]"
      :label="t('routingA.form.default')"
      :disabled="disabled"
      @update:model-value="setDefault"
    />
    <p class="md3-body-small mb-2">{{ t("routingA.form.orderHint") }}</p>
    <v-list bg-color="surface-container-lowest" rounded="lg" class="mb-4">
      <v-list-item
        v-for="({ entry, index }, position) in rows"
        :key="index"
        :disabled="disabled"
        :class="{ 'routing-form__comment': entry.kind === 'comment' }"
        lines="two"
        :link="entry.kind === 'rule'"
        @click="entry.kind === 'rule' && edit(index)"
      >
        <template v-if="entry.kind === 'rule'">
          <v-list-item-title class="routing-form__conditions" dir="ltr">
            <template v-for="(condition, part) in entry.conditions" :key="part">
              <span v-if="part" class="text-on-surface-variant"> && </span>
              <span class="routing-form__fn">{{ condition.fn }}</span
              >({{ condition.args.join(", ") }})
            </template>
          </v-list-item-title>
          <v-list-item-subtitle dir="ltr" class="mt-1">
            <v-chip
              size="small"
              variant="tonal"
              :color="outboundColor(entry.outbound)"
              :prepend-icon="mdiArrowRight"
              >{{ entry.outbound }}</v-chip
            >
          </v-list-item-subtitle>
        </template>
        <div
          v-else-if="entry.kind === 'comment' || entry.kind === 'raw'"
          class="routing-form__raw"
          dir="ltr"
        >
          <span
            v-if="entry.kind === 'raw'"
            class="text-error"
            :title="t('routingA.form.raw')"
            >• </span
          >{{ entry.text }}
        </div>
        <template #append>
          <v-menu>
            <template #activator="{ props: menu }"
              ><v-btn
                v-bind="menu"
                :icon="mdiDotsVertical"
                variant="text"
                size="40"
                :aria-label="t('routingA.form.actions')"
                :disabled="disabled"
                @click.stop
            /></template>
            <v-list density="compact">
              <v-list-item
                v-if="entry.kind === 'rule'"
                :prepend-icon="mdiPencilOutline"
                :title="t('operations.modify')"
                @click="edit(index)"
              />
              <v-list-item
                :prepend-icon="mdiArrowUp"
                :title="t('routingA.form.moveUp')"
                :disabled="position === 0"
                @click="move(index, -1)"
              />
              <v-list-item
                :prepend-icon="mdiArrowDown"
                :title="t('routingA.form.moveDown')"
                :disabled="position === rows.length - 1"
                @click="move(index, 1)"
              />
              <v-list-item
                :prepend-icon="mdiDeleteOutline"
                :title="t('operations.delete')"
                @click="remove(index)"
              />
            </v-list>
          </v-menu>
        </template>
      </v-list-item>
    </v-list>
    <v-btn variant="tonal" :disabled="disabled" @click="edit()">{{
      t("routingA.form.addRule")
    }}</v-btn>
    <h3 class="md3-title-medium mt-6 mb-2">
      {{ t("routingA.form.customOutbounds") }}
    </h3>
    <v-list
      v-if="outbounds.length"
      bg-color="surface-container-lowest"
      rounded="lg"
      class="mb-4"
    >
      <template v-for="{ entry, index } in outbounds" :key="index">
        <v-list-item
          v-if="entry.kind === 'outbound'"
          :title="entry.name"
          :subtitle="`${entry.fn} · ${entry.args.find((arg) => 'address' in arg)?.address ?? ''}:${entry.args.find((arg) => 'port' in arg)?.port ?? ''}`"
          :disabled="disabled"
          @click="edit(index, true)"
        >
          <template #append>
            <v-menu>
              <template #activator="{ props: menu }"
                ><v-btn
                  v-bind="menu"
                  :icon="mdiDotsVertical"
                  variant="text"
                  size="40"
                  :aria-label="t('routingA.form.actions')"
                  :disabled="disabled"
                  @click.stop
              /></template>
              <v-list>
                <v-list-item
                  :title="t('operations.modify')"
                  @click="edit(index, true)"
                />
                <v-list-item
                  :title="t('operations.delete')"
                  @click="remove(index)"
                />
              </v-list>
            </v-menu>
          </template>
        </v-list-item>
      </template>
    </v-list>
    <v-btn variant="text" :disabled="disabled" @click="edit(undefined, true)">{{
      t("routingA.form.addOutbound")
    }}</v-btn>
  </div>
</template>

<style scoped>
.routing-form {
  min-width: 0;
  max-height: 60vh;
  overflow: auto;
  padding-top: 8px;
}
.routing-form__comment {
  color: rgb(var(--v-theme-outline));
}
.routing-form__raw {
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  font-family: ui-monospace, "Cascadia Mono", "Fira Mono", Menlo, monospace;
}
.routing-form__conditions {
  white-space: normal;
  overflow-wrap: anywhere;
  font-family: ui-monospace, "Cascadia Mono", "Fira Mono", Menlo, monospace;
  font-size: 13px;
  line-height: 20px;
}
.routing-form__fn {
  color: rgb(var(--v-theme-tertiary));
}
@media (max-width: 599px) {
  .routing-form :deep(.v-list-item__append) {
    grid-column: 2;
    grid-row: 2;
    justify-self: end;
  }
}
</style>
