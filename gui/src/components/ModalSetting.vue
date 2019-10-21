<template>
  <div class="modal-card" style="max-width: 520px;margin:auto">
    <header class="modal-card-head">
      <p class="modal-card-title">设置</p>
    </header>
    <section class="modal-card-body">
      <b-field
        label="GFWList"
        horizontal
        custom-class="modal-setting-label"
        style="position: relative"
        >5435 条，最后更新版本：
        <a
          href="https://github.com/hq450/fancyss/blob/master/rules/gfwlist.conf"
          class="is-link"
          >2018-09-28</a
        ><b-button
          size="is-small"
          type="is-text"
          style="position: relative;top:-2px;text-decoration:none"
          >更新</b-button
        ></b-field
      >
      <b-field
        label="大陆白名单IP段"
        horizontal
        custom-class="modal-setting-label"
        >8245 行，最后更新版本：
        <a
          href="https://github.com/hq450/fancyss/blob/master/rules/chnroute.txt"
          class="is-link"
          >2018-09-28</a
        ><b-button
          size="is-small"
          type="is-text"
          style="position: relative;top:-2px;text-decoration:none"
          >更新</b-button
        ></b-field
      >
      <hr class="dropdown-divider" style="margin: 1.25rem 0 1.25rem" />
      <b-field label="PAC模式使用" label-position="on-border">
        <b-select expanded v-model="pacMode">
          <option value="gfwlist">GFWList</option>
          <option value="whitelist">大陆白名单</option>
          <option value="custom">自定义PAC</option>
        </b-select>
        <b-input
          v-if="pacMode === 'custom'"
          v-model="customPac"
          placeholder="AutoProxy Rule List URL"
        ></b-input>
      </b-field>
      <b-field label="每日定时检查更新" label-position="on-border">
        <b-select expanded v-model="regularUpdateMode">
          <option value="none">不进行</option>
          <option value="update">更新PAC模式对应文件</option>
        </b-select>
        <b-clockpicker
          v-if="regularUpdateMode !== 'none'"
          placeholder="选择每日更新时间点"
          icon=" iconfont icon-clock2"
          hour-format="24"
          position="is-bottom-left"
          class="modal-setting-clockpicker"
        >
        </b-clockpicker>
      </b-field>
      <b-field label="获取订阅时使用" label-position="on-border">
        <b-select expanded v-model="subscriptionMode">
          <option value="direct">直连模式</option>
          <option value="pac">PAC模式</option>
          <option value="proxy">代理模式</option>
        </b-select>
      </b-field>
    </section>
    <footer class="modal-card-foot flex-end">
      <button class="button" type="button" @click="$parent.close()">
        取消
      </button>
      <button class="button is-primary">确定</button>
    </footer>
  </div>
</template>

<script>
export default {
  name: "ModalSetting",
  data: () => ({
    subscriptionMode: "direct",
    regularUpdateMode: "none",
    pacMode: "gfwlist",
    showClockPicker: true
  }),
  methods: {}
};
</script>

<style lang="scss">
.flex-end {
  justify-content: flex-end !important;
}
.modal-setting-label {
  width: 7em;
  padding: 0 !important;
  text-align: left !important;
}
.modal-setting-clockpicker {
  .background {
    display: unset;
    background-color: rgba(10, 10, 10, 0.6) !important;
  }
  .dropdown-menu {
    position: fixed;
    top: 50% !important;
    left: 50% !important;
    transform: translate3d(-50%, -50%, 0);
    right: unset !important;
    z-index: 50 !important;
  }
}
</style>
