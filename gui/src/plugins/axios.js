"use strict";

import Vue from "vue";
import axios from "axios";
import {
  Modal,
  SnackbarProgrammatic,
  ToastProgrammatic,
  ModalProgrammatic
} from "buefy";
import ModalLogin from "@/components/modalLogin";
import { parseURL } from "@/assets/js/utils";
import browser from "@/assets/js/browser";
import modalCustomPorts from "../components/modalCustomPorts";
import i18n from "../plugins/i18n";
import { nanoid } from "nanoid";

Vue.prototype.$axios = axios;

axios.defaults.timeout = 60 * 1000; // timeout: 60秒

axios.interceptors.request.use(
  config => {
    if (localStorage.hasOwnProperty("token")) {
      config.headers.Authorization = `${localStorage["token"]}`;
      config.headers["X-V2raya-Request-Id"] = nanoid();
    }
    return config;
  },
  err => {
    console.log("!", err.name, err.message);
    ToastProgrammatic.open({
      message: err.message,
      type: "is-warning",
      position: "is-top",
      duration: 5000
    });
    return Promise.reject(err);
  }
);

let informed = "";

function informNotRunning(url = localStorage["backendAddress"]) {
  if (informed === url) {
    return;
  }
  informed = url;
  SnackbarProgrammatic.open({
    message: i18n.t("axios.messages.optimizeBackend"),
    type: "is-primary",
    queue: false,
    duration: 10000,
    position: "is-top",
    actionText: i18n.t("operations.yes"),
    onAction: () => {
      // this.showCustomPorts = true;
      ModalProgrammatic.open({
        component: modalCustomPorts,
        hasModalCard: true,
        customClass: "modal-custom-ports"
      });
    }
  });
  SnackbarProgrammatic.open({
    message: i18n.t("axios.messages.noBackendFound", { url }),
    type: "is-warning",
    queue: false,
    position: "is-top",
    duration: 10000,
    actionText: i18n.t("operations.helpManual"),
    onAction: () => {
      window.open(i18n.t("axios.urls.usage"), "_blank");
    }
  });
}

axios.interceptors.response.use(
  function(res) {
    return res;
  },
  function(err) {
    console.log("!!", err.name, err.message);
    console.log(Object.assign({}, err));
    if (err.code === "ECONNABORTED" && err.isAxiosError) {
      return Promise.reject(err);
    }
    let u, host;
    if (err.config) {
      u = parseURL(err.config.url);
      host = u.host;
    }
    if (err.response && err.response.status === 401) {
      //401未授权
      new Vue({
        components: { Modal, ModalLogin },
        data: () => ({
          show: true
        }),
        render() {
          let first =
            err.response.data &&
            err.response.data.data &&
            err.response.data.data.first === true;
          return (
            <b-modal
              active={this.show}
              trap-focus={true}
              has-modal-card={true}
              aria-role="dialog"
              aria-modal={true}
              full-screen={false}
              style="z-index:1000"
              class="modal-login"
              id="login"
            >
              <ModalLogin
                first={first}
                onClose={() => {
                  this.show = false;
                }}
              />
            </b-modal>
          );
        }
      }).$mount("#login");
    } else if (
      location.protocol.substr(0, 5) === "https" &&
      u.protocol === "http"
    ) {
      //https前端通信http后端
      let msg = i18n.t("axios.messages.cannotCommunicate.0");
      if (host === "localhost" || host === "local" || host === "127.0.0.1") {
        if (browser.versions.webKit) {
          //Chrome等webkit内核浏览器允许访问http://localhost，只有可能是服务端未启动
          informNotRunning(u.source.replace(u.relative, ""));
          return;
        }
        if (browser.versions.gecko) {
          msg = i18n.t("axios.messages.cannotCommunicate.1");
        }
      }
      SnackbarProgrammatic.open({
        message: msg,
        type: "is-warning",
        position: "is-top",
        queue: false,
        duration: 10000,
        actionText: i18n.t("operations.switchSite"),
        onAction: () => {
          window.open("http://v.v2raya.org", "_self");
        }
      });
      SnackbarProgrammatic.open({
        message: i18n.t("axios.messages.optimizeBackend"),
        type: "is-primary",
        queue: false,
        duration: 10000,
        position: "is-top",
        actionText: i18n.t("operations.yes"),
        onAction: () => {
          // this.showCustomPorts = true;
          ModalProgrammatic.open({
            component: modalCustomPorts,
            hasModalCard: true,
            customClass: "modal-custom-ports"
          });
        }
      });
    } else if (
      (err.message && err.message === "Network Error") ||
      (err.config && err.config.url === "/api/version")
    ) {
      informNotRunning(u.source.replace(u.relative, ""));
    } else {
      //其他错误
      if (
        !err.message ||
        (err.message && err.message.indexOf("404") >= 0) ||
        (err.response && err.response.status === 404)
      ) {
        //接口不存在，或是正常错误（如取消），可能服务端是老旧版本，不管
        return Promise.reject(err);
      }
      console.log("!other");
      ToastProgrammatic.open({
        message: err,
        type: "is-warning",
        position: "is-top",
        queue: false,
        duration: 5000
      });
    }
    return Promise.reject(err);
  }
);

export default axios;
