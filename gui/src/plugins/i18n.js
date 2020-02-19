import Vue from "vue";
import VueI18n from "vue-i18n";
import messages from "../locales";

Vue.use(VueI18n);

// Ready translated locale messages

// Create VueI18n instance with options
const i18n = new VueI18n({
  locale: "en",
  fallbackLocale: "zh",
  messages
});

export default i18n;
