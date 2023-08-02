import Vue from "vue";
import Vuex from "vuex";
import i18n from "../plugins/i18n";

Vue.use(Vuex);
export default new Vuex.Store({
  state: {
    nav: "",
    running: i18n.t("common.checkRunning"),
    connectedServer: {},
  },
  mutations: {
    NAV(state, val) {
      state.nav = val;
    },
    RUNNING(state, val) {
      state.running = val;
    },
    CONNECTED_SERVER(state, val) {
      state.connectedServer = val;
    },
  },
  actions: {},
  modules: {},
});
