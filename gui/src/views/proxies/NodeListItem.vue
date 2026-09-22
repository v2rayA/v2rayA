<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { mdiCheck } from "@mdi/js";
import type { Row } from "../nodes/model";
import type { NodeAction } from "./model";
import NodeLatency from "./NodeLatency.vue";
import NodeMenu from "./NodeMenu.vue";
defineOptions({ name: "NodeListItem" });
defineProps<{
  row: Row;
  source: string;
  member: boolean;
  /** every proxy group, and the ones this node is a member of */
  groups: string[];
  memberGroups: string[];
  inUse: boolean;
  checked: boolean;
  disabled: boolean;
  testing: boolean;
}>();
const emit = defineEmits<{
  check: [value: boolean];
  action: [action: NodeAction, group?: string];
}>();
const { t } = useI18n();
</script>
<template>
  <v-list-item class="node-list-item" :disabled="disabled">
    <template #prepend>
      <v-checkbox-btn
        :model-value="checked"
        :aria-label="row.name"
        @click.stop
        @update:model-value="emit('check', !!$event)"
      />
    </template>
    <v-list-item-title class="md3-title-small" dir="auto">{{
      row.name
    }}</v-list-item-title>
    <v-list-item-subtitle class="md3-body-small"
      >{{ row.net }} {{ source }}</v-list-item-subtitle
    >
    <template #append>
      <div class="node-list-item__append d-flex align-center ga-2">
        <NodeLatency :latency="row.pingLatency" />
        <v-chip v-if="inUse" size="small" variant="tonal">{{
          t("proxies.inUse")
        }}</v-chip>
        <v-icon
          v-else-if="member"
          :icon="mdiCheck"
          size="20"
          :aria-label="t('proxies.membersOnly')"
        />
        <NodeMenu
          :groups="groups"
          :member-groups="memberGroups"
          :local="row._type === 'server'"
          :disabled="disabled"
          :testing="testing"
          @action="(action, group) => emit('action', action, group)"
        />
      </div>
    </template>
  </v-list-item>
</template>
<style scoped>
.node-list-item {
  min-height: 64px;
}
.node-list-item__append {
  flex-shrink: 0;
  white-space: nowrap;
}
.node-list-item :deep(.v-list-item__prepend > .v-list-item__spacer) {
  width: 8px;
}
.node-list-item :deep(.v-list-item__append) {
  padding-inline-start: 8px;
}
</style>
