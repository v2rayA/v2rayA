<template>
  <div id="app">
    <b-navbar ref="navs" fixed-top shadow type="is-light">
      <template slot="brand">
        <b-navbar-item href="/">
          <img src="./assets/logo2.png" alt="v2rayA" class="logo no-select" />
        </b-navbar-item>
      </template>
      <template slot="start">
        <b-navbar-item tag="div">
          {{ $t("common.v2rayCoreStatus") }}：
          <b-tag
            id="statusTag"
            :type="statusMap[runningState.running]"
            @mouseenter.native="handleOnStatusMouseEnter"
            @mouseleave.native="handleOnStatusMouseLeave"
            @click.native="handleClickStatus"
            >{{ coverStatusText ? coverStatusText : runningState.running }}
          </b-tag>
        </b-navbar-item>
      </template>

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
        <b-dropdown
          position="is-bottom-left"
          aria-role="menu"
          style="margin-right:10px"
          class="menudropdown"
        >
          <a slot="trigger" class="navbar-item" role="button">
            <span class="no-select">{{ username }}</span>
            <i
              class="iconfont icon-caret-down"
              style="position: relative; top: 1px; left:2px"
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
            style="box-sizing: content-box;height: 16px;width: 60px;justify-content: space-between;"
          >
            <img
              v-for="lang of langs"
              :key="lang.flag"
              :src="`/img/flags/flag_${lang.flag}.svg`"
              :alt="lang.alt"
              style="height:100%;flex-shrink: 0;cursor: pointer"
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
              style="position: relative;top:1px;"
            ></i>
            {{ $t("operations.logout") }}
          </b-dropdown-item>
        </b-dropdown>
      </template>
    </b-navbar>
    <node v-model="runningState" />
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
import ModalCustomAddress from "./components/modalCustomPorts";
import { parseURL } from "./assets/js/utils";
import { waitingConnected } from "./assets/js/networkInspect";
import axios from "./plugins/axios";

export default {
  components: { ModalCustomAddress, node },
  data() {
    return {
      statusMap: {
        [this.$t("common.checkRunning")]: "is-light",
        [this.$t("common.notRunning")]: "is-danger",
        [this.$t("common.isRunning")]: "is-success"
      },
      coverStatusText: "",
      runningState: {
        running: this.$t("common.checkRunning"),
        connectedServer: null,
        lastConnectedServer: null
      },
      showCustomPorts: false,
      langs: [
        { flag: "zh", alt: "简体中文" },
        { flag: "en", alt: "English" }
      ]
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
    }
  },
  created() {
    console.log("app created");
    let ba = localStorage.getItem("backendAddress");
    let u = parseURL(ba);
    document.title = `v2rayA - ${u.host}:${u.port}`;
    this.$axios({
      url: apiRoot + "/version"
    }).then(res => {
      if (res.data.code === "SUCCESS") {
        let toastConf = {
          message: this.$t(
            res.data.data.dockerMode ? "welcome.docker" : "welcome.default",
            { version: res.data.data.version }
          ),
          type: "is-dark",
          position: "is-top",
          duration: 3000
        };
        if (res.data.data.foundNew) {
          toastConf.duration = 5000;
          toastConf.message +=
            ". " +
            this.$t("welcome.newVersion", {
              version: res.data.data.remoteVersion
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
            duration: 10000
          });
        } else {
          localStorage["iptablesMode"] = res.data.data.iptablesMode;
          localStorage["dohValid"] = res.data.data.dohValid;
        }
      }
    });
  },
  methods: {
    handleClickLang(lang) {
      localStorage["_lang"] = lang;
      location.reload();
    },
    handleOnStatusMouseEnter() {
      if (this.runningState.running === this.$t("common.isRunning")) {
        this.coverStatusText = "　" + this.$t("v2ray.stop") + "　";
      } else if (this.runningState.running === this.$t("common.notRunning")) {
        this.coverStatusText = "　" + this.$t("v2ray.start") + "　";
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
          }
        }
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
                        <a class="is-link" href="https://github.com/mzz2017/v2rayA" target="_blank">
                          <img class="leave-right" src="https://img.shields.io/github/stars/mzz2017/v2rayA.svg?style=social" alt="stars">
                          <img class="leave-right" src="https://img.shields.io/github/forks/mzz2017/v2rayA.svg?style=social" alt="forks">
                          <img class="leave-right" src="https://img.shields.io/github/watchers/mzz2017/v2rayA.svg?style=social" alt="watchers">
                        </a>
                    </footer>
                </div>
`
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
            })
          }).then(res => {
            if (res.data.code === "SUCCESS") {
              Object.assign(this.runningState, {
                running: this.$t("common.isRunning"),
                connectedServer: res.data.data.connectedServer,
                lastConnectedServer: null
              });
            } else {
              this.$buefy.toast.open({
                message: res.data.message,
                type: "is-warning",
                duration: 5000,
                position: "is-top"
              });
            }
          }),
          3 * 1000,
          cancel
        );
      } else if (this.runningState.running === this.$t("common.isRunning")) {
        this.$axios({
          url: apiRoot + "/v2ray",
          method: "delete"
        }).then(res => {
          if (res.data.code === "SUCCESS") {
            Object.assign(this.runningState, {
              running: this.$t("common.notRunning"),
              connectedServer: null,
              lastConnectedServer: res.data.data.lastConnectedServer
            });
            console.log(this.runningState);
          } else {
            this.$buefy.toast.open({
              message: res.data.message,
              type: "is-warning",
              duration: 5000,
              position: "is-top"
            });
          }
        });
      }
    },
    handleClickLogout() {
      localStorage.removeItem("token");
      this.$remount();
    }
  }
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

#statusTag:hover {
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
</style>
