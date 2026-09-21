<script setup lang="ts">
// The subscriptions page: the cards with their quota and node counts, the
// auto-update settings, and importing one. It shares the proxies model,
// which loads the subscriptions with the nodes.
import { computed, onMounted } from "vue";
import { useI18n } from "vue-i18n";
import { useDisplay } from "vuetify";
import { mdiCogOutline, mdiRss, mdiTrayArrowDown } from "@mdi/js";
import { useProxies } from "./proxies/model";
import SubscriptionCard from "./proxies/SubscriptionCard.vue";

defineOptions({ name: "SubscriptionsView" });
const { t } = useI18n();
const { width } = useDisplay();
const expanded = computed(() => width.value >= 840);
const model = useProxies();
const { subscriptions, loading, loadError, busy, sync } = model;
const disabled = computed(() => busy.value || loading.value);

onMounted(sync);
defineExpose({ sync });
</script>

<template>
  <div class="subs" :class="{ 'subs--expanded': expanded }">
    <div class="d-flex flex-wrap align-center ga-2 mb-4">
      <v-btn
        variant="text"
        :prepend-icon="mdiCogOutline"
        :disabled="disabled"
        @click="model.subscriptionSettings"
        >{{ t("proxies.autoUpdate") }}</v-btn
      >
      <v-spacer />
      <v-btn
        variant="flat"
        color="primary"
        :prepend-icon="mdiTrayArrowDown"
        :disabled="disabled"
        @click="model.importNodes('subscription')"
        >{{ t("proxies.importSubscription") }}</v-btn
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
      v-if="loading && !subscriptions.length"
      type="card, card"
      class="bg-transparent"
    />
    <div v-else-if="subscriptions.length" class="subs__cards">
      <SubscriptionCard
        v-for="subscription in subscriptions"
        :key="subscription.address"
        :subscription="subscription"
        :disabled="disabled"
        @action="model.subscriptionAction(subscription, $event)"
      />
    </div>
    <v-sheet
      v-else-if="!loadError"
      color="surface-container-low"
      rounded="xl"
      class="subs__empty"
    >
      <v-icon
        :icon="mdiRss"
        size="28"
        color="on-surface-variant"
        class="flex-shrink-0"
      />
      <div class="subs__empty-text">
        <p class="md3-title-medium ma-0">{{ t("proxies.noSubscriptions") }}</p>
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
</template>

<style scoped>
.subs__cards {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 16px;
}
.subs--expanded .subs__cards {
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
}
/* no subscriptions yet: one row, icon, words and the action, wrapping when narrow */
.subs__empty {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px 16px;
  padding: 16px 20px;
}
.subs__empty-text {
  flex: 1 1 240px;
  min-width: 0;
}
</style>
