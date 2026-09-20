<script setup lang="ts">
import { computed } from "vue";
import { useI18n } from "vue-i18n";
import { mdiContentCopy, mdiFileReplaceOutline } from "@mdi/js";
import HighlightedCode from "./HighlightedCode.vue";
import { presets } from "./presets";

defineOptions({ name: "RoutingReference" });
defineProps<{ disabled?: boolean }>();
const emit = defineEmits<{
  insert: [example: string];
  replace: [example: string];
}>();
const { t } = useI18n();
type Example = string | { label: string; code: string };
const codeOf = (example: Example) =>
  typeof example === "string" ? example : example.code;
const sections = computed<
  { title: string; description: string; examples: Example[] }[]
>(() => [
  {
    title: t("routingA.reference.format.title"),
    description: t("routingA.reference.format.description"),
    examples: ["# HTTPS\nport(443) && network(tcp) -> proxy"],
  },
  {
    title: t("routingA.reference.domain.title"),
    description: t("routingA.reference.domain.description"),
    examples: [
      "domain(full: example.com, domain: example.org) -> proxy",
      "domain(contains: example, regexp: '.*\\.example\\.net') -> proxy",
      "domain(geosite: cn) -> direct",
    ],
  },
  {
    title: t("routingA.reference.ip.title"),
    description: t("routingA.reference.ip.description"),
    examples: [
      "ip(1.2.3.4, 10.0.0.0/8) -> direct",
      "ip(geoip: private) -> direct",
      'ip("2001:db8::/32") -> direct',
    ],
  },
  {
    title: t("routingA.reference.ports.title"),
    description: t("routingA.reference.ports.description"),
    examples: [
      "port(80, 443) && network(tcp, udp) -> proxy",
      "port(1-1023) && protocol(http) -> direct",
      "source(192.168.1.0/24) -> direct",
    ],
  },
  {
    title: t("routingA.reference.outbound.title"),
    description: t("routingA.reference.outbound.description"),
    examples: [
      "default: proxy",
      "outbound: name = socks(address: 127.0.0.1, port: 10800)",
      "outbound: authenticated = http(address: 127.0.0.1, port: 8080, user: 'username', pass: 'password')",
    ],
  },
  {
    title: t("routingA.reference.presets.title"),
    description: t("routingA.reference.presets.description"),
    examples: presets.map((preset) => ({
      label: t(`routingA.reference.presets.${preset.key}`),
      code: preset.code,
    })),
  },
]);
</script>

<template>
  <v-sheet
    class="routing-reference pa-4"
    color="surface-container-low"
    rounded="lg"
    :aria-label="t('routingA.reference.title')"
  >
    <v-expansion-panels variant="accordion" flat :model-value="0">
      <v-expansion-panel
        v-for="section in sections"
        :key="section.title"
        bg-color="surface-container-low"
      >
        <v-expansion-panel-title class="md3-title-small px-0">{{
          section.title
        }}</v-expansion-panel-title>
        <v-expansion-panel-text>
          <p class="md3-body-small mb-4">{{ section.description }}</p>
          <div
            v-for="example in section.examples"
            :key="codeOf(example)"
            class="routing-reference__example mb-4"
          >
            <p
              v-if="typeof example !== 'string'"
              class="md3-label-large routing-reference__label"
            >
              {{ example.label }}
            </p>
            <HighlightedCode
              :text="codeOf(example)"
              class="routing-reference__code"
            />
            <div class="routing-reference__actions">
              <v-tooltip :text="t('routingA.insert')">
                <template #activator="{ props: tip }">
                  <v-btn
                    v-bind="tip"
                    :icon="mdiContentCopy"
                    variant="text"
                    size="32"
                    :aria-label="t('routingA.insert')"
                    :disabled="disabled"
                    @click="emit('insert', codeOf(example))"
                  />
                </template>
              </v-tooltip>
              <v-tooltip
                v-if="typeof example !== 'string'"
                :text="t('routingA.replace')"
              >
                <template #activator="{ props: tip }">
                  <v-btn
                    v-bind="tip"
                    :icon="mdiFileReplaceOutline"
                    variant="text"
                    size="32"
                    :aria-label="t('routingA.replace')"
                    :disabled="disabled"
                    @click="emit('replace', codeOf(example))"
                  />
                </template>
              </v-tooltip>
            </div>
          </div>
        </v-expansion-panel-text>
      </v-expansion-panel>
    </v-expansion-panels>
  </v-sheet>
</template>

<style scoped>
.routing-reference {
  min-width: 0;
  max-height: 60vh;
  overflow: auto;
}
.routing-reference :deep(.v-expansion-panel-text__wrapper) {
  padding: 8px 0 0;
}
.routing-reference__example {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 32px;
  align-items: start;
  gap: 8px;
}
.routing-reference__label {
  grid-column: 1 / -1;
  margin: 0;
}
.routing-reference__actions {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.routing-reference__code {
  overflow: auto;
  padding-block: 8px;
}
</style>
