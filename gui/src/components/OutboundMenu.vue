<script setup lang="ts">
// The outbound groups: a button naming the current one, a menu to
// switch, add (a prompt), open a group's settings and delete (a
// confirmation). Membership is edited from the node list.
import { ref } from "vue";
import { useI18n } from "vue-i18n";
import {
  mdiCheck,
  mdiChevronDown,
  mdiCogOutline,
  mdiDeleteOutline,
  mdiPlus,
  mdiSitemapOutline,
} from "@mdi/js";
import { useDialog, useOutboundGroups } from "@/composables";
import OutboundGroupDialog from "@/dialogs/OutboundGroup.vue";
import { useAppStore } from "@/stores/app";

defineProps<{
  /** a list item in the drawer, an icon button in the rail, else a chip */
  variant?: "chip" | "list" | "icon";
}>();
const emit = defineEmits<{ changed: [] }>();
const { t } = useI18n();
const store = useAppStore();
const { open: openDialog } = useDialog();
const open = ref(false);

function settings(outbound: string) {
  open.value = false;
  openDialog(OutboundGroupDialog, { outbound }, { width: 440 });
}

const groups = useOutboundGroups();
const add = () => groups.add();
async function remove(outbound: string) {
  if (await groups.remove(outbound)) emit("changed");
}
</script>

<template>
  <v-menu v-model="open" :close-on-content-click="false">
    <template #activator="{ props: menu }">
      <v-list-item
        v-if="variant === 'list'"
        v-bind="menu"
        :prepend-icon="mdiSitemapOutline"
        :title="t('common.proxyGroups')"
        :subtitle="store.outboundName.toUpperCase()"
        rounded="xl"
      />
      <v-tooltip
        v-else-if="variant === 'icon'"
        :text="`${t('common.proxyGroups')}: ${store.outboundName.toUpperCase()}`"
        location="end"
      >
        <template #activator="{ props: tip }">
          <v-btn
            v-bind="{ ...menu, ...tip }"
            :icon="mdiSitemapOutline"
            variant="text"
            :aria-label="t('common.proxyGroups')"
          />
        </template>
      </v-tooltip>
      <v-btn
        v-else
        v-bind="menu"
        :prepend-icon="mdiSitemapOutline"
        :append-icon="mdiChevronDown"
        color="primary"
        variant="tonal"
        class="text-none"
      >
        {{ store.outboundName.toUpperCase() }}
      </v-btn>
    </template>
    <v-list density="compact" min-width="240">
      <v-list-subheader>{{ t("common.proxyGroups") }}</v-list-subheader>
      <v-list-item
        v-for="outbound in store.outbounds"
        :key="outbound"
        :active="outbound === store.outboundName"
        :title="outbound.toUpperCase()"
        @click="
          store.outboundName = outbound;
          open = false;
        "
      >
        <template #prepend>
          <v-icon
            :icon="mdiCheck"
            :class="{ invisible: outbound !== store.outboundName }"
          />
        </template>
        <template #append>
          <v-btn
            :icon="mdiCogOutline"
            variant="text"
            size="small"
            :aria-label="t('common.setting')"
            @click.stop="settings(outbound)"
          />
          <v-btn
            v-if="outbound !== 'proxy'"
            :icon="mdiDeleteOutline"
            variant="text"
            size="small"
            :aria-label="t('operations.delete')"
            @click.stop="remove(outbound)"
          />
        </template>
      </v-list-item>
      <v-divider class="my-1" />
      <v-list-item
        :prepend-icon="mdiPlus"
        :title="t('operations.create')"
        @click="add"
      />
    </v-list>
  </v-menu>
</template>

<style scoped>
.invisible {
  visibility: hidden;
}
</style>
