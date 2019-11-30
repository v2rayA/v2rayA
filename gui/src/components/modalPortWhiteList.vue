<template>
  <div class="modal-card" style="max-width: 450px;margin:auto">
    <header class="modal-card-head">
      <p class="modal-card-title">
        出方向端口白名单
      </p>
    </header>
    <section class="modal-card-body">
      <b-message type="is-info" class="after-line-dot5">
        <p>
          全局透明代理会使得所有TCP、UDP流量走代理，由于V2Ray并非透明模式，走代理的流量其源IP地址会被替换为代理服务器的IP地址，从而导致被代理方（服务器）提供的对外服务无法正常工作。具体而言，某服务的客户端请求的服务器（被代理方）IP地址和得到回应的IP地址（代理服务器IP地址）不一致。
        </p>
        <p>
          因此，需要将主机提供的对外服务端口包含在白名单中，使其不走代理。如ssh(22)、v2raya({{
            v2rayaPort
          }})。
        </p>
        <p>
          如不对外提供服务则可不设置白名单，仅对局域网内主机提供服务也可不设置白名单。
        </p>
        <p>
          格式：22表示端口22，20170:20172表示20170到20172三个端口。
        </p>
      </b-message>
      <b-field label="TCP端口白名单">
        <b-taginput
          v-model="tcp"
          :before-adding="beforeAdding"
          icon=" iconfont icon-label"
        >
        </b-taginput> </b-field
      ><b-field label="UDP端口白名单">
        <b-taginput
          v-model="udp"
          :before-adding="beforeAdding"
          icon=" iconfont icon-label"
        >
        </b-taginput>
      </b-field>
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
import { handleResponse, parseURL } from "@/assets/js/utils";

export default {
  name: "ModalPortWhiteList",
  data: () => ({
    tcp: [],
    udp: []
  }),
  computed: {
    v2rayaPort() {
      let U = parseURL(apiRoot);
      let port = U.port;
      if (!port) {
        port =
          U.protocol === "http" ? "80" : U.protocol === "https" ? "443" : "";
      }
      return port;
    }
  },
  created() {
    this.$axios({
      url: apiRoot + "/portWhiteList"
    }).then(res => {
      handleResponse(res, this, () => {
        if (res.data.data.tcp && res.data.data.tcp.length > 0) {
          this.tcp = res.data.data.tcp;
        }
        if (res.data.data.udp && res.data.data.udp.length > 0) {
          this.udp = res.data.data.udp;
        }
      });
    });
  },
  methods: {
    handleClickSubmit() {
      this.$axios({
        url: apiRoot + "/portWhiteList",
        method: "put",
        data: {
          tcp: this.tcp,
          udp: this.udp
        }
      }).then(res => {
        handleResponse(res, this, () => {
          this.$emit("close");
        });
      });
    },
    beforeAdding(tag) {
      return /^\d+$/.test(tag) || /^\d+:\d+$/.test(tag);
    }
  }
};
</script>

<style lang="scss" scoped>
.after-line-dot5 {
  font-size: 14px;
  p {
    font-size: 14px;
  }
}
</style>
