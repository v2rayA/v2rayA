<template>
  <div class="modal-card" style="max-width: 450px;margin:auto">
    <header class="modal-card-head">
      <p class="modal-card-title">
        地址与端口
      </p>
    </header>
    <section class="modal-card-body">
      <b-field label="服务端地址" label-position="on-border">
        <b-input
          v-model="table.backendAddress"
          placeholder="http://localhost:2017"
          required
          pattern="https?://.+(:\d+)?"
        >
          >
        </b-input>
      </b-field>
      <template v-if="backendReady && dockerMode === false && !addressChanged">
        <b-field label="socks5端口" label-position="on-border">
          <b-input
            v-model="table.socks5"
            placeholder="20170"
            type="number"
            min="0"
            required
          ></b-input>
        </b-field>
        <b-field label="http端口" label-position="on-border">
          <b-input
            v-model="table.http"
            placeholder="20171"
            type="number"
            min="0"
            required
          ></b-input>
        </b-field>
        <b-field label="http端口(with PAC)" label-position="on-border">
          <b-input
            v-model="table.httpWithPac"
            placeholder="20172"
            type="number"
            min="0"
            required
          ></b-input>
        </b-field>
        <b-message
          v-show="dockerMode"
          type="is-warning"
          style="font-size:13px"
          class="after-line-dot5"
        >
          <p v-show="!dockerMode">
            如需修改后端运行地址(默认0.0.0.0:2017)，可在systemd中添加环境变量<code>V2RAYA_ADDRESS</code>或添加启动参数<code>--address</code>。
          </p>
          <p v-show="dockerMode">
            docker模式下如果未使用<code>--privileged --network host</code
            >参数启动容器，可通过修改端口映射修改socks5、http端口。
          </p>
          <p v-show="dockerMode">
            docker模式下不能正确判断端口占用，请确保输入的端口未被其他程序占用。
          </p>
        </b-message>
        <b-message
          type="is-info"
          style="font-size:13px"
          class="after-line-dot5"
        >
          如将端口设为0则表示关闭该端口
        </b-message>
      </template>
    </section>
    <footer class="modal-card-foot flex-end">
      <button class="button" @click="$emit('close')">
        取消
      </button>
      <button class="button is-primary" @click="handleClickSubmit">
        确定
      </button>
    </footer>
  </div>
</template>

<script>
import { handleResponse } from "../assets/js/utils";

export default {
  name: "ModalCustomPorts",
  data: () => ({
    table: {
      backendAddress: "http://localhost:2017",
      socks5: "20170",
      http: "20171",
      httpWithPac: "20172"
    },
    backendReady: false
  }),
  computed: {
    dockerMode() {
      return window.localStorage["docker"] === "true";
    },
    addressChanged() {
      let backendAddress = this.table.backendAddress;
      if (backendAddress.endsWith("/")) {
        backendAddress = backendAddress.substr(0, backendAddress.length - 1);
      }
      return backendAddress + "/api" !== apiRoot;
    }
  },
  created() {
    this.table.backendAddress = localStorage["backendAddress"];
    this.$axios({
      url: apiRoot + "/ports"
    }).then(res => {
      handleResponse(res, this, () => {
        console.log("!");
        this.backendReady = true;
        Object.assign(this.table, res.data.data);
      });
    });
  },
  methods: {
    handleClickSubmit() {
      //去除末位'/'
      let backendAddress = this.table.backendAddress;
      if (backendAddress.endsWith("/")) {
        backendAddress = backendAddress.substr(0, backendAddress.length - 1);
      }
      //当前服务端是否正常工作
      if (this.backendReady && !this.addressChanged) {
        this.$axios({
          url: backendAddress + "/api/ports",
          method: "put",
          data: {
            socks5: parseInt(this.table.socks5),
            http: parseInt(this.table.http),
            httpWithPac: parseInt(this.table.httpWithPac)
          }
        }).then(res => {
          handleResponse(res, this, () => {
            localStorage["backendAddress"] = backendAddress;
            this.$emit("close");
          });
        });
      } else {
        this.$axios({
          url: backendAddress + "/api/version"
        }).then(() => {
          localStorage["backendAddress"] = backendAddress;
          this.$emit("close");
          window.location.reload();
        });
      }
    }
  }
};
</script>

<style lang="scss">
.modal-custom-ports .modal-background {
  background-color: rgba(0, 0, 0, 0.6);
}
</style>
