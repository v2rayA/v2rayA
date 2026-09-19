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

// Buefy 3 exports SnackbarProgrammatic and friends as classes bound to an
// app; module-level code (axios interceptors, the network inspector) has no
// component to take $buefy from, so it takes the root app's.
export function buefy() {
  return app.config.globalProperties.$buefy;
}

export function registerProgrammatic(handle) {
  // A closed instance unmounts itself and leaves the document; drop those
  // so the list only holds what restart() still has to close.
  for (let i = programmaticHandles.length - 1; i >= 0; i--) {
    const el = programmaticHandles[i].$el;
    if (!el || !document.body.contains(el)) {
      programmaticHandles.splice(i, 1);
    }
  }
  programmaticHandles.push(handle);
}

// Register + return the handle of a programmatic instance opened through the
// component's own $buefy (which carries the app context). The caller still
// gets the handle back for its own .close().
export function openModal(ctx, opts) {
  const handle = ctx.$buefy.modal.open(opts);
  registerProgrammatic(handle);
  return handle;
}

export function openLoading(ctx) {
  // Buefy 3 reads options.onClose when the overlay closes and does not
  // tolerate a missing options object.
  const handle = ctx.$buefy.loading.open({});
  registerProgrammatic(handle);
  return handle;
}

export function restart() {
  for (const handle of programmaticHandles) {
    if (handle && typeof handle.close === "function") {
      handle.close();
    }
  }
  programmaticHandles.length = 0;
  if (app) {
    // app.unmount() runs beforeUnmount on the tree, which closes the
    // WebSocket, window listeners and matchMedia handler in App.vue/node.vue.
    app.unmount();
  }
  app = buildApp();
  app.mount("#app");
}
