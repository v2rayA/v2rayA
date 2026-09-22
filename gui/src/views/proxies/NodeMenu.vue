<script setup lang="ts">
import { computed, ref } from "vue";
import { useI18n } from "vue-i18n";
import { useDisplay } from "vuetify";
import { mdiChevronRight, mdiDotsVertical } from "@mdi/js";
import type { NodeAction } from "./model";
defineOptions({ name: "NodeMenu" });
defineProps<{
  /** every proxy group the core knows, in the app bar's order */
  groups: string[];
  /** the groups this node is already a member of */
  memberGroups: string[];
  local: boolean;
  disabled: boolean;
  testing: boolean;
}>();
const emit = defineEmits<{
  action: [action: NodeAction, group?: string];
}>();
const { t } = useI18n();
const open = ref(false);
// a submenu has no room beside a phone-wide menu: the groups are listed under a heading instead
const { width } = useDisplay();
const compact = computed(() => width.value < 600);
/** pick runs the action and closes the menu, submenu included */
function pick(action: NodeAction, group?: string) {
  emit("action", action, group);
  open.value = false;
}
</script>
<template>
  <v-menu v-model="open">
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
        @click="pick('test')"
      />
      <template v-if="compact">
        <v-list-subheader>{{ t("proxies.addToGroup") }}</v-list-subheader>
        <v-list-item
          v-for="group in groups"
          :key="'add-' + group"
          :title="group.toUpperCase()"
          :active="memberGroups.includes(group)"
          :disabled="memberGroups.includes(group)"
          class="ps-8"
          @click="pick('addToGroup', group)"
        />
        <template v-if="memberGroups.length">
          <v-list-subheader>{{
            t("proxies.removeFromGroup")
          }}</v-list-subheader>
          <v-list-item
            v-for="group in memberGroups"
            :key="'remove-' + group"
            :title="group.toUpperCase()"
            class="ps-8"
            @click="pick('removeFromGroup', group)"
          />
        </template>
        <v-divider class="my-1" />
      </template>
      <template v-else>
        <v-menu submenu location="end">
          <template #activator="{ props: submenu }">
            <v-list-item
              v-bind="submenu"
              :title="t('proxies.addToGroup')"
              :append-icon="mdiChevronRight"
              :disabled="!groups.length"
            />
          </template>
          <v-list density="compact" min-width="200">
            <v-list-item
              v-for="group in groups"
              :key="group"
              :title="group.toUpperCase()"
              :active="memberGroups.includes(group)"
              :disabled="memberGroups.includes(group)"
              @click="pick('addToGroup', group)"
            />
          </v-list>
        </v-menu>
        <v-menu submenu location="end">
          <template #activator="{ props: submenu }">
            <v-list-item
              v-bind="submenu"
              :title="t('proxies.removeFromGroup')"
              :append-icon="mdiChevronRight"
              :disabled="!memberGroups.length"
            />
          </template>
          <v-list density="compact" min-width="200">
            <v-list-item
              v-for="group in memberGroups"
              :key="group"
              :title="group.toUpperCase()"
              @click="pick('removeFromGroup', group)"
            />
          </v-list>
        </v-menu>
      </template>
      <v-list-item :title="t('operations.modify')" @click="pick('edit')" />
      <v-list-item :title="t('operations.share')" @click="pick('share')" />
      <v-list-item
        v-if="local"
        :title="t('operations.delete')"
        @click="pick('delete')"
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
