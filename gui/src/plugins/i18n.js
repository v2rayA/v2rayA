import { createI18n } from "vue-i18n";
import messages from "../locales";

// Ready translated locale messages

// Create i18n instance with options
let locale = "en";
let _lang = typeof localStorage === "undefined" ? "" : localStorage["_lang"];
if (_lang && _lang in messages) {
  locale = _lang;
} else {
  for (let l of window.navigator.languages) {
    l = l.split("-")[0];
    if (l in messages) {
      locale = l;
      break;
    }
  }
}

// Composition mode: components use $t through the global injection, and
// scripts use i18n.global. Plurals are t(key, n); $tc is gone.
// Russian counts in three forms: 1 (не 11), 2–4 (не 12–14), the rest;
// the other locales use vue-i18n's default two or three choices.
const russian = (choice, choicesLength) => {
  if (choicesLength < 3) return choice === 1 ? 0 : 1;
  const mod10 = choice % 10;
  const mod100 = choice % 100;
  if (mod10 === 1 && mod100 !== 11) return 0;
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14)) return 1;
  return 2;
};

const i18n = createI18n({
  locale,
  messages,
  fallbackLocale: "en",
  legacy: false,
  pluralRules: { ru: russian },
});
document.documentElement.lang = locale;

export default i18n;
