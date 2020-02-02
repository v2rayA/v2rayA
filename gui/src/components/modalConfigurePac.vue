<template>
  <div
    class="modal-card modal-configure-pac"
    style="max-width: 400px;margin:auto"
  >
    <header class="modal-card-head">
      <p class="modal-card-title">自定义路由规则</p>
    </header>
    <section class="modal-card-body rules">
      <b-message type="is-info" class="after-line-dot5">
        <p>
          将SiteDat文件放于
          <b>{{ V2RayLocationAsset }}</b> 目录下，V2rayA将自动进行识别
        </p>
        <p>
          制作SiteDat文件：<a href="https://github.com/ToutyRater/V2Ray-SiteDAT"
            >ToutyRater/V2Ray-SiteDAT</a
          >
        </p>
      </b-message>
      <b-message type="is-success" class="after-line-dot5">
        <p>在选择Tags时，可按Ctrl等多选键进行多选。</p>
      </b-message>
      <b-collapse class="card">
        <div
          slot="trigger"
          slot-scope="props"
          class="card-header"
          role="button"
        >
          <p class="card-header-title">
            默认路由规则
          </p>
          <a class="card-header-icon">
            <b-icon :icon="props.open ? 'menu-down' : 'menu-up'"> </b-icon>
          </a>
        </div>
        <div class="card-content">
          <b-field label="默认路由规则" label-position="on-border">
            <b-select v-model="customPac.defaultProxyMode" expanded>
              <option value="direct">直连</option>
              <option value="proxy">代理</option>
              <option value="block">拦截</option>
            </b-select>
          </b-field>
        </div>
      </b-collapse>
      <b-collapse
        v-for="(rule, index) of customPac.routingRules"
        :key="rule.value"
        class="card"
      >
        <div
          slot="trigger"
          slot-scope="props"
          class="card-header"
          role="button"
        >
          <p class="card-header-title" style="position: relative;">
            <span>{{ `规则${index + 1}` }}</span>
            <b-button
              type="is-text"
              size="is-small"
              style="position: absolute;right:0"
              @click="handleClickDeleteRule(...arguments, index)"
              >删除</b-button
            >
          </p>
          <a class="card-header-icon">
            <b-icon
              :icon="
                props.open
                  ? ' iconfont icon-caret-down'
                  : ' iconfont icon-caret-up'
              "
            >
            </b-icon>
          </a>
        </div>
        <div class="card-content">
          <b-field label="域名文件" label-position="on-border">
            <b-select v-model="rule.filename" expanded>
              <option
                v-for="file of siteDatFiles"
                :key="file.filename"
                :value="file.filename"
                >{{ file.filename }}</option
              >
            </b-select>
          </b-field>
          <b-field label="Tags" label-position="on-border">
            <b-select
              v-model="rule.tags"
              multiple
              :native-size="
                siteDatFiles[rule.filename].tags.length > 8
                  ? 8
                  : siteDatFiles[rule.filename].tags.length
              "
              size="is-small"
              expanded
            >
              <option
                v-for="tag of siteDatFiles[rule.filename].tags"
                :key="tag"
                :value="tag"
                >{{ tag }}</option
              >
            </b-select>
          </b-field>
          <p class="content" style="font-size:0.8em;margin-left:0.5em">
            tags: {{ rule.tags }}
          </p>
          <b-field label="规则类型" label-position="on-border">
            <b-select v-model="rule.ruleType" expanded>
              <option value="direct"
                >直连{{
                  customPac.defaultProxyMode === "direct"
                    ? "(与默认规则相同)"
                    : ""
                }}</option
              >
              <option value="proxy"
                >代理{{
                  customPac.defaultProxyMode === "proxy"
                    ? "(与默认规则相同)"
                    : ""
                }}</option
              >
              <option value="block"
                >拦截{{
                  customPac.defaultProxyMode === "block"
                    ? "(与默认规则相同)"
                    : ""
                }}</option
              >
            </b-select>
          </b-field>
        </div>
      </b-collapse>
    </section>
    <footer class="modal-card-foot">
      <div
        style="position:relative;display:flex;justify-content:flex-end;width:100%"
      >
        <button class="button btn-new" type="button" @click="handleNew">
          新建规则
        </button>
        <button class="button" type="button" @click="$parent.close()">
          取消
        </button>
        <button class="button is-primary" @click="handleClickSubmit">
          保存
        </button>
      </div>
    </footer>
  </div>
</template>

<script>
import { handleResponse } from "@/assets/js/utils";
export default {
  name: "ModalConfigurePac",
  data: () => ({
    customPac: {
      defaultProxyMode: "",
      routingRules: []
    },
    siteDatFiles: [],
    firstSiteDatFilename: "",
    V2RayLocationAsset: ""
  }),
  created() {
    (async () => {
      let closing = false;
      let promiseSiteDatFiles = this.$axios({
        url: apiRoot + "/siteDatFiles"
      }).then(res => {
        handleResponse(res, this, () => {
          if (
            res.data.data.siteDatFiles &&
            res.data.data.siteDatFiles.length > 0
          ) {
            //将数组转换为map
            this.siteDatFiles = {};
            res.data.data.siteDatFiles.forEach(x => {
              this.siteDatFiles[x.filename] = x;
              x.tags.sort(); //对tags进行排序，方便查找
            });
            this.firstSiteDatFilename = res.data.data.siteDatFiles[0].filename;
          } else {
            this.$buefy.toast.open({
              message: "未在V2RayLocationAsset中发现siteDat文件",
              type: "is-warning",
              position: "is-top",
              queue: false,
              duration: 5000
            });
            closing = true;
          }
        });
      });
      let customPac;
      let promiseConfigurePac = this.$axios({
        url: apiRoot + "/customPac"
      }).then(res => {
        handleResponse(res, this, () => {
          customPac = res.data.data.customPac;
          this.V2RayLocationAsset = res.data.data.V2RayLocationAsset;
        });
      });
      await Promise.all([promiseConfigurePac, promiseSiteDatFiles]).then(() => {
        this.customPac = customPac;
        if (closing) {
          this.$parent.close();
        }
      });
    })();
  },
  methods: {
    handleNew() {
      this.customPac.routingRules.push({
        filename: this.firstSiteDatFilename,
        tags: [],
        matchType: "domain",
        ruleType:
          this.customPac.defaultProxyMode === "direct" ? "proxy" : "direct"
      });
      let target = document.querySelector(".modal-configure-pac .rules");
      this.$nextTick(() => {
        target.scrollTop = target.scrollHeight - target.clientHeight;
      });
    },
    handleClickDeleteRule(event, index) {
      event.stopImmediatePropagation();
      delete this.customPac.routingRules.splice(index, 1);
    },
    handleClickSubmit() {
      if (this.customPac.routingRules.some(x => x.tags.length <= 0)) {
        this.$buefy.toast.open({
          message: "不能存在tags为空的规则，请检查",
          type: "is-warning",
          position: "is-top",
          duration: 3000
        });
        return;
      }
      this.$axios({
        url: apiRoot + "/customPac",
        method: "put",
        data: {
          customPac: this.customPac
        }
      }).then(res => {
        handleResponse(res, this, () => {
          this.$parent.close();
        });
      });
    }
  }
};
</script>

<style lang="scss" scoped>
.btn-new {
  position: absolute;
  left: 0;
  top: 0;
}
</style>
<style lang="scss">
.icon-label {
  font-size: 24px !important;
}
.after-line-dot5 {
  font-size: 14px;
  p {
    font-size: 14px;
  }
}
</style>
