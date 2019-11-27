"use strict";

import Vue from "vue";
import axios from "axios";
import { Modal } from "buefy";
import { ToastProgrammatic } from "buefy";
import ModalLogin from "@/components/modalLogin";

Vue.prototype.$axios = axios;

axios.interceptors.request.use(
  config => {
    if (localStorage.hasOwnProperty("token")) {
      config.headers.Authorization = `${localStorage["token"]}`;
    }
    return config;
  },
  err => {
    if (
      err.name.indexOf("Cross-Origin") >= 0 ||
      err.message.indexOf("Cross-Origin") >= 0
    ) {
      ToastProgrammatic.open({
        message:
          "出现跨域问题，很有可能是因为Firefox等浏览器不支持在https站点访问http资源，如需使用请换用Chrome，或访问备用站点",
        type: "is-warning",
        position: "is-top",
        duration: 5000,
        actionText: "访问备用站点",
        onAction: () => {
          ToastProgrammatic.open({
            message:
              "暂无备用站点，如果您有意提供自动部署的非ssl站点，可以邮件至m@mzz.pub或直接发起pull request",
            type: "is-warning",
            position: "is-top",
            queue: false,
            duration: 5000
          });
        }
      });
    }
    return Promise.reject(err);
  }
);

axios.interceptors.response.use(
  function(res) {
    return res;
  },
  function(err) {
    console.log(Object.assign({}, err));
    if (err.response && err.response.status === 401) {
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
    }
    return Promise.reject(err);
  }
);
