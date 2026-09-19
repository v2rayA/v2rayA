/// <reference types="vite-plugin-pwa/client" />
import { registerSW } from "virtual:pwa-register";

if (import.meta.env.PROD) {
  registerSW({ immediate: true });
}
