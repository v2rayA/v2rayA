<template>
  <div id="app">
    <b-navbar ref="navs" fixed-top shadow type="is-light">
      <template slot="brand">
        <b-navbar-item href="/">
          <img
            src="@/assets/img/logo2.png"
            alt="v2rayA"
            class="logo no-select"
          />
        </b-navbar-item>
        <b-navbar-item tag="div">
          <b-tag
            id="statusTag"
            class="pointerTag"
            :type="statusMap[runningState.running]"
            @mouseenter.native="handleOnStatusMouseEnter"
            @mouseleave.native="handleOnStatusMouseLeave"
            @click.native="handleClickStatus"
            >{{ coverStatusText ? coverStatusText : runningState.running }}
          </b-tag>
        </b-navbar-item>
        <b-navbar-item tag="div">
          <b-dropdown
            v-if="updateOutboundDropdown"
            :triggers="isMobile ? ['click'] : ['click', 'hover']"
            aria-role="list"
            :close-on-click="false"
            @mouseenter.native="handleOutboundDropdownActiveChange"
            @active-change="handleOutboundDropdownActiveChange"
          >
            <template #trigger>
              <b-tag class="pointerTag" type="is-info" icon-right="menu-down"
                >{{ outboundName.toUpperCase() }}
              </b-tag>
            </template>

            <b-dropdown-item
              v-for="outbound in outbounds"
              :key="outbound"
              aria-role="listitem"
              class="is-flex padding-right-1rem justify-content-space-between outbound-dropdown"
              @mouseenter.native="handleOnOutboundMouseEnter(outbound)"
              @mouseleave.native="handleOnOutboundMouseLeave"
              @click="outboundName = outbound"
              ><p class="is-relative is-fullwidth">
                <span>{{ outboundNameDecorator(outbound) }}</span>
                <span>
                  <i
                    v-show="isMobile || outboundDropdownHover[outbound]"
                    class="iconfont icon-setting outbound-setting"
                    @click="handleClickOutboundSetting($event, outbound)"
                  ></i
                ></span></p
            ></b-dropdown-item>
            <b-dropdown-item
              aria-role="listitem"
              class="is-flex padding-right-1rem"
              separator
            ></b-dropdown-item>
            <b-dropdown-item
              aria-role="listitem"
              class="is-flex padding-right-1rem"
              @click="handleAddOutbound"
              >{{ $t("operations.addOutbound") }}
            </b-dropdown-item>
          </b-dropdown>
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
        <b-dropdown
          position="is-bottom-left"
          aria-role="menu"
          style="margin-right: 10px"
          class="menudropdown"
        >
          <a slot="trigger" class="navbar-item" role="button">
            <span class="no-select">{{ username }}</span>
            <i
              class="iconfont icon-caret-down"
              style="position: relative; top: 1px; left: 2px"
            ></i>
          </a>
          <b-dropdown-item
            custom
            aria-role="menuitem"
            v-html="$t('common.loggedAs', { username })"
          >
          </b-dropdown-item>
          <b-dropdown-item
            custom
            aria-role="menuitem"
            class="is-flex"
            style="
              box-sizing: content-box;
              height: 16px;
              width: 60px;
              justify-content: space-between;
            "
          >
            <img
              v-for="lang of langs"
              :key="lang.flag"
              :src="require(`@/assets/img/flags/flag_${lang.flag}.svg`)"
              :alt="lang.alt"
              style="height: 100%; flex-shrink: 0; cursor: pointer"
              @click="handleClickLang(lang.flag)"
            />
          </b-dropdown-item>
          <hr class="dropdown-divider" />
          <b-dropdown-item
            value="logout"
            aria-role="menuitem"
            class="no-select"
            @click="handleClickLogout"
          >
            <i
              class="iconfont icon-logout"
              style="position: relative; top: 1px"
            ></i>
            {{ $t("operations.logout") }}
          </b-dropdown-item>
        </b-dropdown>
      </template>
    </b-navbar>
    <node
      v-model="runningState"
      :outbound="outboundName"
      :observatory="observatory"
    />
    <b-modal
      :active.sync="showCustomPorts"
      has-modal-card
      trap-focus
      aria-role="dialog"
      aria-modal
      class="modal-custom-ports"
    >
      <ModalCustomAddress @close="showCustomPorts = false" />
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
import { parseURL } from "@/assets/js/utils";
import { waitingConnected } from "@/assets/js/networkInspect";
import axios from "@/plugins/axios";
import ModalLog from "@/components/modalLog";

export default {
  components: { ModalCustomAddress, node },
  data() {
    return {
      ws: null,
      observatory: null,
      showSidebar: true,
      statusMap: {
        [this.$t("common.checkRunning")]: "is-light",
        [this.$t("common.notRunning")]: "is-danger",
        [this.$t("common.isRunning")]: "is-success",
      },
      coverStatusText: "",
      runningState: {
        running: this.$t("common.checkRunning"),
        connectedServer: null,
        outboundToServerName: {},
      },
      showCustomPorts: false,
      langs: [
        { flag: "zh", alt: "简体中文" },
        { flag: "en", alt: "English" },
        { flag: "fa", alt: "فارسی" }
      ],
      outboundName: "proxy",
      outbounds: ["proxy"],
      outboundDropdownHover: {},
      updateOutboundDropdown: true,
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
  },
  mounted() {
    console.log("app created");
    let ba = localStorage.getItem("backendAddress");
    if (ba) {
      let u = parseURL(ba);
      document.title = `v2rayA - ${u.host}:${u.port}`;
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
        if (res.data.data.serviceValid === false) {
          this.$buefy.toast.open({
            message: this.$t("version.v2rayInvalid"),
            type: "is-danger",
            position: "is-top",
            queue: false,
            duration: 10000,
          });
        } else if (!res.data.data.v5) {
          this.$buefy.toast.open({
            message: this.$t("version.v2rayNotV5"),
            type: "is-danger",
            position: "is-top",
            queue: false,
            duration: 10000,
          });
        }
        localStorage["lite"] = res.data.data.lite;
        localStorage["loadBalanceValid"] = res.data.data.loadBalanceValid;
        localStorage["variant"] = res.data.data.variant;
      }
    });
    this.$axios({
      url: apiRoot + "/outbounds",
    }).then((res) => {
      if (res.data.code === "SUCCESS") {
        this.outbounds = res.data.data.outbounds;
      }
    });
    this.connectWsMessage();
  },
  beforeDestroy() {
    if (this.ws) {
      this.ws.close();
    }
  },
  methods: {
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
      url = `${protocol}://${u.host}:${
        u.port
      }/api/message?Authorization=${encodeURIComponent(localStorage["token"])}`;
      if (this.ws) {
        // console.log("ws close");
        this.ws.close();
      }
      const ws = new WebSocket(url);
      ws.onopen = () => {
        // console.log("ws opened");
        //
      };
      ws.onmessage = (msg) => {
        msg.data && that.handleMessage(JSON.parse(msg.data));
      };
      ws.onclose = () => {
        ws.onmessage = null;
        that.ws = null;
        // console.log("ws closed");
        setTimeout(() => {
          if (that.ws === null) {
            that.connectWsMessage();
          }
        }, 3000);
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
    },
    handleOutboundDropdownActiveChange(active) {
      if (active) {
        this.updateOutboundDropdown = false;
        this.updateOutboundDropdown = true;
      }
    },
    outboundNameDecorator(outbound) {
      if (this.runningState.outboundToServerName[outbound]) {
        if (
          typeof this.runningState.outboundToServerName[outbound] === "number"
        ) {
          return `${outbound} - ${this.$t("common.loadBalance")} (${
            this.runningState.outboundToServerName[outbound]
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
                this.outbounds = res.data.data.outbounds;
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
            this.outbounds = res.data.data.outbounds;
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
    handleClickLang(lang) {
      localStorage["_lang"] = lang;
      location.reload();
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
      if (this.runningState.running === this.$t("common.notRunning")) {
        let cancel;
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
  //    // 去掉讨厌的滚动条
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
    // 修复modal模糊问题
    left: 0 !important;
    right: 0 !important;
    margin: auto;
    transform: unset !important;
  }
}

.dropdown-item:focus {
  // 不要丑丑的outline
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
