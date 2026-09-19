import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import path from "path";
import lucideSubset from "./build/lucide-subset.mjs";

export default defineConfig(({ mode }) => ({
  plugins: [lucideSubset(), vue()],
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
}));
