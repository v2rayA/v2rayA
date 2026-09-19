import axios from "../../plugins/axios";
import { ToastProgrammatic } from "buefy";
import i18n from "@/plugins/i18n";
import { backendMessage, handleResponse } from "./utils";

const i18nVm = {
  $te: i18n.global.te.bind(i18n.global),
  $t: i18n.global.t.bind(i18n.global),
};

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
              // Do NOT call $remount() here — it recreates the whole App and can
              // trigger spurious "create account" modals. The connection state is
              // already updated in connectToProxyGroup's .then() handler.
            }
          },
          () => {
            if (
              res.data.errorCode !== "REQUEST_IN_PROGRESS" &&
              res.data.message !== "the last request is being processed"
            ) {
              ToastProgrammatic.open({
                message: i18n.global.t("connection.checkFailed", {
                  message: backendMessage(i18nVm, res) || i18n.global.t("common.fail"),
                }),
                type: "is-warning",
                position: "is-top",
                duration: 5000,
              });
            }
          }
        );
      })
      .catch((err) => {
        if (err && err.response && err.response.status === 401) {
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
