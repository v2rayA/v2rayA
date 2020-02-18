<template>
  <div id="app">
    <b-navbar ref="navs" fixed-top shadow type="is-light">
      <template slot="brand">
        <b-navbar-item href="/">
          <img src="./assets/logo2.png" alt="V2RayA" class="logo no-select" />
        </b-navbar-item>
      </template>
      <template slot="start">
        <b-navbar-item tag="div">
          v2ray-core状态：
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
          设置
        </b-navbar-item>
        <b-navbar-item tag="a" @click.native="handleClickAbout">
          <i class="iconfont icon-heart" style="font-size: 1.25em"></i>
          关于
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

          <b-dropdown-item custom aria-role="menuitem">
            Logged as <b>{{ username }}</b>
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
            Logout
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
import CONST from "@/assets/js/const";
import ModalSetting from "@/components/modalSetting";
import node from "@/components/node";
import { Base64 } from "js-base64";
import ModalCustomAddress from "./components/modalCustomPorts";
import { parseURL } from "./assets/js/utils";

export default {
  components: { ModalCustomAddress, node },
  data() {
    return {
      statusMap: {
        [CONST.INSPECTING_RUNNING]: "is-light",
        [CONST.NOT_RUNNING]: "is-danger",
        [CONST.IS_RUNNING]: "is-success"
      },
      coverStatusText: "",
      runningState: {
        running: CONST.INSPECTING_RUNNING,
        connectedServer: null,
        lastConnectedServer: null
      },
      showCustomPorts: false
    };
  },
  computed: {
    username() {
      let token = localStorage["token"];
      if (!token) {
        return "未登录";
      }
      let payload = JSON.parse(Base64.decode(token.split(".")[1]));
      return payload["uname"];
    }
  },
  created() {
    let ba = localStorage.getItem("backendAddress");
    if (!ba) {
      ba = "http://localhost:2017";
      localStorage.setItem("backendAddress", ba);
    }
    let u = parseURL(ba);
    document.title = `V2RayA - ${u.host}:${u.port}`;
    this.$axios({
      url: apiRoot + "/version"
    }).then(res => {
      if (res.data.code === "SUCCESS") {
        let toastConf = {
          message: `V2RayA服务端正在运行${
            res.data.data.dockerMode ? "于Docker环境中" : ""
          }，Version: ${res.data.data.version}`,
          type: "is-dark",
          position: "is-top",
          duration: 3000
        };
        if (res.data.data.foundNew) {
          toastConf.duration = 5000;
          toastConf.message += `，检测到新版本: ${res.data.data.remoteVersion}`;
          toastConf.type = "is-success";
        }
        this.$buefy.toast.open(toastConf);
        localStorage["docker"] = res.data.data.dockerMode;
        localStorage["version"] = res.data.data.version;
        if (res.data.data.serviceValid === false) {
          this.$buefy.toast.open({
            message: "检测到v2ray-core可能未正确安装，请检查",
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
    handleOnStatusMouseEnter() {
      if (this.runningState.running === CONST.IS_RUNNING) {
        this.coverStatusText = "　关闭　";
      } else if (this.runningState.running === CONST.NOT_RUNNING) {
        this.coverStatusText = "　启动　";
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
                        <p class="modal-card-title">mzz2017 / V2RayA</p>
                    </header>
                    <section class="modal-card-body lazy">
                        <p>V2RayA是V2Ray的一个Web客户端，前端使用Vue.js构建，后端使用Golang构建。</p>
                        <p style="font-size:0.85em;text-indent:1em;color:rgba(0,0,0,0.6)">默认端口：</p>
                        <p style="font-size:0.85em;text-indent:1em;color:rgba(0,0,0,0.6)">2017: V2RayA后端端口</p>
                        <p style="font-size:0.85em;text-indent:1em;color:rgba(0,0,0,0.6)">20170: SOCKS协议</p>
                        <p style="font-size:0.85em;text-indent:1em;color:rgba(0,0,0,0.6)">20171: HTTP协议</p>
                        <p style="font-size:0.85em;text-indent:1em;color:rgba(0,0,0,0.6)">20172: 带PAC的HTTP协议</p>
                        <p style="font-size:0.85em;text-indent:1em;color:rgba(0,0,0,0.6)">其他端口：</p>
                        <p style="font-size:0.85em;text-indent:1em;color:rgba(0,0,0,0.6)">12345: tproxy </p>
                        <p style="font-size:0.85em;text-indent:1em;color:rgba(0,0,0,0.6)">12346: ssr relay</p>
                        <p>应用不会将任何用户数据保存在云端，所有用户数据存放在用户本地配置文件中。若服务端运行于docker，则当相应 docker volume 被清除时配置也将随之消失，请做好备份。
                        <p>在使用中如果发现任何问题，欢迎<a href="https://github.com/mzz2017/V2RayA/issues">提出issue</a>。</p>
                    </section>
                    <footer class="modal-card-foot">
                        <a class="is-link" href="https://github.com/mzz2017/V2RayA" target="_blank">
                          <img class="leave-right" src="https://img.shields.io/github/stars/mzz2017/V2RayA.svg?style=social" alt="stars">
                          <img class="leave-right" src="https://img.shields.io/github/forks/mzz2017/V2RayA.svg?style=social" alt="forks">
                          <img class="leave-right" src="https://img.shields.io/github/watchers/mzz2017/V2RayA.svg?style=social" alt="watchers">
                        </a>
                    </footer>
                </div>
`
      });
    },
    handleClickStatus() {
      if (this.runningState.running === CONST.NOT_RUNNING) {
        this.$axios({
          url: apiRoot + "/v2ray",
          method: "post"
        }).then(res => {
          if (res.data.code === "SUCCESS") {
            Object.assign(this.runningState, {
              running: CONST.IS_RUNNING,
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
        });
      } else if (this.runningState.running === CONST.IS_RUNNING) {
        this.$axios({
          url: apiRoot + "/v2ray",
          method: "delete"
        }).then(res => {
          if (res.data.code === "SUCCESS") {
            Object.assign(this.runningState, {
              running: CONST.NOT_RUNNING,
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
      window.location.reload();
    }
  }
};
</script>

<style lang="scss">
//TODO: 缓冲css到本地
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
</style>
