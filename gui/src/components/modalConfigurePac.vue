<template>
  <div
    class="modal-card modal-configure-pac"
    style="max-width: 400px;margin:auto"
  >
    <header class="modal-card-head">
      <p class="modal-card-title">自定义PAC路由规则</p>
    </header>
    <section class="modal-card-body rules">
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
              @click="handleClickRule(...arguments, index)"
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
          <b-field label="Tags" label-position="on-border">
            <b-taginput
              v-model="rule.tags"
              ellipsis
              icon=" iconfont icon-label"
              placeholder="Add a tag"
            >
            </b-taginput>
          </b-field>
          <b-field label="匹配类型" label-position="on-border">
            <b-select v-model="rule.matchType" expanded>
              <option value="domain">上述标签属于域名匹配</option>
              <option value="ip">上述标签属于IP匹配</option>
            </b-select>
          </b-field>
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
export default {
  name: "ModalConfigurePac",
  props: {
    customPac: {
      type: Object,
      default() {
        return {
          url: "",
          defaultProxyMode: "direct",
          routingRules: []
        };
      }
    }
  },
  data: () => ({
    isOpen: 0
  }),
  methods: {
    handleNew() {
      this.customPac.routingRules.push({
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
    handleClickRule(event, index) {
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
      this.$emit("submit", this.customPac);
      this.$parent.close();
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
</style>
