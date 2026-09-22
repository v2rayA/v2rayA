<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { mdiCheck } from "@mdi/js";
import type { Row } from "../nodes/model";
import type { NodeAction } from "./model";
import NodeLatency from "./NodeLatency.vue";
import NodeMenu from "./NodeMenu.vue";
defineOptions({ name: "NodeCard" });
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
  <v-card
    rounded="lg"
    :variant="member ? 'tonal' : 'flat'"
    :color="member ? 'primary' : 'surface-container-low'"
    class="node-card ps-2 pe-2 py-2"
  >
    <div class="d-flex align-start">
      <v-checkbox-btn
        :model-value="checked"
        :aria-label="row.name"
        :disabled="disabled"
        class="flex-shrink-0"
        @click.stop
        @update:model-value="emit('check', !!$event)"
      />
      <div class="node-card__body flex-grow-1">
        <div class="d-flex align-center ga-2">
          <v-icon
            v-if="member"
            :icon="mdiCheck"
            size="20"
            color="primary"
            class="flex-shrink-0"
          />
          <h3 class="node-card__name md3-title-small" dir="auto">
            {{ row.name }}
          </h3>
          <v-chip
            v-if="inUse"
            size="small"
            variant="tonal"
            class="flex-shrink-0"
            >{{ t("proxies.inUse") }}</v-chip
          >
          <NodeMenu
            class="node-card__menu"
            :groups="groups"
            :member-groups="memberGroups"
            :local="row._type === 'server'"
            :disabled="disabled"
            :testing="testing"
            @action="(action, group) => emit('action', action, group)"
          />
        </div>
        <div class="d-flex align-center ga-2 mt-1">
          <p
            class="md3-body-small text-on-surface-variant node-card__source ma-0"
          >
            {{ row.net }} {{ source }}
          </p>
          <v-spacer />
          <NodeLatency :latency="row.pingLatency" />
        </div>
      </div>
    </div>
  </v-card>
</template>
<style scoped>
.node-card__body {
  min-width: 0;
}
.node-card__menu {
  margin-inline-start: auto;
}
.node-card__name {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  overflow-wrap: anywhere;
  min-width: 0;
  flex: 1;
}
.node-card__source {
  overflow-wrap: anywhere;
  min-width: 0;
}
</style>
