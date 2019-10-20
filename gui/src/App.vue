<template>
  <div id="app">
    <b-navbar fixed-top shadow type="is-light" ref="navs">
      <template slot="brand">
        <b-navbar-item href="/">
          <img src="./assets/logo.png" alt="V2RayA" class="logo no-select" />
        </b-navbar-item>
      </template>
      <template slot="start">
        <b-navbar-item
          tag="router-link"
          to="/status"
          :active="nav === 'status'"
        >
          <i class="iconfont icon-earth" style="font-size: 1.2em"></i>
          状态 </b-navbar-item
        ><b-navbar-item tag="router-link" to="/node" :active="nav === 'node'">
          <i class="iconfont icon-cloud-server" style="font-size: 1.4em"></i>
          节点
        </b-navbar-item>
        <b-navbar-item
          tag="router-link"
          to="/subscription"
          :active="nav === 'subscription'"
        >
          <i class="iconfont icon-cloud-sync" style="font-size: 1.4em"></i>
          订阅
        </b-navbar-item>
        <b-navbar-item
          tag="router-link"
          to="/setting"
          :active="nav === 'settings'"
        >
          <i class="iconfont icon-setting" style="font-size: 1.25em"></i>
          设置
        </b-navbar-item>
        <b-navbar-item tag="router-link" to="/about" :active="nav === 'about'">
          <i class="iconfont icon-heart" style="font-size: 1.25em"></i>
          关于
        </b-navbar-item>
      </template>

      <template slot="end">
        <b-dropdown
          position="is-bottom-left"
          aria-role="menu"
          style="margin-right:10px"
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
    <router-view class="container" />
  </div>
</template>

<script>
import { mapState } from "vuex";
export default {
  data() {
    return {};
  },
  computed: mapState(["nav"]),
  watch: {}
};
</script>

<style lang="scss" scoped>
@import "https://at.alicdn.com/t/font_1467288_00gqxequ9vdv7.css";
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
</style>

<style lang="scss">
html {
  overflow-scrolling: touch;
  -webkit-overflow-scrolling: touch;
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
</style>
