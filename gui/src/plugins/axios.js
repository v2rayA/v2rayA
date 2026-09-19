"use strict";

import axios from "axios";
import { escapeHtml, parseURL } from "@/assets/js/utils";
import browser from "@/assets/js/browser";
import i18n from "../plugins/i18n";
import { nanoid } from "nanoid";
import { registerProgrammatic, buefy } from "@/plugins/session";

// Programmatic instances mount in their own app and outlive the root
// tree; register the handles so session.restart() closes them.
function openSnackbar(params) {
  const handle = buefy().snackbar.open(params);
  registerProgrammatic(handle);
  return handle;
}

function openModalProgrammatic(params) {
  const handle = buefy().modal.open(params);
  registerProgrammatic(handle);
  return handle;
}

// modalCustomPorts is a Vue component. Importing it statically here creates
// axios -> modalCustomPorts -> session -> App -> modalCustomPorts, and App's
// top-level `components` registration reads the default export before
// modalCustomPorts finishes evaluating (TDZ), so the app fails to boot.
// Import it lazily at the moment the action runs.
async function openCustomPortsModal() {
  const { default: modalCustomPorts } = await import(
    "../components/modalCustomPorts"
  );
  openModalProgrammatic({
    component: modalCustomPorts,
    hasModalCard: true,
    customClass: "modal-custom-ports",
  });
}

axios.defaults.timeout = 60 * 1000; // timeout: 60秒
// The backend reads object query parameters (touch, whiches) as JSON text,
// which is how axios 0.21 serialized them; 0.28+ expands them into
// touch[id]=… instead. Keep the JSON form.
axios.defaults.paramsSerializer = (params) =>
  Object.entries(params)
    .filter(([, v]) => v !== undefined && v !== null)
    .map(([k, v]) => {
      const value = typeof v === "object" ? JSON.stringify(v) : String(v);
      return `${encodeURIComponent(k)}=${encodeURIComponent(value)}`;
    })
    .join("&");

axios.interceptors.request.use(
  (config) => {
    // Only the backend that issued the token gets it. The backend-address
    // dialog probes a user-typed URL with /api/version, which needs no auth.
    if (
      localStorage.getItem("token") !== null &&
      typeof config.url === "string" &&
      config.url.startsWith(apiRoot)
    ) {
      config.headers.Authorization = `${localStorage["token"]}`;
      config.headers["X-V2raya-Request-Id"] = nanoid();
    }
    return config;
  },
  (err) => {
    console.log("!", err.name, err.message);
    buefy().toast.open({
      message: err.message,
      type: "is-warning",
      position: "is-top",
      duration: 5000,
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
  openSnackbar({
    message: i18n.global.t("axios.messages.optimizeBackend"),
    type: "is-primary",
    duration: 10000,
    position: "is-top",
    actionText: i18n.global.t("operations.yes"),
    onAction: () => {
      // this.showCustomPorts = true;
      openCustomPortsModal();
    },
  });
  openSnackbar({
    message: i18n.global.t("axios.messages.noBackendFound", { url }),
    type: "is-warning",
    position: "is-top",
    duration: 10000,
    actionText: i18n.global.t("operations.helpManual"),
    onAction: () => {
      window.open(i18n.global.t("axios.urls.usage"), "_blank");
    },
  });
}

axios.interceptors.response.use(
  function (res) {
    // Buefy toasts and snackbars render their message with v-html, and
    // backend error strings embed user data such as node remarks.
    if (res.data && typeof res.data.message === "string") {
      res.data.message = escapeHtml(res.data.message);
    }
    return res;
  },
  function (err) {
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
      const reqUrl = (err.config && err.config.url) || "";
      const isAuthAction = reqUrl.includes("/api/login") || reqUrl.includes("/api/account");
      if (isAuthAction) {
        // Let login/register request callers handle their own UI state.
        return Promise.reject(err);
      }

      if (localStorage["token"]) {
        // Centralize auth recovery in App.vue's mounted() flow to avoid
        // programmatic modal stacking and overlay conflicts.
        localStorage.removeItem("token");
        window.location.reload();
      }
      return Promise.reject(err);
    } else if (
      u &&
      location.protocol.substr(0, 5) === "https" &&
      u.protocol === "http" &&
      // parseURL fabricates http:// for a relative apiRoot; only an absolute
      // http:// backend address is the mixed-content case
      /^http:\/\//i.test(err.config.url)
    ) {
      // https frontend communicating with http backend
      let msg = i18n.global.t("axios.messages.cannotCommunicate.0");
      if (host === "localhost" || host === "local" || host === "127.0.0.1") {
        if (browser.versions.webKit) {
          // Chrome and other WebKit browsers allow access to http://localhost, 
          // failures are likely due to backend service not being started.
          informNotRunning(u.source.replace(u.relative, ""));
          return Promise.reject(err);
        }
        if (browser.versions.gecko) {
          msg = i18n.global.t("axios.messages.cannotCommunicate.1");
        }
      }
      openSnackbar({
        message: msg,
        type: "is-warning",
        position: "is-top",
        duration: 10000,
        actionText: i18n.global.t("operations.switchSite"),
        onAction: () => {
          window.open("http://v.v2raya.org", "_self");
        },
      });
      openSnackbar({
        message: i18n.global.t("axios.messages.optimizeBackend"),
        type: "is-primary",
        duration: 10000,
        position: "is-top",
        actionText: i18n.global.t("operations.yes"),
        onAction: () => {
          // this.showCustomPorts = true;
          openCustomPortsModal();
        },
      });
    } else if (
      u &&
      ((err.message && err.message === "Network Error") ||
        (err.config && err.config.url === "/api/version"))
    ) {
      informNotRunning(u.source.replace(u.relative, ""));
    } else {
      // other errors
      if (
        !err.message ||
        (err.message && err.message.indexOf("404") >= 0) ||
        (err.response && err.response.status === 404)
      ) {
        // Interface doesn't exist, or expected error (e.g. cancellation), maybe legacy server version - ignore
        return Promise.reject(err);
      }
      console.log("!other");
      buefy().toast.open({
        message: err,
        type: "is-warning",
        position: "is-top",
        duration: 5000,
      });
    }
    return Promise.reject(err);
  }
);

export function install(app) {
  app.config.globalProperties.$axios = axios;
}

export default axios;
