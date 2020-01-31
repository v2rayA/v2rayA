<template>
  <div class="modal-card" style="max-width: 450px;margin:auto">
    <header class="modal-card-head">
      <p class="modal-card-title">
        DoH设置
      </p>
    </header>
    <section class="modal-card-body">
      <b-message type="is-info" class="after-line-dot5">
        <p>
          DoH即DNS over
          HTTPS，能够有效避免DNS污染，但一些DoH提供商的DoH服务可能被墙，请自行选择非代理条件下直连速度最快的DoH提供商
        </p>
        <p>
          大陆较好的DoH服务有geekdns: 233py.com、红鱼: rubyfish.cn等
        </p>
        <p>台湾有quad101: dns.twnic.tw等</p>
        <p>美国有cloudflare: 1.0.0.1等</p>
        <p>
          清单：<a href="https://dnscrypt.info/public-servers" target="_blank"
            >public-servers</a
          >
        </p>
        <p>
          另外，您可以在未受到DNS污染的国内服务器上自架DoH服务，以纵享丝滑。<a
            href="https://dnscrypt.info/implementations"
            target="_blank"
            >Server Implementations</a
          >
        </p>
      </b-message>
      <b-field label="DoH服务优先级列表">
        <b-input v-model="dohlist" type="textarea"
      /></b-field>
      <b-message type="is-danger" class="after-line-dot5">
        <p>
          建议上述列表2-3行即可，留空保存可恢复默认
        </p>
      </b-message>
    </section>
    <footer class="modal-card-foot flex-end">
      <button class="button" @click="$emit('close')">
        取消
      </button>
      <button class="button is-primary" @click="handleClickSubmit">
        保存
      </button>
    </footer>
  </div>
</template>

<script>
import { handleResponse, parseURL } from "@/assets/js/utils";

export default {
  name: "ModalDohSetting",
  data: () => ({
    dohlist: ""
  }),
  created() {
    this.$axios({
      url: apiRoot + "/dohList"
    }).then(res => {
      handleResponse(res, this, () => {
        if (res.data.data.dohlist) {
          let dohlist = res.data.data.dohlist;
          dohlist.trim();
          let arr = dohlist.split("\n");
          if (arr.length > 0) {
            this.dohlist = dohlist;
          }
        }
      });
    });
  },
  methods: {
    handleClickSubmit() {
      this.$axios({
        url: apiRoot + "/dohList",
        method: "put",
        data: {
          dohlist: this.dohlist
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
