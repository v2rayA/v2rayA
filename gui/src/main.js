import "@/plugins/buefy";
import "@/plugins/axios";
import Vue from "vue";
import App from "./App.vue";
import store from "./store";
import "normalize.css";
import "@/registerServiceWorker";

Vue.config.productionTip = false;

new Vue({
  store,
  render: h => h(App)
}).$mount("#app");
