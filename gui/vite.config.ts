import { defineConfig } from "vitest/config";
import vue from "@vitejs/plugin-vue";
import vuetify from "vite-plugin-vuetify";
import { VitePWA } from "vite-plugin-pwa";
import path from "path";
import { Marked, type Tokens, type RendererThis } from "marked";
import type { Plugin } from "vite";
import en from "./src/locales/en.js";

// The in-app documentation is written in Markdown under src/docs and
// compiled to HTML at build time, so the page ships no parser. Headings get
// an id from their text for the table of contents.
function markdown(): Plugin {
  const parser = new Marked({
    gfm: true,
    renderer: {
      heading(this: RendererThis, { tokens, depth }: Tokens.Heading): string {
        const text: string = this.parser.parseInline(tokens);
        const id = text
          .replace(/<[^>]+>/g, "")
          .toLowerCase()
          .replace(/[^\p{L}\p{N}]+/gu, "-")
          .replace(/^-|-$/g, "");
        return `<h${depth} id="${id}">${text}</h${depth}>\n`;
      },
    },
  });
  return {
    name: "v2raya-markdown",
    transform(code, id) {
      if (!id.endsWith(".md")) return null;
      const html = parser.parse(code, { async: false }) as string;
      return { code: `export default ${JSON.stringify(html)};`, map: null };
    },
  };
}

export default defineConfig(({ mode }) => ({
  // vuetify(): per-component style and component imports; nothing of the
  // library ends up in the bundle that a template does not use.
  plugins: [
    markdown(),
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
