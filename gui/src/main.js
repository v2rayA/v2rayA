import "@/plugins/buefy";
import Vue from "vue";
import App from "./App.vue";
import router from "./router";
import store from "./store";
import axios from "axios";
import "normalize.css";
import "@/assets/font_1467288_zafhp5y171/iconfont.css";

Vue.prototype.$axios = axios;

Vue.config.productionTip = false;

new Vue({
  router,
  store,
  render: h => h(App)
}).$mount("#app");
