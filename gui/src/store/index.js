import { createStore } from "vuex";
import i18n from "../plugins/i18n";

export default createStore({
  state: {
    nav: "",
    running: i18n.global.t("common.checkRunning"),
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
