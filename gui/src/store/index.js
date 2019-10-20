import Vue from "vue";
import Vuex from "vuex";

Vue.use(Vuex);

export default new Vuex.Store({
  state: {
    nav: ""
  },
  mutations: {
    NAV(state, val) {
      state.nav = val;
    }
  },
  actions: {},
  modules: {}
});
