<script setup lang="ts">
import { useI18n } from "vue-i18n";
import { mdiDotsVertical } from "@mdi/js";
import type { NodeAction } from "./model";
defineOptions({ name: "NodeMenu" });
defineProps<{
  member: boolean;
  selected: boolean;
  local: boolean;
  disabled: boolean;
  testing: boolean;
}>();
const emit = defineEmits<{ action: [action: NodeAction] }>();
const { t } = useI18n();
</script>
<template>
  <v-menu>
    <template #activator="{ props }">
      <v-btn
        v-bind="props"
        :icon="mdiDotsVertical"
        variant="text"
        size="40"
        class="node-menu"
        :aria-label="t('operations.name')"
        :disabled="disabled"
        @click.stop
        @keydown.stop
      >
        <v-icon :icon="mdiDotsVertical" size="18" />
        <v-tooltip activator="parent" location="top">{{
          t("operations.name")
        }}</v-tooltip>
      </v-btn>
    </template>
    <v-list>
      <v-list-item
        :title="t('proxies.testLatency')"
        :disabled="testing"
        @click="emit('action', 'test')"
      />
      <v-list-item
        v-if="member"
        :title="t(selected ? 'proxies.unselect' : 'proxies.useThis')"
        @click="emit('action', 'select')"
      />
      <v-list-item
        :title="t(member ? 'proxies.removeFromGroup' : 'proxies.addToGroup')"
        @click="emit('action', 'membership')"
      />
      <v-list-item
        :title="t('operations.modify')"
        @click="emit('action', 'edit')"
      />
      <v-list-item
        :title="t('operations.share')"
        @click="emit('action', 'share')"
      />
      <v-list-item
        v-if="local"
        :title="t('operations.delete')"
        @click="emit('action', 'delete')"
      />
    </v-list>
  </v-menu>
</template>
<style scoped>
.node-menu {
  margin: 4px;
}
.node-menu::after {
  content: "";
  position: absolute;
  inset: -4px;
}
</style>
