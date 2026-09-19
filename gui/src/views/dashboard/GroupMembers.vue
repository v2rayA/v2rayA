<script setup lang="ts">
// The group's members chosen from every node: a searchable list with a
// checkbox per node; Save replaces the group's member list at once.
import { computed, ref } from "vue";
import { useI18n } from "vue-i18n";
import { mdiMagnify } from "@mdi/js";
import type { Touch, TouchServer, Which } from "@/api/types";
import { filterRows, rowKey, whichOf } from "@/views/nodes/model";

defineOptions({ name: "GroupMembersDialog" });
const props = defineProps<{
  outbound: string;
  touch: Touch;
  members: Which[];
}>();
const emit = defineEmits<{ close: [touches?: Which[]] }>();
const { t } = useI18n();
const query = ref("");
/** one key per node, the same for a row and for a connected Which */
const keyOfWhich = (w: Which) =>
  `${w._type}|${w.id}|${w._type === "subscriptionServer" ? (w.sub ?? 0) : ""}`;
const chosen = ref(new Set(props.members.map(keyOfWhich)));

const rows = computed<TouchServer[]>(() =>
  filterRows(
    [
      ...props.touch.servers,
      ...props.touch.subscriptions.flatMap((s) => s.servers),
    ],
    query.value,
  ),
);
const sourceOf = (row: TouchServer) =>
  row._type === "server"
    ? t("proxies.sources.local")
    : props.touch.subscriptions[row.sub ?? -1]?.remarks ||
      props.touch.subscriptions[row.sub ?? -1]?.host ||
      "";
const keyOf = (row: TouchServer) => keyOfWhich(whichOf(row));
const isChosen = (row: TouchServer) => chosen.value.has(keyOf(row));
function toggle(row: TouchServer) {
  const key = keyOf(row);
  if (chosen.value.has(key)) chosen.value.delete(key);
  else chosen.value.add(key);
}
function save() {
  const all = [
    ...props.touch.servers,
    ...props.touch.subscriptions.flatMap((s) => s.servers),
  ];
  emit(
    "close",
    all.filter(isChosen).map((row) => ({
      ...whichOf(row),
      sub: row.sub ?? 0,
      outbound: props.outbound,
    })),
  );
}
</script>

<template>
  <v-card rounded="xl">
    <v-card-item class="px-6 pt-6 pb-2">
      <v-card-title class="md3-headline-small pa-0">
        {{ t("dashboard.editGroup") }}
        <span class="text-on-surface-variant ms-2">{{
          outbound.toUpperCase()
        }}</span>
      </v-card-title>
    </v-card-item>
    <v-card-text class="px-6 pb-0">
      <v-text-field
        v-model="query"
        :placeholder="t('proxyGroup.searchNodes')"
        :prepend-inner-icon="mdiMagnify"
        variant="solo-filled"
        bg-color="surface-container-high"
        rounded="pill"
        density="compact"
        flat
        hide-details
        clearable
        class="mb-2"
        @click:clear="query = ''"
      />
      <p class="md3-body-small text-on-surface-variant mb-2">
        {{ t("common.selectedCount", chosen.size) }}
      </p>
      <v-list class="group-members__list pa-0" bg-color="transparent">
        <v-list-item
          v-for="row in rows"
          :key="rowKey(row)"
          :title="row.name || row.address"
          :subtitle="[row.net, sourceOf(row)].filter((x) => x).join('  ')"
          rounded="lg"
          @click="toggle(row)"
        >
          <template #prepend>
            <v-checkbox-btn
              :model-value="isChosen(row)"
              :aria-label="row.name || row.address"
              @click.stop="toggle(row)"
            />
          </template>
        </v-list-item>
      </v-list>
      <v-empty-state v-if="!rows.length" :title="t('proxyGroup.noMatch')" />
    </v-card-text>
    <v-card-actions class="px-6 pb-4">
      <v-spacer />
      <v-btn variant="text" @click="emit('close')">{{
        t("operations.cancel")
      }}</v-btn>
      <v-btn variant="flat" color="primary" @click="save">{{
        t("operations.save")
      }}</v-btn>
    </v-card-actions>
  </v-card>
</template>

<style scoped>
.group-members__list {
  max-height: 50vh;
  overflow-y: auto;
}
</style>
