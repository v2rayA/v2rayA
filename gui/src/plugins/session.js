import { createApp } from "vue";
import store from "@/store";
import App from "@/App";
import i18n from "@/plugins/i18n";
import Buefy from "@/plugins/buefy";
import VirtualScroller from "@/plugins/virtual-scroll";
import { install as installAxios } from "@/plugins/axios";
import { install as installDayjs } from "@/plugins/dayjs";

// Programmatic instances that mount on <body> (modals, loadings, snackbars)
// outlive the root tree, so restart() must close them before unmounting.
// Commit 3 registers the loading/snackbar/modal handles here.
const programmaticHandles = [];

let app = null;

export function buildApp() {
  app = createApp(App);
  app.use(store);
  app.use(i18n);
  app.use(Buefy);
  app.use(VirtualScroller);
  installAxios(app);
  installDayjs(app);
  // Components call this.$remount() (same API as the Vue 2.7 prototype
  // method). Exposing it here — instead of importing session.js from inside
  // the component tree — avoids a circular import (session -> App -> component
  // -> session) that would put modalCustomPorts in the temporal dead zone.
  app.config.globalProperties.$remount = restart;
  return app;
}

export function registerProgrammatic(handle) {
  programmaticHandles.push(handle);
}

export function restart() {
  for (const handle of programmaticHandles) {
    if (handle && typeof handle.close === "function") {
      handle.close();
    }
  }
  if (app) {
    // app.unmount() runs beforeUnmount on the tree, which closes the
    // WebSocket, window listeners and matchMedia handler in App.vue/node.vue.
    app.unmount();
  }
  app = buildApp();
  app.mount("#app");
}
