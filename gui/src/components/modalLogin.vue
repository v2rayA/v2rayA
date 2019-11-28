<template>
  <div class="modal-card" style="max-width: 450px;margin:auto">
    <header class="modal-card-head">
      <p class="modal-card-title">
        {{ first ? "初来乍到，请先注册账号" : "登陆 - V2rayA" }}
      </p>
    </header>
    <section class="modal-card-body">
      <p style="text-align: center">
        <img src="../assets/logo2.png" alt="V2RayA" />
      </p>
      <b-field label="Username" type="is-success">
        <b-input v-model="username" @keyup.enter.native="handleEnter" />
      </b-field>
      <b-field label="Password" type="is-success">
        <b-input
          v-model="password"
          type="password"
          maxlength="32"
          @keyup.enter.native="handleEnter"
        />
      </b-field>
    </section>
    <footer class="modal-card-foot flex-end">
      <b-button
        :class="{ 'is-primary': !first, 'is-twitter': first }"
        @click="handleClickSubmit"
      >
        {{ first ? "注册" : "登陆" }}
      </b-button>
    </footer>
  </div>
</template>

<script>
import { handleResponse } from "../assets/js/utils";

export default {
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
            window.location.reload();
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
            window.location.reload();
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
