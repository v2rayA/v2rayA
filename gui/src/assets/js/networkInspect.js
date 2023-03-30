import axios from "../../plugins/axios";
import Vue from "vue";
import { handleResponse } from "./utils";

// 如果节点已连接，reload页面
function waitingConnected(promise, interval, cancel, timeout) {
  let timer = setInterval(() => {
    axios({
      url: apiRoot + "/touch",
      timeout: interval,
    })
      .then((res) => {
        handleResponse(
          res,
          null,
          () => {
            if (res.data.data.running && res.data.data.touch.connectedServer) {
              clearInterval(timer);
              cancel && cancel();
              Vue.prototype.$remount();
            }
          },
          () => {
            if (res.data.message !== "the last request is being processed") {
              this.$buefy.toast.open({
                message: res.data.message,
                type: "is-warning",
                position: "is-top",
                queue: false,
                duration: 5000,
              });
            }
          }
        );
      })
      .catch((err) => {
        if (err.response.status === 401) {
          clearInterval(timer);
          cancel && cancel();
        }
      });
  }, interval);
  // weird cancelable promise, can not use Promise.race directly
  promise.then(() => {
    Promise.race([
      promise,
      new Promise((resolve) => {
        setTimeout(resolve, timeout ? timeout : 30 * 1000);
      }),
    ]).finally(() => {
      clearInterval(timer);
    });
  });
}

export { waitingConnected };
