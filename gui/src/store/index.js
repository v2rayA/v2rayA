import Vue from "vue";
import Vuex from "vuex";
import CONST from "@/assets/js/const";

Vue.use(Vuex);

export default new Vuex.Store({
  state: {
    nav: "",
    running: CONST.INSPECTING_RUNNING,
    connectedServer: {}
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
    }
  },
  actions: {},
  modules: {}
});
