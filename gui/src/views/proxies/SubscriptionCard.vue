<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import dayjs from "dayjs";
import {
  mdiDotsVertical,
  mdiRefresh,
  mdiPencilOutline,
  mdiShareVariantOutline,
  mdiDeleteOutline,
} from "@mdi/js";
import type { TouchSubscription } from "@/api/types";
import { parseQuota } from "@/lib/quota";
import type { SubscriptionAction } from "./model";
defineOptions({ name: "SubscriptionCard" });
const props = defineProps<{
  subscription: TouchSubscription;
  disabled: boolean;
}>();
const emit = defineEmits<{ action: [action: SubscriptionAction] }>();
const { t } = useI18n();
const usage = computed(() => parseQuota(props.subscription.info));
</script>
<template>
  <v-card
    color="surface-container-high"
    variant="flat"
    rounded="xl"
    class="subscription-card"
    :loading="disabled"
  >
    <v-card-item class="pa-4">
      <v-card-title class="md3-title-medium text-wrap" dir="auto">{{
        subscription.remarks || subscription.host
      }}</v-card-title>
      <template #append>
        <v-menu>
          <template #activator="{ props: menu }">
            <v-btn
              v-bind="menu"
              :icon="mdiDotsVertical"
              variant="text"
              size="40"
              class="subscription-card__menu"
              :aria-label="t('operations.name')"
              :disabled="disabled"
            >
              <v-icon :icon="mdiDotsVertical" size="18" />
              <v-tooltip activator="parent" location="top">{{
                t("operations.name")
              }}</v-tooltip>
            </v-btn>
          </template>
          <v-list :disabled="disabled">
            <v-list-item
              :prepend-icon="mdiRefresh"
              :title="t('operations.update')"
              @click="emit('action', 'update')"
            />
            <v-list-item
              :prepend-icon="mdiPencilOutline"
              :title="t('operations.modify')"
              @click="emit('action', 'edit')"
            />
            <v-list-item
              :prepend-icon="mdiShareVariantOutline"
              :title="t('operations.share')"
              @click="emit('action', 'share')"
            />
            <v-list-item
              :prepend-icon="mdiDeleteOutline"
              :title="t('operations.delete')"
              @click="emit('action', 'delete')"
            />
          </v-list>
        </v-menu>
      </template>
    </v-card-item>
    <v-card-text class="px-4 pb-4 pt-0">
      <template v-if="usage">
        <div
          class="d-flex flex-wrap justify-space-between ga-2 md3-body-medium mb-2"
        >
          <span dir="ltr">{{
            usage.used && usage.total
              ? `${usage.used} / ${usage.total}`
              : usage.used || usage.total
          }}</span>
          <span class="text-on-surface-variant" dir="ltr">{{
            usage.expires
          }}</span>
        </div>
        <v-progress-linear
          v-if="usage.percent !== undefined"
          :model-value="usage.percent"
          color="primary"
          bg-color="surface-container-highest"
          rounded
          height="6"
          class="mb-4"
        />
      </template>
      <p v-else-if="subscription.info" class="md3-body-medium mb-4">
        {{ subscription.info }}
      </p>
      <div class="d-flex flex-wrap ga-2">
        <v-chip size="small" variant="tonal"
          >{{ t("subscription.numberServers") }}:
          {{ subscription.servers.length }}</v-chip
        >
        <v-chip v-if="subscription.status" size="small" variant="outlined"
          ><span dir="ltr">{{
            dayjs(subscription.status).format("YYYY-MM-DD HH:mm")
          }}</span></v-chip
        >
      </div>
    </v-card-text>
  </v-card>
</template>
<style scoped>
.subscription-card {
  overflow-wrap: anywhere;
}
.subscription-card__menu {
  margin: 4px;
}
.subscription-card__menu::after {
  content: "";
  position: absolute;
  inset: -4px;
}
</style>
