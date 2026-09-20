<script setup lang="ts">
import { computed, onUnmounted, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { useDisplay } from "vuetify";
import {
  mdiPower,
  mdiSpeedometer,
  mdiServerNetwork,
  mdiChevronDown,
  mdiDotsVertical,
  mdiPlus,
  mdiShieldOutline,
  mdiRoutes,
  mdiChartLine,
  mdiRss,
} from "@mdi/js";
import OutboundMenu from "@/components/OutboundMenu.vue";
import TrafficChart from "@/components/TrafficChart.vue";
import { useTraffic } from "@/composables/useTraffic";
import { formatBytes, formatRate } from "@/lib/format";
import { useDashboard } from "./dashboard/model";
import type { SubscriptionAction } from "./proxies/model";
import {
  transparentModes,
  transparentTypes,
  pacModes,
} from "./settings/options";

defineOptions({ name: "DashboardView" });
const { t } = useI18n();
/** the subscription menu's labels, by action */
const actionKeys: Record<string, string> = {
  update: "update",
  edit: "modify",
  share: "share",
  delete: "delete",
};
const { width } = useDisplay();
const wide = computed(() => width.value >= 600);
const {
  store,
  loading,
  busy,
  error,
  members,
  nodeInUse,
  stateLabel,
  canToggle,
  quick,
  quickLoading,
  quickSaving,
  quickDisabled,
  editPorts,
  editRoutingA,
  editGroup,
  importNodes,
  subscriptions,
  selecting,
  testing,
  updatingAll,
  subscriptionsBusy,
  toggleRunning,
  sync,
  setQuick,
  selectNode,
  testMembers,
  updateAll,
  subscriptionAction,
  updating,
} = useDashboard();
const traffic = useTraffic();
const transparentItems = computed(() => transparentModes(t));
const transparentTypeItems = computed(() =>
  transparentTypes(t, {
    lite: store.lite,
    os: store.version?.os ?? "",
    isRoot: store.version?.isRoot ?? false,
    tunSupported: store.version?.tunSupported ?? false,
  }),
);
const pacItems = computed(() => pacModes(t));
const ranked = computed(() =>
  [...members.value].sort(
    (a, b) =>
      Number(b.alive) - Number(a.alive) ||
      (a.ms ?? Infinity) - (b.ms ?? Infinity),
  ),
);
// the latency list fills the height its row gives the tile: as many rows
// as fit, at least four, measured whenever the tile resizes
const latencyList = ref<{ $el: HTMLElement } | null>(null);
const latencyRows = ref(4);
let latencyObserver: ResizeObserver | undefined;
watch(latencyList, (list) => {
  latencyObserver?.disconnect();
  latencyObserver = undefined;
  if (!list?.$el || typeof ResizeObserver === "undefined") return;
  latencyObserver = new ResizeObserver(([entry]) => {
    // a rendered row says how tall rows are; 36 px until one exists
    const row = list.$el.querySelector<HTMLElement>(".v-list-item");
    const rowHeight = row?.offsetHeight || 36;
    latencyRows.value = Math.max(
      4,
      Math.floor(entry.contentRect.height / rowHeight),
    );
  });
  latencyObserver.observe(list.$el);
});
onUnmounted(() => latencyObserver?.disconnect());
const slowest = computed(() =>
  Math.max(1, ...members.value.map((member) => member.ms ?? 0)),
);

// the node picker: "auto" or one member's key; picking calls selectNode
const selectionValue = computed(
  () => members.value.find((m) => m.which.selected)?.key ?? "auto",
);
const selectionItems = computed(() => [
  { value: "auto", title: t("dashboard.autoFastest"), subtitle: "" },
  ...members.value.map((m) => ({
    value: m.key,
    title: m.row.name || m.row.address,
    subtitle: `${m.row.net}  ${m.latency}`,
  })),
]);
function pick(value: string) {
  const outbound = store.outboundName;
  const member = members.value.find((m) => m.key === value);
  void selectNode(member ? member.which : null, outbound);
}
defineExpose({ sync });
</script>

<template>
  <div class="dashboard">
    <v-alert v-if="error" type="error" variant="tonal" class="mb-4">{{
      error
    }}</v-alert>
    <div class="dashboard-grid" :class="{ 'dashboard-grid--wide': wide }">
      <v-card
        color="surface-container-low"
        rounded="xl"
        class="dashboard-wide pa-4"
      >
        <div class="d-flex align-center flex-wrap ga-2 mb-3">
          <v-icon :icon="mdiChartLine" size="20" color="on-surface-variant" />
          <h2 class="md3-title-small flex-grow-1">
            {{ t("dashboard.networkSpeed") }}
          </h2>
          <div
            class="d-flex flex-wrap ga-4 md3-label-large dashboard-figures"
            dir="ltr"
          >
            <span
              :aria-label="`${t('traffic.download')}: ${formatRate(traffic.down.value)}`"
              >↓ {{ formatRate(traffic.down.value) }}</span
            >
            <span
              :aria-label="`${t('traffic.upload')}: ${formatRate(traffic.up.value)}`"
              >↑ {{ formatRate(traffic.up.value) }}</span
            >
          </div>
        </div>
        <div class="dashboard-chart">
          <TrafficChart
            :down="traffic.downSeries.value"
            :up="traffic.upSeries.value"
          />
        </div>
        <p
          class="md3-body-small text-on-surface-variant dashboard-figures mt-2 mb-0"
          dir="ltr"
        >
          {{ t("dashboard.trafficUsage") }}: ↓
          {{ formatBytes(traffic.downTotal.value) }} ↑
          {{ formatBytes(traffic.upTotal.value) }}
        </p>
      </v-card>

      <v-card
        color="surface-container-low"
        rounded="xl"
        class="dashboard-status pa-4"
      >
        <div class="d-flex align-center ga-2 mb-3">
          <v-icon :icon="mdiPower" size="20" color="on-surface-variant" />
          <h2 class="md3-title-small">{{ t("dashboard.status") }}</h2>
        </div>
        <p
          class="md3-headline-small mb-1 d-flex align-center ga-3"
          role="status"
        >
          <span
            class="dashboard-state-dot"
            :class="`dashboard-state-dot--${store.running}`"
            aria-hidden="true"
          />
          {{ stateLabel }}
        </p>
        <p class="md3-body-small text-on-surface-variant mb-5" dir="ltr">
          {{ store.version?.variant || "" }}
          {{ store.version?.coreVersion || "" }}
        </p>
        <div class="dashboard-actions d-flex flex-wrap align-center ga-2">
          <v-btn
            color="primary"
            variant="flat"
            :prepend-icon="mdiPower"
            :loading="busy"
            :disabled="!canToggle"
            @click="toggleRunning"
          >
            {{ t(store.running === "running" ? "v2ray.stop" : "v2ray.start") }}
          </v-btn>
          <v-btn variant="text" @click="store.view = 'logs'">{{
            t("common.log")
          }}</v-btn>
        </div>
      </v-card>

      <v-card
        color="surface-container-low"
        rounded="xl"
        class="dashboard-connection pa-4"
      >
        <div class="d-flex align-center flex-wrap ga-2 mb-3">
          <v-icon
            :icon="mdiServerNetwork"
            size="20"
            color="on-surface-variant"
          />
          <h2 class="md3-title-small flex-grow-1">
            {{ t("dashboard.inUse") }}
          </h2>
          <OutboundMenu variant="chip" />
        </div>
        <v-skeleton-loader
          v-if="loading"
          type="list-item-two-line"
          class="bg-transparent"
        />
        <template v-else-if="members.length">
          <v-menu>
            <template #activator="{ props: menu }">
              <v-btn
                v-bind="menu"
                variant="text"
                class="dashboard-node-name text-none px-2 ms-n2 mb-1"
                :append-icon="mdiChevronDown"
                :disabled="selecting"
                :aria-label="t('dashboard.switchNode')"
              >
                <span class="md3-title-medium dashboard-wrap" dir="auto">{{
                  nodeInUse
                    ? nodeInUse.row.name || nodeInUse.row.address
                    : t("dashboard.autoFastest")
                }}</span>
              </v-btn>
            </template>
            <v-list density="compact" min-width="280">
              <v-list-item
                v-for="item in selectionItems"
                :key="item.value"
                :title="item.title"
                :subtitle="item.subtitle || undefined"
                :active="item.value === selectionValue"
                role="menuitemradio"
                :aria-checked="item.value === selectionValue"
                @click="pick(item.value)"
              />
            </v-list>
          </v-menu>
          <div class="d-flex align-center flex-wrap ga-2 mb-3">
            <span
              v-if="nodeInUse"
              class="md3-body-medium text-on-surface-variant"
              dir="ltr"
              >{{ nodeInUse.row.net }}</span
            >
            <span
              v-if="nodeInUse"
              class="md3-body-medium text-on-surface-variant dashboard-figures"
              dir="ltr"
              >{{ nodeInUse.latency }}</span
            >
            <v-chip
              v-if="nodeInUse?.which.selected"
              size="small"
              variant="tonal"
              >{{ t("dashboard.pinned") }}</v-chip
            >
            <v-chip
              v-else-if="members.length >= 2"
              size="small"
              variant="tonal"
              >{{ t("dashboard.balanced", members.length) }}</v-chip
            >
          </div>
          <div class="dashboard-actions d-flex justify-end">
            <v-btn variant="text" @click="editGroup">{{
              t("dashboard.editGroup")
            }}</v-btn>
          </div>
        </template>
        <template v-else>
          <p class="md3-body-medium mb-4">{{ t("dashboard.emptyGroup") }}</p>
          <v-btn variant="text" @click="editGroup">{{
            t("dashboard.editGroup")
          }}</v-btn>
        </template>
      </v-card>

      <v-card
        color="surface-container-low"
        rounded="xl"
        class="dashboard-transparent pa-4"
        :loading="quickSaving"
      >
        <div class="d-flex align-center ga-2 mb-3">
          <v-icon
            :icon="mdiShieldOutline"
            size="20"
            color="on-surface-variant"
          />
          <h2 class="md3-title-small">{{ t("setting.transparentProxy") }}</h2>
        </div>
        <v-skeleton-loader
          v-if="quickLoading"
          type="list-item"
          class="bg-transparent"
        />
        <template v-else>
          <v-select
            :model-value="quick.transparent"
            :aria-label="t('setting.transparentProxy')"
            :items="transparentItems"
            :disabled="quickDisabled"
            density="compact"
            hide-details
            @update:model-value="(value) => setQuick('transparent', value)"
          />
          <v-expand-transition>
            <v-select
              v-if="quick.transparent !== 'close'"
              :model-value="quick.transparentType"
              :aria-label="t('setting.transparentType')"
              :label="t('setting.transparentType')"
              :items="transparentTypeItems"
              :disabled="quickDisabled"
              density="compact"
              hide-details
              class="mt-3"
              @update:model-value="
                (value) => setQuick('transparentType', value)
              "
            />
          </v-expand-transition>
          <v-switch
            :model-value="quick.portSharing"
            :label="t('setting.portSharingOn')"
            :disabled="quickDisabled"
            hide-details
            class="mt-2"
            @update:model-value="(value) => setQuick('portSharing', !!value)"
          />
          <v-switch
            v-if="store.version?.os === 'linux' && !store.lite"
            :model-value="quick.ipforward"
            :label="t('setting.ipForwardOn')"
            :disabled="quickDisabled"
            hide-details
            @update:model-value="(value) => setQuick('ipforward', !!value)"
          />
        </template>
      </v-card>

      <v-card
        color="surface-container-low"
        rounded="xl"
        class="dashboard-splitting pa-4"
        :loading="quickSaving"
      >
        <div class="d-flex align-center ga-2 mb-3">
          <v-icon :icon="mdiRoutes" size="20" color="on-surface-variant" />
          <h2 class="md3-title-small">{{ t("setting.pacMode") }}</h2>
        </div>
        <v-skeleton-loader
          v-if="quickLoading"
          type="list-item@3"
          class="bg-transparent"
        />
        <v-radio-group
          v-else
          :model-value="quick.pacMode"
          :aria-label="t('setting.pacMode')"
          :disabled="quickDisabled"
          hide-details
          @update:model-value="
            (value) => value !== null && setQuick('pacMode', value)
          "
        >
          <v-radio
            v-for="item in pacItems"
            :key="item.value"
            :value="item.value"
            :label="item.title"
          />
        </v-radio-group>
        <div class="dashboard-actions d-flex flex-wrap ga-2">
          <v-btn variant="text" @click="editRoutingA">{{
            t("routingA.title")
          }}</v-btn>
          <v-btn variant="text" @click="editPorts">{{
            t("customAddressPort.title")
          }}</v-btn>
        </div>
      </v-card>

      <v-card
        color="surface-container-low"
        rounded="xl"
        class="dashboard-latency pa-4"
      >
        <div class="d-flex align-center ga-2 mb-3">
          <v-icon :icon="mdiSpeedometer" size="20" color="on-surface-variant" />
          <h2 class="md3-title-small flex-grow-1">
            {{ t("dashboard.nodeLatency") }}
          </h2>
          <v-btn
            variant="tonal"
            size="small"
            :loading="testing === 'all'"
            :disabled="!members.length || !!testing"
            @click="testMembers"
            >{{ t("dashboard.testLatency") }}</v-btn
          >
        </div>
        <v-skeleton-loader
          v-if="loading"
          type="list-item-two-line@2"
          class="bg-transparent"
        />
        <v-list
          v-else-if="members.length"
          ref="latencyList"
          bg-color="transparent"
          class="pa-0 dashboard-latency-list"
        >
          <v-list-item
            v-for="member in ranked.slice(0, latencyRows)"
            :key="member.key"
            class="px-0 py-1"
            min-height="0"
          >
            <div class="d-flex align-center ga-2 mb-1">
              <span
                v-if="member.key === nodeInUse?.key"
                class="dashboard-active-dot"
                :aria-label="t('dashboard.inUse')"
                role="img"
              />
              <span
                class="md3-body-medium flex-grow-1 dashboard-wrap"
                dir="auto"
                >{{ member.row.name || member.row.address }}</span
              >
              <span
                class="md3-label-medium flex-shrink-0 dashboard-figures"
                dir="ltr"
                >{{ member.latency }}</span
              >
            </div>
            <v-progress-linear
              :model-value="
                member.ms === undefined ? 0 : (member.ms / slowest) * 100
              "
              :aria-label="member.row.name || member.row.address"
              height="4"
              rounded
              color="primary"
            />
          </v-list-item>
        </v-list>
        <p v-else class="md3-body-medium">{{ t("dashboard.emptyGroup") }}</p>
        <div
          v-if="members.length > latencyRows"
          class="dashboard-actions d-flex"
        >
          <v-btn variant="text" @click="store.view = 'proxies'">{{
            t("dashboard.moreMembers", members.length - latencyRows)
          }}</v-btn>
        </div>
      </v-card>

      <v-card
        color="surface-container-low"
        rounded="xl"
        class="dashboard-subscriptions pa-4"
      >
        <div class="d-flex align-center flex-wrap ga-2 mb-3">
          <v-icon :icon="mdiRss" size="20" color="on-surface-variant" />
          <h2 class="md3-title-small flex-grow-1">
            {{ t("common.subscriptions") }}
            <span class="md3-label-medium text-on-surface-variant ms-1">{{
              loading ? "" : subscriptions.length
            }}</span>
          </h2>
          <v-tooltip :text="t('operations.import')">
            <template #activator="{ props }">
              <v-btn
                v-bind="props"
                :icon="mdiPlus"
                size="40"
                variant="text"
                :aria-label="t('operations.import')"
                :disabled="loading"
                @click="importNodes('subscription')"
              />
            </template>
          </v-tooltip>
          <v-btn
            variant="tonal"
            class="dashboard-text-button"
            :loading="updatingAll"
            :disabled="loading || !subscriptions.length || subscriptionsBusy"
            @click="updateAll"
            >{{ t("dashboard.updateAll") }}</v-btn
          >
        </div>
        <v-skeleton-loader
          v-if="loading"
          type="list-item-two-line@2"
          class="bg-transparent"
        />
        <v-list
          v-else-if="subscriptions.length"
          bg-color="transparent"
          class="pa-0"
        >
          <v-list-item
            v-for="subscription in subscriptions.slice(0, 2)"
            :key="subscription.id"
            class="px-0 py-2"
          >
            <p class="md3-title-small mb-1 dashboard-wrap" dir="auto">
              {{ subscription.remarks || subscription.host }}
            </p>
            <p
              class="md3-body-small text-on-surface-variant mb-2 dashboard-wrap"
              dir="auto"
            >
              {{ subscription.summary }}
            </p>
            <v-progress-linear
              v-if="subscription.usage?.percent !== undefined"
              :model-value="subscription.usage.percent"
              :aria-label="subscription.summary"
              height="4"
              rounded
              color="primary"
              class="mb-2"
            />
            <p class="md3-body-small text-on-surface-variant ma-0">
              {{ t("dashboard.updatedAt", { time: subscription.updatedAt }) }}
            </p>
            <template #append>
              <v-menu>
                <template #activator="{ props: menu }">
                  <v-btn
                    v-bind="menu"
                    :icon="mdiDotsVertical"
                    size="40"
                    variant="text"
                    :aria-label="`${t('common.menu')}: ${subscription.remarks || subscription.host}`"
                    :loading="updating === subscription.id"
                    :disabled="subscriptionsBusy"
                  >
                    <v-icon :icon="mdiDotsVertical" size="20" />
                  </v-btn>
                </template>
                <v-list density="compact" min-width="200">
                  <v-list-item
                    v-for="action in ['update', 'edit', 'share', 'delete']"
                    :key="action"
                    :title="t(`operations.${actionKeys[action]}`)"
                    @click="
                      subscriptionAction(
                        subscription,
                        action as SubscriptionAction,
                      )
                    "
                  />
                </v-list>
              </v-menu>
            </template>
          </v-list-item>
        </v-list>
        <div v-else>
          <p class="md3-body-medium my-4">
            {{ t("dashboard.noSubscriptions") }}
          </p>
          <v-btn variant="text" @click="store.view = 'proxies'">{{
            t("operations.import")
          }}</v-btn>
        </div>
        <div v-if="subscriptions.length > 2" class="dashboard-actions d-flex">
          <v-btn variant="text" @click="store.view = 'proxies'">{{
            t("dashboard.moreSubscriptions")
          }}</v-btn>
        </div>
      </v-card>
    </div>
  </div>
</template>

<style scoped>
.dashboard {
  container-type: inline-size;
}
.dashboard-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 16px;
  align-items: stretch;
}
.dashboard-grid > * {
  min-width: 0;
}
/* every tile is a column whose action row sits on the bottom edge, so
   the rows of one grid row line up whatever the tiles hold */
.dashboard-grid > .v-card {
  display: flex;
  flex-direction: column;
}
/* nothing but the action row and the latency list takes the tile's spare height (a v-input would) */
.dashboard-grid > .v-card > :not(.dashboard-actions, .dashboard-latency-list) {
  flex: 0 0 auto;
}
.dashboard-latency-list {
  flex: 1 1 auto;
  min-height: 144px;
  overflow: hidden;
}
.dashboard-actions {
  margin-top: auto;
  padding-top: 12px;
}
.dashboard-grid--wide {
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
}
.dashboard-grid--wide .dashboard-wide {
  grid-column: span 2;
}
/* the subscriptions run the whole row, their rows side by side */
.dashboard-grid--wide .dashboard-full {
  grid-column: 1 / -1;
}
@container (width < 536px) {
  .dashboard-grid--wide {
    grid-template-columns: minmax(0, 1fr);
  }
  .dashboard-grid--wide .dashboard-wide {
    grid-column: span 1;
  }
}
.dashboard-wrap {
  overflow-wrap: anywhere;
}
/* the node's name is the switch: a text button that wraps like a title */
.dashboard-node-name {
  height: auto;
  min-height: 40px;
  max-width: 100%;
  white-space: normal;
  text-align: start;
  justify-content: flex-start;
  align-self: flex-start;
}
/* the core's state at a glance: a 12 dp dot, grey until the core runs */
.dashboard-state-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  flex-shrink: 0;
  background: rgb(var(--v-theme-outline-variant));
}
.dashboard-state-dot--running {
  background: rgb(var(--v-theme-primary));
}
.dashboard-state-dot--paused {
  background: rgb(var(--v-theme-tertiary));
}

.dashboard-chart {
  height: 120px;
}
.dashboard-figures {
  font-variant-numeric: tabular-nums;
}
.dashboard-text-button {
  height: auto;
  min-height: 40px;
  max-width: 100%;
}
.dashboard-text-button :deep(.v-btn__content) {
  white-space: normal;
}
.dashboard-active-dot {
  width: 8px;
  height: 8px;
  flex-shrink: 0;
  border-radius: 50%;
  background: rgb(var(--v-theme-primary));
}
</style>
