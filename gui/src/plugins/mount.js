import Vue from "vue";
import store from "@/store";
import App from "@/App";
import i18n from "@/plugins/i18n";

let vue = null;

Vue.prototype.$remount = () => {
  function f(node) {
    if (!node) {
      return;
    }
    if (typeof node.close == "function") {
      node.close();
      return;
    }
    if (!("$children" in node)) {
      return;
    }
    for (let i in node.$children) {
      f(node.$children[i]);
    }
  }
  f(vue);
  if (vue) {
    // Otherwise the old root keeps its WebSocket, window listeners and
    // matchMedia handler alive next to the new one.
    vue.$destroy();
  }
  vue = new Vue({
    i18n,
    store,
    render: (h) => h(App),
  }).$mount("#app");
};

Vue.prototype.$remount();
