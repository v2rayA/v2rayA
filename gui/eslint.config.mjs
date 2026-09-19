import pluginVue from "eslint-plugin-vue";
import {
  defineConfigWithVueTs,
  vueTsConfigs,
} from "@vue/eslint-config-typescript";
import prettier from "@vue/eslint-config-prettier";

export default defineConfigWithVueTs(
  { files: ["**/*.{ts,mts,js,mjs,vue}"] },
  { ignores: ["dist/**", "node_modules/**", "build/**"] },
  pluginVue.configs["flat/recommended"],
  vueTsConfigs.recommended,
  prettier,
  {
    languageOptions: {
      globals: {
        apiRoot: "readonly",
        localStorage: "readonly",
        sessionStorage: "readonly",
        window: "readonly",
        document: "readonly",
        navigator: "readonly",
        console: "readonly",
        setTimeout: "readonly",
        clearTimeout: "readonly",
        setInterval: "readonly",
        clearInterval: "readonly",
        fetch: "readonly",
        Blob: "readonly",
        URL: "readonly",
        WebSocket: "readonly",
        Image: "readonly",
        FileReader: "readonly",
        HTMLElement: "readonly",
        Event: "readonly",
        KeyboardEvent: "readonly",
        MouseEvent: "readonly",
        requestAnimationFrame: "readonly",
        cancelAnimationFrame: "readonly",
        location: "readonly",
        history: "readonly",
        performance: "readonly",
        crypto: "readonly",
        Element: "readonly",
        HTMLInputElement: "readonly",
        Node: "readonly",
        MutationObserver: "readonly",
        ResizeObserver: "readonly",
        getComputedStyle: "readonly",
        alert: "readonly",
        confirm: "readonly",
        prompt: "readonly",
        atob: "readonly",
        btoa: "readonly",
        AbortController: "readonly",
        Notification: "readonly",
        process: "readonly",
        __dirname: "readonly",
      },
    },
    rules: {
      "no-console": "off",
      "no-debugger": "off",
      "@typescript-eslint/no-explicit-any": "warn",
      // Vuetify names table cell slots item.<column>
      "vue/valid-v-slot": ["error", { allowModifiers: true }],
    },
  },
  {
    // The files the visual redo replaces one by one. They predate the
    // TypeScript preset; its findings there are warnings until the file is
    // rewritten, and errors in everything new.
    files: [
      "src/*.vue",
      "src/*.js",
      "src/components/**",
      "src/plugins/**",
      "src/assets/js/**",
      "src/store/**",
      "src/locales/**",
    ],
    rules: {
      "vue/multi-word-component-names": "off",
      "vue/block-lang": "off",
      "vue/no-mutating-props": "warn",
      "prefer-const": "warn",
      "prefer-rest-params": "warn",
      "@typescript-eslint/no-this-alias": "warn",
      "@typescript-eslint/no-unused-expressions": "warn",
      "@typescript-eslint/no-unused-vars": "warn",
    },
  },
);
