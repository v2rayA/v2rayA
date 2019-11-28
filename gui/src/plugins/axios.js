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
import { parseURL } from "../assets/js/utils";
import browser from "@/assets/js/browser";
import modalCustomPorts from "../components/modalCustomPorts";

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

let informed = false;

function informNotRunning() {
  if (informed) {
    return;
  }
  informed = true;
  SnackbarProgrammatic.open({
    message: "您是否需要调整服务端地址？",
    type: "is-primary",
    queue: false,
    indefinite: true,
    position: "is-top",
    actionText: "是",
    onAction: () => {
      // this.showCustomPorts = true;
      ModalProgrammatic.open({
        parent: this,
        component: modalCustomPorts,
        hasModalCard: true,
        customClass: "modal-custom-ports"
      });
    }
  });
  SnackbarProgrammatic.open({
    message: `未在 ${
      localStorage["backendAddress"]
    } 检测到V2RayA服务端，请确定V2RayA已正确安装且配置正确`,
    type: "is-warning",
    queue: false,
    position: "is-top",
    indefinite: true,
    actionText: "查看帮助",
    onAction: () => {
      window.open(
        "https://github.com/mzz2017/V2RayA#%E4%BD%BF%E7%94%A8",
        "_blank"
      );
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
      let msg = `无法通信。如果您的服务端已正常运行，且端口正常开放，原因可能是当前浏览器不允许https站点访问http资源，您可以尝试切换为http备用站点。`;
      let host = parseURL(err.config.url).host;
      if (host === "localhost" || host === "local" || host === "127.0.0.1") {
        if (browser.versions.webKit) {
          //Chrome等webkit内核浏览器允许访问http://localhost，只有可能是服务端未启动
          informNotRunning();
          return;
        }
        if (browser.versions.gecko) {
          msg = `无法通信。即使您的服务端正常运行，火狐浏览器也不允许https站点访问http资源，包括${host}，您可以换用Chrome浏览器或切换为http备用站点。`;
        } else {
          msg = `无法通信。如果您的服务端已正常运行，原因可能是当前浏览器不允许https站点访问http资源，包括${host}，您可以换用Chrome浏览器或切换为http备用站点。`;
        }
      }
      SnackbarProgrammatic.open({
        message: msg,
        type: "is-warning",
        position: "is-top",
        queue: false,
        duration: 10000,
        actionText: "切换为备用站点",
        onAction: () => {
          window.open("http://v.mzz.pub", "_self");
          // ToastProgrammatic.open({
          //   message:
          //     "暂无备用站点，如果您有意提供自动部署的HTTP站点，可以邮件至m@mzz.pub或直接发起pull request",
          //   type: "is-warning",
          //   position: "is-top",
          //   queue: false,
          //   duration: 10000
          // });
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
