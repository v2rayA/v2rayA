"use strict";

import Vue from "vue";
import axios from "axios";

Vue.prototype.$axios = axios;

axios.interceptors.response.use(
  function(response) {
    return response;
  },
  function(error) {
    return Promise.reject(error);
  }
)