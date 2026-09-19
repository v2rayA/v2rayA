import { defineConfig } from "vitest/config";
import vue from "@vitejs/plugin-vue";
import vuetify from "vite-plugin-vuetify";
import { VitePWA } from "vite-plugin-pwa";
import path from "path";
import en from "./src/locales/en.js";

export default defineConfig(({ mode }) => ({
  // vuetify(): per-component style and component imports; nothing of the
  // library ends up in the bundle that a template does not use.
  plugins: [
    vue(),
    vuetify({ autoImport: true }),
    VitePWA({
      registerType: "autoUpdate",
      manifest: {
        name: "v2rayA",
        short_name: "v2rayA",
        description: en.about.intro,
        display: "standalone",
        start_url: "./",
        scope: "./",
        theme_color: "#fffbff",
        background_color: "#fffbff",
        icons: [
          {
            src: "pwa-192.png",
            sizes: "192x192",
            type: "image/png",
            purpose: "any maskable",
          },
          {
            src: "pwa-512.png",
            sizes: "512x512",
            type: "image/png",
            purpose: "any maskable",
          },
        ],
      },
      workbox: {
        globPatterns: ["**/*.{js,css,html,ico,png,svg,woff,woff2,ttf}"],
        globIgnores: ["**/api/**"],
        navigateFallback: "index.html",
        navigateFallbackDenylist: [/\/api(?:\/|$)/],
        runtimeCaching: [
          {
            urlPattern: ({ url }) =>
              url.pathname.includes("/api/") || url.pathname.endsWith("/api"),
            handler: "NetworkOnly",
          },
        ],
      },
    }),
  ],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "src"),
    },
    extensions: [".mjs", ".js", ".ts", ".jsx", ".tsx", ".json", ".vue"],
  },
  server: {
    port: 8081,
  },
  build: {
    outDir: process.env.OUTPUT_DIR || "../web",
    sourcemap: false,
    assetsDir: "static",
    emptyOutDir: true,
  },
  base: process.env.publicPath || (mode === "production" ? "./" : "/"),
  test: {
    environment: "node",
    include: ["src/**/*.spec.ts"],
    // the colour library's ESM has extensionless imports and Vuetify's components
    // import their CSS; Node resolves neither, Vite does
    server: {
      deps: { inline: ["@material/material-color-utilities", "vuetify"] },
    },
  },
}));
