import "vuetify/styles";
import { createVuetify, type ThemeDefinition } from "vuetify";
import { md3 } from "vuetify/blueprints";
import { aliases, mdi } from "vuetify/iconsets/mdi-svg";
import { en, fa, ko, pt, ru, zhHans } from "vuetify/locale";
import "./typography.scss";
import "./components.scss";
import { brandSeed, schemeColors } from "./scheme";

// The interface's locale codes (localStorage _lang, the locales/ files)
// and Vuetify's own message sets for its built-in texts (the table's
// pagination, the date picker).
export const vuetifyLocales: Record<string, string> = {
  en: "en",
  zh: "zhHans",
  fa: "fa",
  ru: "ru",
  pt: "pt",
  ko: "ko",
};

// The two themes start from the brand seed; the shell recomputes them when
// the user picks another seed (theme/scheme.ts).
export const light: ThemeDefinition = {
  dark: false,
  colors: schemeColors(brandSeed, false),
};
export const dark: ThemeDefinition = {
  dark: true,
  colors: schemeColors(brandSeed, true),
};

// The theme name Vuetify starts with; App switches it from the stored
// preference (auto/light/dark) and the OS query.
export const vuetify = createVuetify({
  blueprint: md3,
  icons: { defaultSet: "mdi", aliases, sets: { mdi } },
  locale: {
    locale: "en",
    fallback: "en",
    messages: { en, zhHans, fa, ko, pt, ru },
  },
  theme: {
    defaultTheme: "light",
    themes: { light, dark },
    // the state layer opacities MD3 specifies; the blueprint does not set them
    variations: {
      colors: ["primary", "secondary", "tertiary"],
      lighten: 1,
      darken: 1,
    },
  },
  defaults: {
    global: { ripple: false },
    VBtn: { variant: "flat", height: 40 },
    VCard: { elevation: 0, rounded: "lg" },
    VDialog: { scrim: "on-surface" },
    // Material's menu: surface-container at elevation 2, 8 dp from its anchor
    VMenu: {
      offset: 8,
      VList: { bgColor: "surface-container", elevation: 2, rounded: "lg" },
      VCard: { color: "surface-container", elevation: 2, rounded: "lg" },
    },
    VTooltip: { location: "top" },
    VTextField: { variant: "outlined", density: "comfortable" },
    VSelect: { variant: "outlined", density: "comfortable" },
    VCombobox: { variant: "outlined", density: "comfortable" },
    VTextarea: { variant: "outlined", density: "comfortable" },
    VSwitch: { inset: true, color: "primary" },
    // Material's chips: 8 dp corners; a selected filter chip is secondary-container
    VChip: { rounded: "lg" },
    VChipGroup: {
      selectedClass: "bg-secondary-container text-on-secondary-container",
    },
  },
});
