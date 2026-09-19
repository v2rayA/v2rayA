<script setup lang="ts">
// The theme choices, as list content: the appearance (auto, light, dark)
// and the seed colour the palettes derive from: a row of preset swatches
// and a rainbow one that unfolds the colour picker for any other seed
// (the Material Theme Builder's hex pastes in). In the shell's menus.
import { computed, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import {
  mdiCheck,
  mdiThemeLightDark,
  mdiWeatherNight,
  mdiWeatherSunny,
} from "@mdi/js";
import { useAppStore, type ThemePreference } from "@/stores/app";
import { hueOf, isSeed, presetSeeds, seedFromHue } from "@/theme/scheme";

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
// the custom seed: a hue slider (Material's slider, the track painted with
// the hues) and the hex, so a Theme Builder seed pastes in
const hue = computed({
  get: () => hueOf(store.themeSeed),
  set: (h: number) => store.setThemeSeed(seedFromHue(h)),
});
const hex = ref(store.themeSeed);
watch(
  () => store.themeSeed,
  (v) => (hex.value = v),
);
function applyHex() {
  const value = hex.value.trim().startsWith("#")
    ? hex.value.trim()
    : `#${hex.value.trim()}`;
  if (isSeed(value)) store.setThemeSeed(value);
  else hex.value = store.themeSeed;
}
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
      <div v-if="pickerOpen" class="mt-3">
        <v-slider
          v-model="hue"
          :min="0"
          :max="359"
          :step="1"
          hide-details
          track-size="8"
          thumb-size="20"
          class="hue-slider"
          :aria-label="t('theme.custom')"
        />
        <v-text-field
          v-model="hex"
          label="HEX"
          variant="outlined"
          density="compact"
          hide-details
          dir="ltr"
          maxlength="7"
          class="mt-3 hue-hex"
          @change="applyHex"
          @keydown.enter="applyHex"
        >
          <template #prepend-inner>
            <span class="hue-swatch" aria-hidden="true" />
          </template>
        </v-text-field>
      </div>
    </v-expand-transition>
  </v-card-text>
</template>

<style scoped>
/* the hue slider's track shows the hues it selects from */
.hue-slider :deep(.v-slider-track__background),
.hue-slider :deep(.v-slider-track__fill) {
  background: linear-gradient(
    to right,
    hsl(0 80% 60%),
    hsl(60 80% 60%),
    hsl(120 80% 60%),
    hsl(180 80% 60%),
    hsl(240 80% 60%),
    hsl(300 80% 60%),
    hsl(360 80% 60%)
  );
  opacity: 1;
}
.hue-hex {
  max-width: 200px;
}
.hue-swatch {
  display: inline-block;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: rgb(var(--v-theme-primary));
  border: 1px solid rgb(var(--v-theme-outline-variant));
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
