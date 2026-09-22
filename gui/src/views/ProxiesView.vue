<script setup lang="ts">
import { computed, onMounted } from "vue";
import { useI18n } from "vue-i18n";
import { useDisplay } from "vuetify";
import {
  mdiChevronDown,
  mdiContentCopy,
  mdiDownloadOutline,
  mdiMagnify,
  mdiPlus,
  mdiServerNetworkOutline,
  mdiSpeedometer,
  mdiTrayArrowDown,
  mdiViewGridOutline,
  mdiViewListOutline,
} from "@mdi/js";
import { rowKey } from "./nodes/model";
import { useProxies } from "./proxies/model";
import { focusInput, useHotkeys } from "@/composables";
import NodeCard from "./proxies/NodeCard.vue";
import NodeListItem from "./proxies/NodeListItem.vue";

defineOptions({ name: "ProxiesView" });
const { t } = useI18n();
const { width } = useDisplay();
const expanded = computed(() => width.value >= 840);
/** phones: the batch bar keeps the test button and folds the rest into a menu */
const compact = computed(() => width.value < 600);
const model = useProxies();
const { store } = model;
const {
  query,
  source,
  sources,
  view,
  loading,
  loadError,
  busy,
  testing,
  rows,
  listed,
  selected,
  selectedKeys,
  allSelected,
  canDelete,
  preferred,
  sync,
} = model;
const disabled = computed(() => busy.value || loading.value);
// Ctrl/Cmd+A selects every listed node, Escape clears (list view), "/" goes to the search
useHotkeys((event) => {
  const ctrl = event.ctrlKey || event.metaKey;
  if (ctrl && event.key.toLowerCase() === "a" && view.value === "list") {
    model.selectAll(true);
    event.preventDefault();
  } else if (event.key === "Escape" && selected.value.length) {
    model.selectAll(false);
  } else if (event.key === "/" && !ctrl) {
    focusInput(".proxies__search");
    event.preventDefault();
  }
});
defineExpose({ sync });
onMounted(sync);
</script>

<template>
  <div class="proxies" :class="{ 'proxies--expanded': expanded }">
    <div class="proxies__toolbar">
      <v-text-field
        v-model="query"
        :label="t('proxyGroup.searchNodes')"
        :prepend-inner-icon="mdiMagnify"
        variant="solo-filled"
        bg-color="surface-container-high"
        rounded="pill"
        flat
        hide-details
        clearable
        class="proxies__search"
        @click:clear="query = ''"
      />
      <v-spacer />
      <v-btn
        variant="outlined"
        :prepend-icon="mdiPlus"
        :disabled="disabled"
        @click="model.newNode"
        >{{ t("proxies.newNode") }}</v-btn
      >
      <v-btn
        v-if="rows.length || loading || loadError"
        variant="flat"
        color="primary"
        :prepend-icon="mdiTrayArrowDown"
        :disabled="disabled"
        @click="model.importNodes()"
        >{{ t("operations.import") }}</v-btn
      >
    </div>
    <v-alert v-if="loadError" type="error" variant="tonal" class="mb-4">
      {{ loadError }}
      <template #append
        ><v-btn variant="text" @click="sync">{{
          t("operations.update")
        }}</v-btn></template
      >
    </v-alert>
    <v-skeleton-loader
      v-if="loading && !rows.length"
      type="card, card"
      class="bg-transparent"
    />
    <template v-else-if="!loadError || rows.length">
      <section>
        <div class="proxies__row mb-3">
          <h2 class="md3-title-medium ma-0 me-2">
            {{ t("common.nodes") }}
            <span class="md3-label-medium text-on-surface-variant ms-1">{{
              rows.length
            }}</span>
          </h2>
          <v-select
            v-model="source"
            :items="sources"
            :label="t('proxies.source')"
            variant="outlined"
            density="compact"
            hide-details
            class="proxies__source"
          />
          <v-spacer />
          <v-btn
            variant="tonal"
            :prepend-icon="mdiSpeedometer"
            :loading="testing"
            :disabled="disabled || !listed.length"
            @click="model.testListed()"
            >{{ t("proxies.testLatency") }}</v-btn
          >
          <v-btn-toggle
            v-model="view"
            mandatory
            divided
            variant="outlined"
            rounded="xl"
            selected-class="bg-secondary-container text-on-secondary-container"
            :aria-label="t('operations.view')"
          >
            <v-btn value="cards" :aria-label="t('proxies.cards')" width="48">
              <v-icon :icon="mdiViewGridOutline" size="20" />
            </v-btn>
            <v-btn value="list" :aria-label="t('proxies.list')" width="48">
              <v-icon :icon="mdiViewListOutline" size="20" />
            </v-btn>
          </v-btn-toggle>
        </div>
        <v-empty-state
          v-if="!rows.length"
          :icon="mdiServerNetworkOutline"
          :title="t('common.empty')"
          :text="t('proxies.emptyHint')"
        >
          <template #actions
            ><v-btn
              variant="flat"
              color="primary"
              :disabled="disabled"
              @click="model.importNodes()"
              >{{ t("operations.import") }}</v-btn
            ></template
          >
        </v-empty-state>
        <v-empty-state
          v-else-if="!listed.length"
          :title="t('proxyGroup.noMatch')"
        />
        <template v-else>
          <div class="proxies__batch mb-2">
            <v-checkbox-btn
              :model-value="allSelected"
              :indeterminate="selected.length > 0 && !allSelected"
              :aria-label="t('proxies.selectAll')"
              :disabled="disabled"
              @update:model-value="model.selectAll(!!$event)"
            />
            <span class="md3-label-large">{{
              t("common.selectedCount", selected.length)
            }}</span>
            <template v-if="selected.length">
              <v-btn
                variant="text"
                :disabled="disabled || testing"
                @click="model.testRows(selected)"
                >{{ t("proxies.testLatency") }}</v-btn
              >
              <v-menu v-if="compact">
                <template #activator="{ props: menu }">
                  <v-btn
                    v-bind="menu"
                    variant="text"
                    :append-icon="mdiChevronDown"
                    :disabled="disabled"
                    >{{ t("operations.moreActions") }}</v-btn
                  >
                </template>
                <v-list density="compact" min-width="240">
                  <template v-for="add in [true, false]" :key="String(add)">
                    <v-list-subheader>{{
                      t(add ? "proxies.addToGroup" : "proxies.removeFromGroup")
                    }}</v-list-subheader>
                    <v-list-item
                      v-for="g in store.outbounds"
                      :key="g"
                      :title="g.toUpperCase()"
                      class="ps-8"
                      @click="model.batchMembership(add, g)"
                    />
                  </template>
                  <v-divider class="my-1" />
                  <v-list-item
                    :title="t('operations.delete')"
                    :subtitle="
                      canDelete
                        ? undefined
                        : t('proxies.deleteSubscriptionNodes')
                    "
                    :disabled="disabled || !canDelete"
                    @click="model.removeRows(selected)"
                  />
                  <v-divider class="my-1" />
                  <v-list-item
                    :title="t('operations.exportClipboard')"
                    :prepend-icon="mdiContentCopy"
                    :disabled="disabled"
                    @click="model.exportSelected('clipboard')"
                  />
                  <v-list-item
                    :title="t('operations.exportTxt')"
                    :prepend-icon="mdiDownloadOutline"
                    :disabled="disabled"
                    @click="model.exportSelected('file')"
                  />
                </v-list>
              </v-menu>
              <template v-else>
                <v-menu v-for="add in [true, false]" :key="String(add)">
                  <template #activator="{ props: menu }">
                    <v-btn
                      v-bind="menu"
                      variant="text"
                      :append-icon="mdiChevronDown"
                      :disabled="disabled"
                      >{{
                        t(
                          add
                            ? "proxies.addToGroup"
                            : "proxies.removeFromGroup",
                        )
                      }}</v-btn
                    >
                  </template>
                  <v-list density="compact" min-width="200">
                    <v-list-item
                      v-for="g in store.outbounds"
                      :key="g"
                      :title="g.toUpperCase()"
                      @click="model.batchMembership(add, g)"
                    />
                  </v-list>
                </v-menu>
                <v-tooltip
                  :disabled="canDelete"
                  :text="t('proxies.deleteSubscriptionNodes')"
                >
                  <template #activator="{ props }"
                    ><span v-bind="props" :tabindex="canDelete ? undefined : 0"
                      ><v-btn
                        variant="text"
                        :disabled="disabled || !canDelete"
                        @click="model.removeRows(selected)"
                        >{{ t("operations.delete") }}</v-btn
                      ></span
                    ></template
                  >
                </v-tooltip>
                <v-menu>
                  <template #activator="{ props: exportMenu }">
                    <v-btn
                      v-bind="exportMenu"
                      variant="text"
                      :append-icon="mdiChevronDown"
                      :disabled="disabled"
                      >{{ t("operations.export") }}</v-btn
                    >
                  </template>
                  <v-list density="compact" min-width="240">
                    <v-list-item
                      :title="t('operations.exportClipboard')"
                      :prepend-icon="mdiContentCopy"
                      @click="model.exportSelected('clipboard')"
                    />
                    <v-list-item
                      :title="t('operations.exportTxt')"
                      :prepend-icon="mdiDownloadOutline"
                      @click="model.exportSelected('file')"
                    />
                  </v-list>
                </v-menu>
              </template>
            </template>
          </div>
          <div v-if="view === 'cards'" class="proxies__cards">
            <NodeCard
              v-for="row in listed"
              :key="rowKey(row)"
              :row="row"
              :source="model.sourceName(row)"
              :member="model.isMember(row)"
              :groups="store.outbounds"
              :member-groups="model.memberGroups(row)"
              :in-use="!!preferred && rowKey(preferred) === rowKey(row)"
              :checked="selectedKeys.includes(rowKey(row))"
              :disabled="disabled"
              :testing="testing"
              @check="model.selectRow(row, $event)"
              @action="(action, group) => model.nodeAction(row, action, group)"
            />
          </div>
          <v-list
            v-else
            bg-color="surface-container-low"
            rounded="xl"
            class="proxies__list"
          >
            <NodeListItem
              v-for="row in listed"
              :key="rowKey(row)"
              :row="row"
              :source="model.sourceName(row)"
              :member="model.isMember(row)"
              :groups="store.outbounds"
              :member-groups="model.memberGroups(row)"
              :in-use="!!preferred && rowKey(preferred) === rowKey(row)"
              :checked="selectedKeys.includes(rowKey(row))"
              :disabled="disabled"
              :testing="testing"
              @check="model.selectRow(row, $event)"
              @action="(action, group) => model.nodeAction(row, action, group)"
            />
          </v-list>
        </template>
      </section>
    </template>
  </div>
</template>

<style scoped>
.proxies__row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}
.proxies__label {
  min-width: 40px;
}
.proxies__source {
  max-width: 220px;
  min-width: 160px;
}
.proxies {
  padding-bottom: 96px;
}
.proxies__toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin-bottom: 24px;
}
.proxies__search {
  flex: 1 1 280px;
  max-width: 480px;
}
.proxies__cards {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 12px;
}
.proxies--expanded .proxies__cards {
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
}
.proxies__batch {
  display: flex;
  align-items: center;
  gap: 8px;
  white-space: nowrap;
  overflow-x: auto;
  min-height: 48px;
}
.proxies__batch > * {
  flex-shrink: 0;
}
.proxies__batch :deep(.v-selection-control) {
  flex: 0 0 48px;
}
.proxies :deep(.v-btn) {
  min-height: 48px;
}
.proxies :deep(.v-chip--link) {
  min-height: 48px;
}
.proxies :deep(.node-menu),
.proxies :deep(.subscription-card__menu) {
  min-height: 40px;
}
</style>
