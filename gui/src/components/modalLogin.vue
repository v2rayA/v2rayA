<template>
  <div class="modal-card" style="max-width: 450px; margin: auto">
    <header class="modal-card-head">
      <p class="modal-card-title">
        {{ first ? $t("register.title") : `${$t("login.title")} - v2rayA` }}
      </p>
    </header>
    <section class="modal-card-body">
      <p style="text-align: center">
        <img src="@/assets/img/logo2.png" alt="v2rayA" />
      </p>
      <b-field :label="$t('login.username')" type="is-success">
        <b-input
          ref="username"
          v-model="username"
          @keyup.enter.native="handleEnter"
        ></b-input>
      </b-field>
      <b-field :label="$t('login.password')" type="is-success">
        <b-input
          v-model="password"
          type="password"
          :maxlength="first ? '32' : ''"
          @keyup.enter.native="handleEnter"
        ></b-input>
      </b-field>
      <b-message v-if="first" type="is-info" class="after-line-dot5">
        <p>{{ $t("register.messages.0") }}</p>
        <p>{{ $t("register.messages.1") }}</p>
        <p>{{ $t("register.messages.2") }}</p>
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
import { handleResponse } from "@/assets/js/utils";
import i18n from "@/plugins/i18n";

export default {
  i18n,
  name: "ModalLogin",
  props: {
    first: {
      type: Boolean,
      default: false,
    },
  },
  data: () => ({
    username: "",
    password: "",
  }),
  mounted() {
    this.$refs.username.focus();
  },
  methods: {
    handleClickSubmit() {
      if (this.first) {
        //register
        this.$axios({
          url: apiRoot + "/account",
          method: "post",
          data: {
            username: this.username,
            password: this.password,
          },
        }).then((res) => {
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
            password: this.password,
          },
        }).then((res) => {
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
    },
  },
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
