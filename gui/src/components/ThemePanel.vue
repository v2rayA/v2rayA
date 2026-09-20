<script setup lang="ts">
// The theme choices, as list content: the appearance (auto, light, dark)
// and the seed colour the palettes derive from: a row of preset swatches
// and a rainbow one that unfolds the colour picker for any other seed
// (the Material Theme Builder's hex pastes in). In the shell's menus.
import { computed, ref } from "vue";
import { useI18n } from "vue-i18n";
import {
  mdiCheck,
  mdiThemeLightDark,
  mdiWeatherNight,
  mdiWeatherSunny,
} from "@mdi/js";
import { useAppStore, type ThemePreference } from "@/stores/app";
import { presetSeeds } from "@/theme/scheme";

const { t } = useI18n();
const store = useAppStore();

const modes: { value: ThemePreference; icon: string; key: string }[] = [
  { value: "auto", icon: mdiThemeLightDark, key: "theme.auto" },
  { value: "light", icon: mdiWeatherSunny, key: "theme.light" },
  { value: "dark", icon: mdiWeatherNight, key: "theme.dark" },
];
const seed = computed({
  get: () => store.themeSeed,
  set: (v: string) => store.setThemeSeed(v),
});
const isPreset = computed(() =>
  presetSeeds.some((p) => p.seed === store.themeSeed),
);
const pickerOpen = ref(false);
</script>

<template>
  <v-card-text class="pb-3">
    <p class="md3-label-large text-on-surface-variant mb-2">
      {{ t("theme.appearance") }}
    </p>
    <v-btn-toggle
      :model-value="store.themePreference"
      mandatory
      divided
      variant="outlined"
      rounded="xl"
      density="comfortable"
      selected-class="bg-secondary-container text-on-secondary-container"
      class="w-100"
      @update:model-value="(v: ThemePreference) => store.setTheme(v)"
    >
      <v-btn
        v-for="m in modes"
        :key="m.value"
        :value="m.value"
        :prepend-icon="m.icon"
        class="flex-grow-1 text-none"
      >
        {{ t(m.key) }}
      </v-btn>
    </v-btn-toggle>
    <p class="md3-label-large text-on-surface-variant mt-5 mb-2">
      {{ t("theme.color") }}
    </p>
    <div class="d-flex flex-wrap ga-2">
      <v-btn
        v-for="p in presetSeeds"
        :key="p.seed"
        icon
        size="small"
        variant="flat"
        :style="{ background: p.seed }"
        :aria-label="p.seed"
        :aria-pressed="store.themeSeed === p.seed"
        @click="seed = p.seed"
      >
        <v-icon
          v-if="store.themeSeed === p.seed"
          :icon="mdiCheck"
          class="swatch__check"
        />
      </v-btn>
      <v-btn
        icon
        size="small"
        variant="flat"
        class="swatch--custom"
        :aria-label="t('theme.custom')"
        :aria-pressed="!isPreset"
        @click="pickerOpen = !pickerOpen"
      >
        <v-icon v-if="!isPreset" :icon="mdiCheck" class="swatch__check" />
      </v-btn>
    </div>
    <v-expand-transition>
      <v-color-picker
        v-if="pickerOpen"
        v-model="seed"
        mode="hex"
        :modes="['hex']"
        elevation="0"
        width="100%"
        :canvas-height="120"
        rounded="lg"
        class="mt-3 seed-picker"
      />
    </v-expand-transition>
  </v-card-text>
</template>

<style scoped>
.seed-picker {
  background: transparent;
}
.swatch__check {
  position: relative;
  color: rgba(0, 0, 0, 0.72);
}
.swatch--custom {
  position: relative;
  overflow: hidden;
}
.swatch--custom::before {
  content: "";
  position: absolute;
  inset: 0;
  border-radius: inherit;
  background: conic-gradient(
    #f44336,
    #ff9800,
    #ffeb3b,
    #4caf50,
    #2196f3,
    #9c27b0,
    #f44336
  );
}
</style>
