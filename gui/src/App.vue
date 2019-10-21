<template>
  <div id="app">
    <b-navbar fixed-top shadow type="is-light" ref="navs">
      <template slot="brand">
        <b-navbar-item href="/">
          <img src="./assets/logo.png" alt="V2RayA" class="logo no-select" />
        </b-navbar-item>
      </template>
      <template slot="start">
        <b-navbar-item tag="router-link" to="/node" :active="nav === 'node'">
          <i class="iconfont icon-cloud" style="font-size: 1.4em"></i>
          节点
        </b-navbar-item>
      </template>

      <template slot="end">
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
          <a class="navbar-item" slot="trigger" role="button">
            <span class="no-select">mzz2017</span>
            <i
              class="iconfont icon-caret-down"
              style="position: relative; top: 1px; left:2px"
            ></i>
          </a>

          <b-dropdown-item custom aria-role="menuitem">
            Logged as <b>mzz2017</b>
          </b-dropdown-item>
          <hr class="dropdown-divider" />
          <b-dropdown-item aria-role="menuitem" custom>
            V2Ray状态：<b-tag
              :type="statusMap[V2RayStatus]"
              @mouseenter.native="handleOnStatusMouseEnter"
              @mouseleave.native="handleOnStatusMouseLeave"
              id="statusTag"
              >{{ coverStatusText ? coverStatusText : V2RayStatus }}</b-tag
            >
          </b-dropdown-item>
          <template v-if="V2RayStatus === '正在运行'">
            <b-dropdown-item aria-role="menuitem" custom>
              节点名称：{{ nodeInfo.name }}
            </b-dropdown-item>
            <b-dropdown-item aria-role="menuitem" custom>
              节点地址：{{ nodeInfo.address }}
            </b-dropdown-item>
            <b-dropdown-item aria-role="menuitem" custom>
              Ping时延：{{ pingLatency }}
            </b-dropdown-item>
            <b-dropdown-item aria-role="menuitem" custom>
              HTTP时延：{{ httpLatency }}
            </b-dropdown-item>
          </template>
          <hr class="dropdown-divider" />
          <b-dropdown-item
            value="logout"
            aria-role="menuitem"
            class="no-select"
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
    <router-view />
  </div>
</template>

<script>
import { mapState } from "vuex";
import ModalSetting from "@/components/ModalSetting";
export default {
  data() {
    return {
      statusMap: {
        检测中: "is-light",
        尚未运行: "is-warning",
        正在运行: "is-success"
      },
      V2RayStatus: "正在运行",
      coverStatusText: "",
      nodeInfo: {
        name: "神出鬼没之无敌小豆豆",
        address: "192.168.50.111:12345"
      },
      pingLatency: "...",
      httpLatency: "1237ms"
    };
  },
  computed: mapState(["nav"]),
  watch: {},
  methods: {
    handleOnStatusMouseEnter() {
      if (this.V2RayStatus === "正在运行") {
        this.coverStatusText = "　关闭　";
      } else if (this.V2RayStatus === "尚未运行") {
        this.coverStatusText = "　启动　";
      }
    },
    handleOnStatusMouseLeave() {
      this.coverStatusText = "";
    },
    handleClickSetting() {
      this.$buefy.modal.open({
        parent: this,
        component: ModalSetting,
        hasModalCard: true,
        canCancel: false
      });
    },
    handleClickAbout() {
      this.$buefy.modal.open(
        `
<div class="modal-card" style="max-width: 500px;margin:auto">
                    <header class="modal-card-head">
                        <p class="modal-card-title">mzz2017 / V2RayA</p>
                    </header>
                    <section class="modal-card-body lazy">
                        <p>V2RayA是V2Ray的一个Web GUI，前端使用Vue构建，后端使用Golang构建。</p>
                        <p>整个项目依赖于Docker，如果你想修改socks或http的端口号，请修改docker参数并重新启动容器。</p>
                        <p>应用不会将任何用户数据保存在云端，所有用户数据存放在docker容器中，当docker容器被清除时配置也将随之消失。</p>
                        <p>在使用中如果发现任何问题，欢迎提出issue。</p>
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
      );
    }
  }
};
</script>

<style lang="scss" scoped>
@import "https://at.alicdn.com/t/font_1467288_5oe9ao15lkp.css";
#app {
  margin: 0;
}
.menucontainer {
  padding: 20px;
}
.logo {
  min-height: 2.5rem;
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
  overflow-scrolling: touch;
  -webkit-overflow-scrolling: touch;
  &::-webkit-scrollbar {
    // 去掉讨厌的滚动条
    display: none;
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
  $twitter: #4099ff;
  color: $twitter;
}
</style>
