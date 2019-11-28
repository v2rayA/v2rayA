"use strict";

import Vue from "vue";
import axios from "axios";
import { Modal, SnackbarProgrammatic, ToastProgrammatic } from "buefy";
import ModalLogin from "@/components/modalLogin";
import { parseURL } from "../assets/js/utils";

Vue.prototype.$axios = axios;

axios.interceptors.request.use(
  config => {
    if (localStorage.hasOwnProperty("token")) {
      config.headers.Authorization = `${localStorage["token"]}`;
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

axios.interceptors.response.use(
  function(res) {
    return res;
  },
  function(err) {
    console.log("!!", err.name, err.message);
    console.log(Object.assign({}, err));
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
      parseURL(err.config.url).protocol === "http"
    ) {
      //https前端通信http后端
      SnackbarProgrammatic.open({
        message:
          "当前站点为https站点，您设置的服务端地址为http，由于浏览器限制，将无法进行通信",
        type: "is-warning",
        position: "is-top",
        queue: false,
        duration: 8000,
        actionText: "查看解决方案",
        onAction: () => {
          window.open(
            "https://github.com/mzz2017/V2RayA/issues/7#issuecomment-559546844",
            "_blank"
          );
        }
      });
    } else {
      //其他错误
      console.log("!other");
      ToastProgrammatic.open({
        message: err.message,
        type: "is-warning",
        position: "is-top",
        queue: false,
        duration: 5000
      });
    }
    return Promise.reject(err);
  }
);
