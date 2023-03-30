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
  vue = new Vue({
    i18n,
    store,
    render: (h) => h(App),
  }).$mount("#app");
};

Vue.prototype.$remount();
