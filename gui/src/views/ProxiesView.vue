<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { useDisplay } from "vuetify";
import {
  mdiCheck,
  mdiChevronDown,
  mdiChevronUp,
  mdiCogOutline,
  mdiDeleteOutline,
  mdiDotsVertical,
  mdiMagnify,
  mdiPlus,
  mdiRss,
  mdiServerNetworkOutline,
  mdiSpeedometer,
  mdiTrayArrowDown,
  mdiViewGridOutline,
  mdiViewListOutline,
} from "@mdi/js";
import { rowKey, sameWhich, whichOf } from "./nodes/model";
import { useProxies } from "./proxies/model";
import NodeCard from "./proxies/NodeCard.vue";
import NodeListItem from "./proxies/NodeListItem.vue";
import SubscriptionCard from "./proxies/SubscriptionCard.vue";

defineOptions({ name: "ProxiesView" });
const { t } = useI18n();
const { width } = useDisplay();
const expanded = computed(() => width.value >= 840);
const model = useProxies();
const { store } = model;
const {
  query,
  source,
  sources,
  membersOnly,
  view,
  loading,
  loadError,
  busy,
  testing,
  rows,
  subscriptions,
  members,
  groupList,
  listed,
  selected,
  selectedKeys,
  allSelected,
  canDelete,
  preferred,
  sync,
} = model;
const disabled = computed(() => busy.value || loading.value);
/** the member the group routes through alone, as a row; null while balancing */
const currentMember = computed(() => {
  const which = model.selectedMember.value;
  return which
    ? (members.value.find((row) => sameWhich(whichOf(row), which)) ?? null)
    : null;
});
// the subscriptions fold away once the user has seen them
const subscriptionsOpen = ref(
  localStorage.getItem("proxies.subscriptions") !== "closed",
);
watch(subscriptionsOpen, (open) =>
  localStorage.setItem("proxies.subscriptions", open ? "open" : "closed"),
);
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
      <section class="mb-6">
        <div
          class="d-flex align-center ga-2"
          :class="subscriptionsOpen ? 'mb-4' : ''"
        >
          <v-btn
            :icon="subscriptionsOpen ? mdiChevronUp : mdiChevronDown"
            variant="text"
            size="40"
            :aria-label="t('common.subscriptions')"
            :aria-expanded="subscriptionsOpen"
            @click="subscriptionsOpen = !subscriptionsOpen"
          />
          <h2 class="md3-title-medium ma-0">
            {{ t("common.subscriptions") }}
            <span
              v-if="!subscriptionsOpen"
              class="md3-label-medium text-on-surface-variant ms-1"
              >{{ subscriptions.length }}</span
            >
          </h2>
          <v-spacer />
          <v-btn
            variant="text"
            :prepend-icon="mdiCogOutline"
            :disabled="disabled"
            @click="model.subscriptionSettings"
            >{{ t("proxies.autoUpdate") }}</v-btn
          >
        </div>
        <v-expand-transition>
          <div v-if="subscriptionsOpen">
            <div v-if="subscriptions.length" class="proxies__subscriptions">
              <SubscriptionCard
                v-for="subscription in subscriptions"
                :key="subscription.address"
                :subscription="subscription"
                :disabled="disabled"
                @action="model.subscriptionAction(subscription, $event)"
              />
            </div>
            <v-sheet
              v-else
              color="surface-container-low"
              rounded="xl"
              class="proxies__empty"
            >
              <v-icon
                :icon="mdiRss"
                size="28"
                color="on-surface-variant"
                class="flex-shrink-0"
              />
              <div class="proxies__empty-text">
                <p class="md3-title-medium ma-0">
                  {{ t("proxies.noSubscriptions") }}
                </p>
                <p class="md3-body-medium text-on-surface-variant ma-0">
                  {{ t("proxies.noSubscriptionsHint") }}
                </p>
              </div>
              <v-btn
                variant="tonal"
                color="primary"
                class="flex-shrink-0"
                :disabled="disabled"
                @click="model.importNodes('subscription')"
                >{{ t("proxies.importSubscription") }}</v-btn
              >
            </v-sheet>
          </div>
        </v-expand-transition>
      </section>
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
          <v-chip
            :model-value="true"
            :aria-pressed="membersOnly"
            link
            variant="text"
            :prepend-icon="membersOnly ? mdiCheck : undefined"
            class="proxies__chip"
            :class="{ 'proxies__chip--on': membersOnly }"
            @click="membersOnly = !membersOnly"
            >{{ t("proxies.membersOnly") }}</v-chip
          >
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
        <div class="proxies__row mb-3">
          <span class="proxies__label md3-label-large text-on-surface-variant">
            {{ t("proxyGroup.group") }}
          </span>
          <v-chip-group
            :model-value="store.outboundName"
            mandatory
            :disabled="disabled"
            selected-class="proxies__chip--on"
            @update:model-value="(v: string) => v && (store.outboundName = v)"
          >
            <v-chip
              v-for="g in groupList"
              :key="g.name"
              :value="g.name"
              variant="text"
              filter
              class="proxies__chip"
            >
              <span class="proxies__name">{{ g.name.toUpperCase() }}</span>
              <span class="ms-2 md3-label-medium">
                {{ t("proxies.members", g.count) }}
              </span>
            </v-chip>
          </v-chip-group>
          <v-btn
            variant="text"
            :prepend-icon="mdiPlus"
            :disabled="disabled"
            @click="model.newGroup"
            >{{ t("proxies.newGroup") }}</v-btn
          >
          <v-spacer />
          <v-menu>
            <template #activator="{ props: menu }">
              <v-chip
                v-bind="menu"
                variant="text"
                class="proxies__chip"
                :append-icon="mdiChevronDown"
                :disabled="disabled || !members.length"
                :aria-label="t('proxies.inUse')"
                ><span class="text-on-surface-variant me-1"
                  >{{ t("proxies.inUse") }}:</span
                >
                <span dir="auto">{{
                  currentMember
                    ? currentMember.name
                    : preferred
                      ? `${t("proxies.mode.auto")}  ${preferred.name}`
                      : t("proxies.mode.auto")
                }}</span></v-chip
              >
            </template>
            <v-list density="compact" min-width="280">
              <v-list-item
                :title="t('proxies.mode.auto')"
                :subtitle="t('proxies.modeHint.auto')"
                :active="!currentMember"
                role="menuitemradio"
                :aria-checked="!currentMember"
                @click="model.selectMember(null)"
              />
              <v-divider />
              <v-list-item
                v-for="row in members"
                :key="rowKey(row)"
                :title="row.name || row.address"
                :subtitle="`${row.net}${row.pingLatency ? ' ' + row.pingLatency : ''}`"
                :active="currentMember === row"
                role="menuitemradio"
                :aria-checked="currentMember === row"
                @click="model.selectMember(row)"
              />
            </v-list>
          </v-menu>
          <v-menu>
            <template #activator="{ props: menu }">
              <v-btn
                v-bind="menu"
                :icon="mdiDotsVertical"
                variant="text"
                size="40"
                :disabled="disabled"
                :aria-label="t('common.menu')"
              >
                <v-icon :icon="mdiDotsVertical" size="20" />
              </v-btn>
            </template>
            <v-list density="compact" min-width="220">
              <v-list-item
                :prepend-icon="mdiCogOutline"
                :title="t('proxies.groupSettings')"
                @click="model.groupSettings"
              />
              <v-list-item
                v-if="store.outboundName !== 'proxy'"
                :prepend-icon="mdiDeleteOutline"
                :title="t('proxies.deleteGroup')"
                @click="model.removeGroup"
              />
            </v-list>
          </v-menu>
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
        <div v-else-if="view === 'cards'" class="proxies__cards">
          <NodeCard
            v-for="row in listed"
            :key="rowKey(row)"
            :row="row"
            :source="model.sourceName(row)"
            :member="model.isMember(row)"
            :selected="model.isSelected(row)"
            :in-use="!!preferred && rowKey(preferred) === rowKey(row)"
            :disabled="disabled"
            :testing="testing"
            @toggle="model.toggleGroup(row)"
            @action="model.nodeAction(row, $event)"
          />
        </div>
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
              <v-btn
                variant="text"
                :disabled="disabled"
                @click="model.batchMembership(true)"
                >{{ t("proxies.addToGroup") }}</v-btn
              >
              <v-btn
                variant="text"
                :disabled="disabled"
                @click="model.batchMembership(false)"
                >{{ t("proxies.removeFromGroup") }}</v-btn
              >
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
              <v-btn
                variant="text"
                :disabled="disabled"
                @click="model.exportSelected"
                >{{ t("operations.export") }}</v-btn
              >
            </template>
          </div>
          <v-list
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
              :selected="model.isSelected(row)"
              :in-use="!!preferred && rowKey(preferred) === rowKey(row)"
              :checked="selectedKeys.includes(rowKey(row))"
              :disabled="disabled"
              :testing="testing"
              @toggle="model.toggleGroup(row)"
              @check="model.selectRow(row, $event)"
              @action="model.nodeAction(row, $event)"
            />
          </v-list>
        </template>
      </section>
    </template>
  </div>
</template>

<style scoped>
/* no subscriptions yet: one row, icon, words and the action, wrapping when narrow */
.proxies__empty {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px 16px;
  padding: 16px 20px;
}
.proxies__empty-text {
  flex: 1 1 240px;
  min-width: 0;
}
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
/* choice chips without the outline: a quiet pill, tonal when chosen */
.proxies__chip {
  background: rgb(var(--v-theme-surface-container-high));
  color: rgb(var(--v-theme-on-surface));
}
.proxies__chip--on {
  background: rgb(var(--v-theme-secondary-container));
  color: rgb(var(--v-theme-on-secondary-container));
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
.proxies__subscriptions,
.proxies__cards {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
}
.proxies__subscriptions {
  gap: 16px;
}
.proxies__cards {
  gap: 12px;
}
.proxies--expanded .proxies__subscriptions {
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
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
