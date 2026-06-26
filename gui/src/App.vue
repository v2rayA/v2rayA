<template>
  <div id="app">
    <b-navbar ref="navs" fixed-top shadow type="is-light">
      <template slot="brand">
        <b-navbar-item href="/">
          <img src="@/assets/img/logo2.png" alt="v2rayA" class="logo no-select" />
        </b-navbar-item>
        <b-navbar-item tag="div">
          <b-tag id="statusTag" class="pointerTag" role="button" tabindex="0" :type="statusMap[runningState.running]"
            @mouseenter.native="handleOnStatusMouseEnter" @mouseleave.native="handleOnStatusMouseLeave"
            @click.native="handleClickStatus">{{ coverStatusText ? coverStatusText : runningState.running }}
          </b-tag>
        </b-navbar-item>
        <b-navbar-item tag="div">
          <OutboundGroupPanel
            :outbounds="outbounds"
            :current-outbound="outboundName"
            :is-mobile="isMobile"
            @select="outboundName = $event"
            @add-outbound="handleAddOutbound"
            @changed="handleGroupChanged"
            @group-deleted="handleGroupDeleted"
          />
        </b-navbar-item>
      </template>
      <template slot="start"></template>

      <template slot="end">
        <!--        <b-navbar-item tag="router-link" to="/node" :active="nav === 'node'">-->
        <!--          <i class="iconfont icon-cloud" style="font-size: 1.4em"></i>-->
        <!--          节点-->
        <!--        </b-navbar-item>-->
        <b-navbar-item tag="a" @click.native="handleClickSetting">
          <i class="iconfont icon-setting" style="font-size: 1.25em"></i>
          {{ $t("common.setting") }}
        </b-navbar-item>
        <b-navbar-item tag="a" @click.native="handleClickAbout">
          <i class="iconfont icon-heart" style="font-size: 1.25em"></i>
          {{ $t("common.about") }}
        </b-navbar-item>
        <b-navbar-item tag="a" @click.native="handleClickLogs">
          <i class="iconfont icon-info" style="font-size: 1.25em"></i>
          {{ $t("common.log") }}
        </b-navbar-item>
        <b-navbar-item tag="a" @click.native="toggleTheme">
          <i
            :class="themePreference === 'auto' ? 'mdi mdi-theme-light-dark' : (isDarkTheme ? 'mdi mdi-weather-sunny' : 'mdi mdi-weather-night')"
            style="font-size: 1.25em"
          ></i>
          {{ themeSwitchLabel }}
        </b-navbar-item>
        <b-dropdown position="is-bottom-left" aria-role="menu" class="langdropdown">
          <a slot="trigger" class="navbar-item" role="button">
            <i class="iconfont icon-earth" style="font-size: 1.25em; margin-right: 4px"></i>
            <span class="no-select">{{ currentLangLabel }}</span>
            <i class="iconfont icon-caret-down" style="position: relative; top: 1px; left: 2px"></i>
          </a>
          <b-dropdown-item v-for="lang of langs" :key="lang.code" aria-role="menuitem" class="no-select"
            @click="handleClickLang(lang.code)">
            <span style="font-weight: 500; min-width: 120px; display: inline-block">{{ lang.label }}</span>
            <span style="margin-left: 8px; color: #7a7a7a">{{ lang.code }}</span>
          </b-dropdown-item>
        </b-dropdown>
        <b-dropdown position="is-bottom-left" aria-role="menu" style="margin-right: 10px" class="menudropdown">
          <a slot="trigger" class="navbar-item" role="button">
            <span class="no-select">{{ username }}</span>
            <i class="iconfont icon-caret-down" style="position: relative; top: 1px; left: 2px"></i>
          </a>
          <b-dropdown-item custom aria-role="menuitem" v-html="$t('common.loggedAs', { username })">
          </b-dropdown-item>
          <hr class="dropdown-divider" />
          <b-dropdown-item value="logout" aria-role="menuitem" class="no-select" @click="handleClickLogout">
            <i class="iconfont icon-logout" style="position: relative; top: 1px"></i>
            {{ $t("operations.logout") }}
          </b-dropdown-item>
        </b-dropdown>
      </template>
    </b-navbar>
    <node ref="nodeRef" v-model="runningState" :outbound="outboundName" :outbounds="outbounds" :observatory="observatory" />
    <b-modal :active.sync="showCustomPorts" has-modal-card trap-focus aria-role="dialog" aria-modal
      class="modal-custom-ports">
      <ModalCustomAddress @close="showCustomPorts = false" />
    </b-modal>
    <b-modal :active.sync="loginModalActive" has-modal-card trap-focus aria-role="dialog" aria-modal class="modal-login modal-login-app">
      <ModalLogin :first="loginModalFirst" @close="loginModalActive = false" />
    </b-modal>
    <div id="login"></div>
  </div>
</template>

<script>
import ModalSetting from "@/components/modalSetting";
import node from "@/node";
import { Base64 } from "js-base64";
import ModalCustomAddress from "@/components/modalCustomPorts";
import ModalOutboundSetting from "@/components/modalOutboundSetting";
import OutboundGroupPanel from "@/components/outboundGroupPanel";
import { parseURL } from "@/assets/js/utils";
import { waitingConnected } from "@/assets/js/networkInspect";
import axios from "@/plugins/axios";
import ModalLog from "@/components/modalLog";
import ModalLogin from "@/components/modalLogin";
import { ModalProgrammatic } from "buefy";

export default {
  components: { ModalCustomAddress, node, OutboundGroupPanel, ModalLogin },
  data() {
    return {
      ws: null,
      observatory: null,
      loginModalActive: false,
      loginModalFirst: false,
      showSidebar: true,
      statusMap: {
        [this.$t("common.checkRunning")]: "is-light",
        [this.$t("common.notRunning")]: "is-danger",
        [this.$t("common.isRunning")]: "is-success",
        [this.$t("common.waitingNetwork")]: "is-warning",
      },
      coverStatusText: "",
      runningState: {
        running: this.$t("common.checkRunning"),
        networkPaused: false,
        connectedServer: null,
        outboundToServerName: {},
      },
      showCustomPorts: false,
      langs: [
        { code: "zh_CN", label: "中文-中国", flag: "zh" },
        { code: "en_US", label: "English-US", flag: "en" },
        { code: "fa_IR", label: "فارسی", flag: "fa" },
        { code: "ru_RU", label: "Русский", flag: "ru" },
        { code: "pt_BR", label: "Português-Brasil", flag: "pt" },
      ],
      outboundName: "proxy",
      outbounds: ["proxy"],
      outboundDropdownHover: {},
      updateOutboundDropdown: true,
      themePreference: 'auto',
      systemDark: window.matchMedia('(prefers-color-scheme: dark)').matches,
    };
  },
  computed: {
    username() {
      let token = localStorage["token"];
      if (!token) {
        return this.$t("common.notLogin");
      }
      let payload = JSON.parse(Base64.decode(token.split(".")[1]));
      return payload["uname"];
    },
    isMobile() {
      return window.screen.width < 800;
    },
    currentLangLabel() {
      const currentLang = localStorage["_lang"] || "zh";
      const lang = this.langs.find(l => l.flag === currentLang);
      return lang ? lang.label : "中文-中国";
    },
    isDarkTheme() {
      if (this.themePreference === 'dark') return true;
      if (this.themePreference === 'light') return false;
      return this.systemDark;
    },
    themeSwitchLabel() {
      if (this.themePreference === 'auto') return this.$t('common.autoTheme');
      if (this.themePreference === 'dark') return this.$t('common.darkTheme');
      return this.$t('common.lightTheme');
    },
  },
  mounted() {
    console.log("app created");
    this.initTheme();
    let ba = localStorage.getItem("backendAddress");
    if (ba) {
      let u = parseURL(ba);
      document.title = `v2rayA - ${u.host}:${u.port}`;
    }
    // 没有 token：先检查是否需要注册，避免触发需要认证的请求导致 401 二次弹窗
    if (!localStorage["token"]) {
      // 使用重试机制确保注册页面可靠弹出
      const checkAccount = (retries = 3) => {
        this.$axios({
          url: apiRoot + "/account",
          method: "get",
        }).then((res) => {
          if (res.data.code === "SUCCESS") {
            const hasAnyAccounts = !!(res.data.data && res.data.data.hasAnyAccounts);
            // first=true -> register flow, first=false -> login flow
            this.showLoginModal(!hasAnyAccounts);
          }
        }).catch(() => {
          if (retries > 0) {
            // 网络错误时延迟重试，避免因后端尚未就绪而错过注册页面
            setTimeout(() => checkAccount(retries - 1), 2000);
          }
        });
      };
      checkAccount();
      // 无 token 时跳过需要认证的请求，注册/登录后会自动 remount
      return;
    }
    this.$axios({
      url: apiRoot + "/version",
    }).then((res) => {
      if (res.data.code === "SUCCESS") {
        let toastConf = {
          message: this.$t(
            res.data.data.dockerMode ? "welcome.docker" : "welcome.default",
            { version: res.data.data.version }
          ),
          type: "is-dark",
          position: "is-top",
          duration: 3000,
          queue: false,
        };
        if (res.data.data.foundNew) {
          toastConf.duration = 5000;
          toastConf.message +=
            ". " +
            this.$t("welcome.newVersion", {
              version: res.data.data.remoteVersion,
            });
          toastConf.type = "is-success";
        }
        this.$buefy.toast.open(toastConf);
        localStorage["docker"] = res.data.data.dockerMode;
        localStorage["version"] = res.data.data.version;
        if (res.data.data.coreVersionValid === false) {
          this.$buefy.toast.open({
            message: this.$t("version.coreVersionMismatch", { err: res.data.data.coreVersionErr || "" }),
            type: "is-danger",
            position: "is-top",
            queue: false,
            duration: 0,
          });
        } else if (res.data.data.serviceValid === false) {
          this.$buefy.toast.open({
            message: this.$t("version.v2rayInvalid"),
            type: "is-danger",
            position: "is-top",
            queue: false,
            duration: 10000,
          });
        }
        localStorage["lite"] = res.data.data.lite;
        localStorage["loadBalanceValid"] = res.data.data.loadBalanceValid;
        localStorage["variant"] = res.data.data.variant;
        localStorage["coreVersionValid"] = res.data.data.coreVersionValid;
        localStorage["coreVersionErr"] = res.data.data.coreVersionErr || "";
      }
    });
    this.$axios({
      url: apiRoot + "/outbounds",
    }).then((res) => {
      if (res.data.code === "SUCCESS") {
        this.outbounds = this.normalizeOutbounds(res.data.data.outbounds);
      }
    }).catch(() => {
      // 静默处理认证错误（如 token 过期），避免触发 axios 401 拦截器弹出登录框
    });
    this.connectWsMessage();
  },
  beforeDestroy() {
    if (this.ws) {
      this.ws.close();
    }
    if (this._darkMediaQuery && this._onSystemThemeChange) {
      this._darkMediaQuery.removeEventListener('change', this._onSystemThemeChange);
    }
  },
  methods: {
    showLoginModal(first) {
      this.loginModalFirst = first;
      this.loginModalActive = true;
    },
    connectWsMessage() {
      const that = this;
      let url = apiRoot;
      if (!url.trim() || url.startsWith("/")) {
        url = location.protocol + "//" + location.host + url;
      }
      let protocol = "ws";
      let u = parseURL(url);
      if (u.protocol === "https") {
        protocol = "wss";
      }
      url = `${protocol}://${u.host}:${u.port
        }/api/message?Authorization=${encodeURIComponent(localStorage["token"])}`;
      if (this.ws) {
        // console.log("ws close");
        this.ws.close();
      }
      const ws = new WebSocket(url);
      // WebSocket 重连指数退避参数
      if (typeof this._wsRetries === "undefined") {
        this._wsRetries = 0;
      }
      ws.onopen = () => {
        // console.log("ws opened");
        this._wsRetries = 0; // 连接成功后重置重试计数
        // Re-sync running state on every (re)connect.  WebSocket messages are
        // not replayed, so if the connection was broken while Tun was setting up
        // routes the frontend might miss the running_state message and stay
        // stuck on "检测中" (Checking) forever.
        this.$nextTick(() => {
          if (this.$refs.nodeRef) {
            this.$refs.nodeRef.syncLatestNodeOverview();
          }
        });
      };
      ws.onmessage = (msg) => {
        msg.data && that.handleMessage(JSON.parse(msg.data));
      };
      ws.onclose = () => {
        ws.onmessage = null;
        that.ws = null;
        // 指数退避重连：1s, 2s, 4s, 8s... 最大 30 秒
        const delay = Math.min(1000 * Math.pow(2, that._wsRetries), 30000);
        that._wsRetries++;
        setTimeout(() => {
          if (that.ws === null) {
            that.connectWsMessage();
          }
        }, delay);
      };
      this.ws = ws;
    },
    handleMessage(msg) {
      if (
        msg.type === "observatory" &&
        msg.body.outboundName === this.outboundName
      ) {
        this.observatory = msg;
      }
      if (msg.type === "running_state" && msg.body) {
        if (msg.body.running === false) {
          this.$refs.nodeRef && this.$refs.nodeRef.notifyStopped(!!msg.body.networkPaused);
        } else {
          this.$refs.nodeRef && this.$refs.nodeRef.notifyRunning(!!msg.body.networkPaused);
        }
      }
    },
    handleOutboundDropdownActiveChange(active) {
      if (active) {
        this.updateOutboundDropdown = false;
        this.updateOutboundDropdown = true;
      }
    },
    handleGroupChanged() {
      // Refresh node.vue's data after group membership change from the panel
      if (this.$refs.nodeRef && this.$refs.nodeRef.created) {
        this.$refs.nodeRef.$axios({ url: apiRoot + "/touch" }).then((res) => {
          if (res.data && res.data.code === "SUCCESS") {
            this.$refs.nodeRef.refreshTableData(res.data.data.touch, res.data.data.running);
            this.$refs.nodeRef.updateConnectView();
          }
        }).catch(() => {});
      }
    },
    handleGroupDeleted(newOutbounds) {
      this.outbounds = this.normalizeOutbounds(newOutbounds);
      // If deleted group was selected, fall back to proxy
      if (!this.outbounds.includes(this.outboundName)) {
        this.outboundName = "proxy";
      }
    },
    normalizeOutbounds(outbounds) {
      const seen = new Set();
      const normalized = [];
      if (outbounds instanceof Array) {
        for (const outbound of outbounds) {
          if (typeof outbound !== "string") {
            continue;
          }
          const name = outbound.trim();
          if (!name || seen.has(name)) {
            continue;
          }
          seen.add(name);
          normalized.push(name);
        }
      }
      if (!seen.has("proxy")) {
        normalized.unshift("proxy");
      }
      return normalized;
    },
    outboundNameDecorator(outbound) {
      if (this.runningState.outboundToServerName[outbound]) {
        if (
          typeof this.runningState.outboundToServerName[outbound] === "number"
        ) {
          return `${outbound} - ${this.$t("common.loadBalance")} (${this.runningState.outboundToServerName[outbound]
            })`;
        } else {
          return `${outbound} - ${this.runningState.outboundToServerName[outbound]}`;
        }
      }
      return outbound;
    },
    handleAddOutbound() {
      this.$buefy.dialog.prompt({
        message: this.$t("outbound.addMessage"),
        inputAttrs: {
          maxlength: 10,
        },
        trapFocus: true,
        onConfirm: (outbound) => {
          let cancel;
          waitingConnected(
            this.$axios({
              url: apiRoot + "/outbound",
              method: "post",
              data: {
                outbound,
              },
              cancelToken: new axios.CancelToken(function executor(c) {
                cancel = c;
              }),
            }).then((res) => {
              if (res.data.code === "SUCCESS") {
                this.$buefy.toast.open({
                  message: this.$t("common.success"),
                  type: "is-success",
                  duration: 2000,
                  position: "is-top",
                  queue: false,
                });
                this.outbounds = this.normalizeOutbounds(res.data.data.outbounds);
              } else {
                this.$buefy.toast.open({
                  message: res.data.message,
                  type: "is-warning",
                  duration: 5000,
                  position: "is-top",
                  queue: false,
                });
              }
            }),
            3 * 1000,
            cancel
          );
        },
      });
    },
    handleDeleteOutbound(outbound) {
      let cancel;
      waitingConnected(
        this.$axios({
          url: apiRoot + "/outbound",
          method: "delete",
          data: {
            outbound,
          },
          cancelToken: new axios.CancelToken(function executor(c) {
            cancel = c;
          }),
        }).then((res) => {
          if (res.data.code === "SUCCESS") {
            this.$buefy.toast.open({
              message: this.$t("common.success"),
              type: "is-success",
              duration: 2000,
              position: "is-top",
              queue: false,
            });
            this.outbounds = this.normalizeOutbounds(res.data.data.outbounds);
            if (this.outboundName === outbound) {
              this.outboundName = "proxy";
            }
            if (outbound in this.runningState.outboundToServerName) {
              this.runningState.connectedServer =
                this.runningState.connectedServer.filter(
                  (cs) => cs.outbound !== outbound
                );
            }
          } else {
            this.$buefy.toast.open({
              message: res.data.message,
              type: "is-warning",
              duration: 5000,
              position: "is-top",
              queue: false,
            });
          }
        }),
        3 * 1000,
        cancel
      );
    },
    handleClickOutboundSetting(event, outbound) {
      event.stopPropagation();
      const that = this;
      this.$buefy.modal.open({
        parent: this,
        component: ModalOutboundSetting,
        hasModalCard: true,
        canCancel: true,
        props: {
          outbound: outbound,
        },
        events: {
          delete() {
            that.handleDeleteOutbound(outbound);
          },
        },
      });
    },
    handleClickLang(langCode) {
      const lang = this.langs.find(l => l.code === langCode);
      if (lang) {
        localStorage["_lang"] = lang.flag;
        location.reload();
      }
    },
    handleOnOutboundMouseEnter(outbound) {
      this.outboundDropdownHover = { [outbound]: true };
    },
    handleOnOutboundMouseLeave() {
      this.outboundDropdownHover = {};
    },
    handleOnStatusMouseEnter() {
      if (this.runningState.running === this.$t("common.isRunning")) {
        this.coverStatusText = this.$t("v2ray.stop");
      } else if (this.runningState.running === this.$t("common.notRunning")) {
        this.coverStatusText = this.$t("v2ray.start");
      } else if (this.runningState.running === this.$t("common.waitingNetwork")) {
        this.coverStatusText = this.$t("common.waitingNetwork");
      }
    },
    handleOnStatusMouseLeave() {
      this.coverStatusText = "";
    },
    handleClickSetting() {
      const that = this;
      this.$buefy.modal.open({
        parent: this,
        component: ModalSetting,
        hasModalCard: true,
        canCancel: true,
        events: {
          clickPorts() {
            that.showCustomPorts = true;
          },
        },
      });
    },
    handleClickAbout() {
      this.$buefy.modal.open({
        width: 640,
        content: `
<div class="modal-card" style="margin:auto">
                    <header class="modal-card-head">
                        <p class="modal-card-title">mzz2017 / v2rayA</p>
                    </header>
                    <section class="modal-card-body lazy">
                        ${this.$t(`about`)}
                    </section>
                    <footer class="modal-card-foot">
                        <a class="is-link" href="https://github.com/v2rayA/v2rayA" target="_blank">
                          <img class="leave-right" src="https://img.shields.io/github/stars/mzz2017/v2rayA.svg?style=social" alt="stars">
                          <img class="leave-right" src="https://img.shields.io/github/forks/mzz2017/v2rayA.svg?style=social" alt="forks">
                          <img class="leave-right" src="https://img.shields.io/github/watchers/mzz2017/v2rayA.svg?style=social" alt="watchers">
                        </a>
                    </footer>
                </div>
`,
      });
    },
    handleClickStatus() {
      if (
        this.runningState.running === this.$t("common.notRunning") ||
        this.runningState.running === this.$t("common.waitingNetwork")
      ) {
        let cancel;
        let loading = this.$buefy.loading.open();
        waitingConnected(
          this.$axios({
            url: apiRoot + "/v2ray",
            method: "post",
            cancelToken: new axios.CancelToken(function executor(c) {
              cancel = c;
            }),
          }).then((res) => {
            if (res.data.code === "SUCCESS") {
              Object.assign(this.runningState, {
                running: this.$t("common.isRunning"),
                networkPaused: false,
                connectedServer: res.data.data.touch.connectedServer,
              });
            } else {
              this.$buefy.toast.open({
                message: res.data.message,
                type: "is-warning",
                duration: 5000,
                position: "is-top",
                queue: false,
              });
            }
          }).finally(() => {
            loading.close();
          }),
          3 * 1000,
          cancel
        );
      } else if (this.runningState.running === this.$t("common.isRunning")) {
        this.$axios({
          url: apiRoot + "/v2ray",
          method: "delete",
        }).then((res) => {
          if (res.data.code === "SUCCESS") {
            Object.assign(this.runningState, {
              running: this.$t("common.notRunning"),
              networkPaused: false,
              connectedServer: res.data.data.touch.connectedServer,
            });
          } else {
            this.$buefy.toast.open({
              message: res.data.message,
              type: "is-warning",
              duration: 5000,
              position: "is-top",
              queue: false,
            });
          }
        });
      }
    },
    handleClickLogout() {
      localStorage.removeItem("token");
      this.$remount();
    },
    initTheme() {
      const stored = localStorage.getItem('theme');
      this.themePreference = (stored === 'dark' || stored === 'light') ? stored : 'auto';
      this._darkMediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
      this._onSystemThemeChange = (e) => {
        this.systemDark = e.matches;
        this.applyThemeClass();
      };
      this._darkMediaQuery.addEventListener('change', this._onSystemThemeChange);
      this.applyThemeClass();
    },
    applyThemeClass() {
      document.body.classList.toggle('theme-dark', this.isDarkTheme);
    },
    toggleTheme() {
      const order = ['auto', 'light', 'dark'];
      const idx = order.indexOf(this.themePreference);
      this.themePreference = order[(idx + 1) % order.length];
      localStorage.setItem('theme', this.themePreference);
      this.applyThemeClass();
    },
    handleClickLogs() {
      this.$buefy.modal.open({
        parent: this,
        component: ModalLog,
        hasModalCard: true,
        canCancel: true,
      });
    },
  },
};
</script>

<style lang="scss">
@import "assets/iconfont/fonts/font.css";
@import "assets/scss/reset.scss";
@import "assets/scss/dark-theme.scss";
</style>

<style lang="scss" scoped>
#app {
  margin: 0;
}

.menucontainer {
  padding: 20px;
}

.logo {
  min-height: 2.5rem;
  margin-left: 1em;
  margin-right: 1em;
}

.navbar-item .iconfont {
  margin-right: 0.15em;
}

.pointerTag:hover {
  cursor: pointer;
}
</style>

<style lang="scss">
html {
  //  &::-webkit-scrollbar {
  //    // remove annoying scrollbar
  //    display: none;
  //  }

  #app {
    height: calc(100vh - 3.25rem);
    /*overflow-y: auto;*/
    //overflow-scrolling: touch;
    //-webkit-overflow-scrolling: touch;
  }
}

@media screen and (max-width: 1023px) {
  .dropdown.is-mobile-modal .dropdown-menu {
    // fix modal blur issues
    left: 0 !important;
    right: 0 !important;
    margin: auto;
    transform: unset !important;
  }
}

.dropdown-item:focus {
  // remove ugly outline
  outline: none !important;
}

.no-select {
  user-select: none;
  -webkit-user-drag: none;
}

.menudropdown .dropdown-item {
  font-size: 0.8em;
}

.leave-right {
  margin-right: 0.6rem;
}

.lazy {
  p {
    margin: 0.5em 0;
  }
}

a.navbar-item:focus,
a.navbar-item:focus-within,
a.navbar-item:hover,
a.navbar-item.is-active,
.navbar-link:focus,
.navbar-link:focus-within,
.navbar-link:hover,
.navbar-link.is-active,
.is-link,
a {
  $success: #506da4;
  color: $success;
}

.icon-loading_ico-copy {
  font-size: 2.5rem;
  color: rgba(0, 0, 0, 0.45);
  animation: loading-rotate 2s infinite linear;
}

.modal-custom-ports {
  z-index: 999;
}

/* Keep App-level login modal above any other modal overlays. */
.modal-login-app {
  z-index: 5000 !important;
}

.modal-login-app .modal-background {
  z-index: 0 !important;
}

.modal-login-app .modal-content,
.modal-login-app .animation-content,
.modal-login-app .modal-card {
  position: relative;
  z-index: 1 !important;
  pointer-events: auto !important;
}

.after-line-dot5 {
  p {
    margin-bottom: 0.5em;
  }
}

.about-small {
  font-size: 0.85em;
  text-indent: 1em;
  color: rgba(0, 0, 0, 0.6);
}

.margin-right-2em {
  margin-right: 2em;
}

.justify-content-space-between {
  justify-content: space-between;
}

.padding-right-1rem {
  padding-right: 2rem !important;
}

#statusTag {
  width: 5em;
}

.dropdown-menu .is-fullwidth {
  width: 100%;
}

.dropdown-menu .outbound-setting {
  position: absolute;
  right: -1.5rem;
  top: 0;
  font-size: 1rem;
}

.navbar-item .dropdown-menu .dropdown-content {
  max-height: calc(100vh - 60px);
  overflow-y: auto;
}
</style>
