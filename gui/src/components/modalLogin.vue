<template>
  <div class="modal-card" style="max-width: 450px;margin:auto">
    <header class="modal-card-head">
      <p class="modal-card-title">
        {{ first ? $t("register.title") : `${$t("login.title")} - V2rayA` }}
      </p>
    </header>
    <section class="modal-card-body">
      <p style="text-align: center">
        <img src="../assets/logo2.png" alt="V2RayA" />
      </p>
      <b-field :label="$t('login.username')" type="is-success">
        <b-input v-model="username" @keyup.enter.native="handleEnter"></b-input>
      </b-field>
      <b-field :label="$t('login.password')" type="is-success">
        <b-input
          v-model="password"
          type="password"
          maxlength="32"
          @keyup.enter.native="handleEnter"
        ></b-input>
      </b-field>
      <b-message v-if="first" type="is-info" class="after-line-dot5">
        <p>请记住您创建的管理员账号，用于登录该管理页面。</p>
        <p>账号信息位于本地，我们不会上传任何信息到服务器。</p>
        <p>如不慎忘记密码，可通过清除配置文件重置。</p>
      </b-message>
    </section>
    <footer class="modal-card-foot flex-end">
      <b-button
        :class="{ 'is-primary': !first, 'is-twitter': first }"
        @click="handleClickSubmit"
      >
        {{ first ? $t("operations.create") : $t("operations.login") }}
      </b-button>
    </footer>
  </div>
</template>

<script>
import { handleResponse } from "../assets/js/utils";
import i18n from "@/plugins/i18n";

export default {
  i18n,
  name: "ModalLogin",
  props: {
    first: {
      type: Boolean,
      default: false
    }
  },
  data: () => ({
    username: "",
    password: ""
  }),
  methods: {
    handleClickSubmit() {
      const that = this;
      if (this.first) {
        //register
        this.$axios({
          url: apiRoot + "/account",
          method: "post",
          data: {
            username: this.username,
            password: this.password
          }
        }).then(res => {
          handleResponse(res, this, () => {
            localStorage["token"] = res.data.data.token;
            this.$emit("close");
            this.$remount();
          });
        });
      } else {
        //login
        this.$axios({
          url: apiRoot + "/login",
          method: "post",
          data: {
            username: this.username,
            password: this.password
          }
        }).then(res => {
          handleResponse(res, this, () => {
            localStorage["token"] = res.data.data.token;
            this.$emit("close");
            this.$remount();
          });
        });
      }
    },
    handleEnter() {
      this.handleClickSubmit();
    }
  }
};
</script>

<style lang="scss">
.modal-login .modal-background {
  background-color: rgba(10, 10, 10, 0.7) !important;
}
</style>
<style lang="scss" scoped>
.after-line-dot5 {
  font-size: 14px;
  p {
    font-size: 14px;
  }
}
</style>
