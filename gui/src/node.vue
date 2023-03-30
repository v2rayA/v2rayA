<template>
  <section id="node-section" class="node-section container hero">
    <b-sidebar
      v-show="connectedServerInfo.length"
      :open="true"
      class="node-status-sidebar-reduced"
      :can-cancel="false"
      @mouseenter.native="showSidebar = true"
      @click.native="showSidebar = true"
    >
      <img src="@/assets/img/switch-menu.svg" width="36px" />
    </b-sidebar>
    <b-sidebar
      :open="showSidebar"
      type="is-light"
      :fullheight="false"
      :fullwidth="false"
      :overlay="false"
      :right="false"
      class="node-status-sidebar"
      :can-cancel="['outside']"
      @close="showSidebar = false"
      @mouseleave.native="showSidebar = false"
    >
      <b-message
        v-for="v of connectedServerInfo"
        :key="v.value"
        :title="`${v.info.name}${
          v.info.subscription_name ? ` [${v.info.subscription_name}]` : ''
        }`"
        :closable="false"
        size="is-small"
        :type="
          v.info.alive
            ? v.selected
              ? 'is-primary'
              : 'is-success'
            : v.info.alive === null
            ? 'is-light'
            : 'is-danger'
        "
        @click.native="handleClickConnectedServer(v.which)"
      >
        <div v-if="v.showContent">
          <p>{{ $t("server.protocol") }}: {{ v.info.net }}</p>
          <p v-if="v.info.delay && v.info.delay < 99999">
            {{ $t("server.latency") }}: {{ v.info.delay }}ms
          </p>
          <p v-if="!v.info.alive && v.info.last_seen_time">
            {{ $t("server.lastSeenTime") }}:
            {{ v.info.last_seen_time | unix2datetime }}
          </p>
          <p v-if="v.info.last_try_time">
            {{ $t("server.lastTryTime") }}:
            {{ v.info.last_try_time | unix2datetime }}
          </p>
        </div>
      </b-message>
    </b-sidebar>
    <div v-if="ready" class="hero-body">
      <b-field
        id="toolbar"
        grouped
        group-multiline
        :style="{
          background: overHeight
            ? isCheckedRowsPingable() || isCheckedRowsDeletable()
              ? 'rgba(0, 0, 0, 0.1)'
              : 'rgba(0, 0, 0, 0.05)'
            : 'transparent',
        }"
        :class="{ 'float-toolbar': overHeight }"
      >
        <div style="max-width: 60%">
          <button
            :class="{
              button: true,
              field: true,
              'is-info': true,
              'mobile-small': true,
              'not-display': !overHeight && !isCheckedRowsPingable(),
            }"
            :disabled="!isCheckedRowsPingable()"
            @click="handleClickLatency(true)"
          >
            <i class="iconfont icon-wave" />
            <span>PING</span>
          </button>
          <button
            :class="{
              button: true,
              field: true,
              'is-info': true,
              'mobile-small': true,
              'not-display': !overHeight && !isCheckedRowsPingable(),
            }"
            :disabled="!isCheckedRowsPingable()"
            @click="handleClickLatency(false)"
          >
            <i class="iconfont icon-wave" />
            <span>HTTP</span>
          </button>
          <button
            :class="{
              button: true,
              field: true,
              'is-delete': true,
              'mobile-small': true,
              'not-display': !overHeight && !isCheckedRowsDeletable(),
            }"
            :disabled="!isCheckedRowsDeletable()"
            @click="handleClickDelete"
          >
            <i class="iconfont icon-delete" />
            <span>{{ $t("operations.delete") }}</span>
          </button>
          <button
            :class="{
              button: true,
              field: true,
              'is-delete': true,
              'mobile-small': true,
              'not-show': true,
            }"
          >
            <i class="iconfont icon-delete" />
            <span>placeholder</span>
          </button>
          <span class="field not-show mobile-small">placeholder</span>
        </div>
        <div class="right">
          <b-button
            class="field mobile-small"
            type="is-primary"
            @click="handleClickCreate"
          >
            <i class="iconfont icon-chuangjiangongdan1" />
            <span>{{ $t("operations.create") }}</span>
          </b-button>
          <b-button
            class="field mobile-small"
            type="is-primary"
            @click="handleClickImport"
          >
            <i class="iconfont icon-daoruzupu-xianxing" />
            <span>{{ $t("operations.import") }}</span>
          </b-button>
        </div>
      </b-field>

      <b-collapse
        v-if="!tableData.subscriptions.length && !tableData.servers.length"
        class="card welcome-driver"
        aria-id="contentIdForA11y3"
      >
        <div
          slot="trigger"
          slot-scope="props"
          class="card-header"
          role="button"
          aria-controls="contentIdForA11y3"
        >
          <p class="card-header-title">
            {{ $t("welcome.title") }}
          </p>
          <a class="card-header-icon">
            <b-icon :icon="props.open ? 'menu-down' : 'menu-up'"></b-icon>
          </a>
        </div>
        <div class="card-content">
          <div class="content">
            <p>{{ $t("welcome.messages.0") }}</p>
            <p>{{ $t("welcome.messages.1") }}</p>
          </div>
        </div>
        <footer class="card-footer">
          <a class="card-footer-item" @click="handleClickCreate">{{
            $t("operations.create")
          }}</a>
          <a class="card-footer-item" @click="handleClickImport">{{
            $t("operations.import")
          }}</a>
        </footer>
      </b-collapse>
      <b-tabs
        v-if="tableData.subscriptions.length || tableData.servers.length"
        v-model="tab"
        position="is-centered"
        type="is-toggle-rounded"
        class="main-tabs"
        @input="handleTabsChange"
      >
        <b-tab-item label="SUBSCRIPTION">
          <b-field :label="`SUBSCRIPTION(${tableData.subscriptions.length})`">
            <b-table
              :data="tableData.subscriptions"
              :checked-rows.sync="checkedRows"
              :row-class="
                (row, index) =>
                  row.connected &&
                  runningState.running === $t('common.isRunning')
                    ? 'is-connected-running'
                    : row.connected
                    ? 'is-connected-not-running'
                    : null
              "
              default-sort="id"
              checkable
            >
              <b-table-column
                v-slot="props"
                field="id"
                label="ID"
                numeric
                sortable
              >
                {{ props.row.id }}
              </b-table-column>
              <b-table-column
                v-slot="props"
                field="host"
                :label="$t('subscription.host')"
                sortable
              >
                {{ props.row.host }}
              </b-table-column>
              <b-table-column
                v-slot="props"
                field="remarks"
                :label="$t('subscription.remarks')"
                sortable
              >
                {{ props.row.remarks }}
              </b-table-column>
              <b-table-column
                v-slot="props"
                field="status"
                :label="$t('subscription.timeLastUpdate')"
                width="260"
                sortable
              >
                {{ props.row.status }}
              </b-table-column>
              <b-table-column
                v-slot="props"
                :label="$t('subscription.numberServers')"
                centered
                numeric
                sortable
                :custom-sort="sortNumberServers"
              >
                {{ props.row.servers.length }}
              </b-table-column>
              <b-table-column
                v-slot="props"
                :label="$t('operations.name')"
                width="300"
              >
                <div class="operate-box">
                  <b-button
                    size="is-small"
                    icon-left=" github-circle iconfont icon-sync"
                    outlined
                    type="is-warning"
                    @click="handleClickUpdateSubscription(props.row)"
                  >
                    {{ $t("operations.update") }}
                  </b-button>
                  <b-button
                    size="is-small"
                    icon-left=" github-circle iconfont icon-wendangxiugai"
                    outlined
                    type="is-info"
                    @click="handleClickModifySubscription(props.row)"
                  >
                    {{ $t("operations.modify") }}
                  </b-button>
                  <b-button
                    size="is-small"
                    icon-left=" github-circle iconfont icon-share"
                    outlined
                    type="is-success"
                    @click="handleClickShare(props.row)"
                  >
                    {{ $t("operations.share") }}
                  </b-button>
                </div>
              </b-table-column>
            </b-table>
          </b-field>
        </b-tab-item>
        <b-tab-item
          label="SERVER"
          :icon="`${
            connectedServerInTab['server'] ? ' iconfont icon-dian' : ''
          }`"
        >
          <b-field :label="`SERVER(${tableData.servers.length})`">
            <b-table
              per-page="100"
              :current-page.sync="currentPage.servers"
              :data="tableData.servers"
              :checked-rows.sync="checkedRows"
              checkable
              :row-class="
                (row, index) =>
                  row.connected &&
                  runningState.running === $t('common.isRunning')
                    ? 'is-connected-running'
                    : row.connected
                    ? 'is-connected-not-running'
                    : null
              "
              default-sort="id"
            >
              <b-table-column
                v-slot="props"
                field="id"
                label="ID"
                numeric
                sortable
              >
                {{ props.row.id }}
              </b-table-column>
              <b-table-column
                v-slot="props"
                field="name"
                :label="$t('server.name')"
                sortable
              >
                {{ props.row.name }}
              </b-table-column>
              <b-table-column
                v-slot="props"
                field="address"
                :label="$t('server.address')"
                sortable
              >
                {{ props.row.address }}
              </b-table-column>
              <b-table-column
                v-slot="props"
                field="net"
                :label="$t('server.protocol')"
                sortable
              >
                {{ props.row.net }}
              </b-table-column>
              <b-table-column
                v-slot="props"
                field="pingLatency"
                :label="$t('server.latency')"
                class="ping-latency"
                sortable
                :custom-sort="sortping"
              >
                {{ props.row.pingLatency }}
              </b-table-column>
              <b-table-column
                v-slot="props"
                :label="$t('operations.name')"
                sortable
                :custom-sort="sortConnections"
                width="300"
              >
                <div class="operate-box">
                  <b-button
                    size="is-small"
                    :icon-left="` github-circle iconfont ${
                      props.row.connected
                        ? 'icon-Link_disconnect'
                        : 'icon-lianjie'
                    }`"
                    :outlined="!props.row.connected"
                    :type="props.row.connected ? 'is-warning' : 'is-warning'"
                    @click="handleClickAboutConnection(props.row)"
                  >
                    {{
                      loadBalanceValid
                        ? props.row.connected
                          ? $t("operations.cancel")
                          : $t("operations.select")
                        : props.row.connected
                        ? $t("operations.disconnect")
                        : $t("operations.connect")
                    }}
                  </b-button>
                  <b-button
                    size="is-small"
                    icon-left=" github-circle iconfont icon-wendangxiugai"
                    :outlined="!props.row.connected"
                    type="is-info"
                    @click="handleClickModifyServer(props.row)"
                  >
                    {{ $t("operations.modify") }}
                  </b-button>
                  <b-button
                    size="is-small"
                    icon-left=" github-circle iconfont icon-share"
                    :outlined="!props.row.connected"
                    type="is-success"
                    @click="handleClickShare(props.row)"
                  >
                    {{ $t("operations.share") }}
                  </b-button>
                </div>
              </b-table-column>
            </b-table>
          </b-field>
        </b-tab-item>
        <b-tab-item
          v-for="(sub, subi) of tableData.subscriptions"
          :key="sub.id"
          :label="
            (sub.remarks && sub.remarks.toUpperCase()) || sub.host.toUpperCase()
          "
          :icon="`${
            connectedServerInTab['subscriptionServer'][subi]
              ? ' iconfont icon-dian'
              : ''
          }`"
        >
          <b-field
            v-if="tab === subi + 2"
            :label="`${sub.host.toUpperCase()}(${sub.servers.length}${
              sub.info ? ') (' : ''
            }${sub.info})`"
          >
            <b-table
              :current-page.sync="currentPage[sub.id]"
              per-page="100"
              :data="sub.servers"
              :checked-rows.sync="checkedRows"
              checkable
              :row-class="
                (row, index) =>
                  row.connected &&
                  runningState.running === $t('common.isRunning')
                    ? 'is-connected-running'
                    : row.connected
                    ? 'is-connected-not-running'
                    : null
              "
              default-sort="id"
            >
              <b-table-column
                v-slot="props"
                field="id"
                label="ID"
                numeric
                sortable
              >
                {{ props.row.id }}
              </b-table-column>
              <b-table-column
                v-slot="props"
                field="name"
                :label="$t('server.name')"
                sortable
              >
                {{ props.row.name }}
              </b-table-column>
              <b-table-column
                v-slot="props"
                field="address"
                :label="$t('server.address')"
                sortable
              >
                {{ props.row.address }}
              </b-table-column>
              <b-table-column
                v-slot="props"
                field="net"
                :label="$t('server.protocol')"
                style="font-size: 0.9em"
                sortable
              >
                {{ props.row.net }}
              </b-table-column>
              <b-table-column
                v-slot="props"
                field="pingLatency"
                :label="$t('server.latency')"
                class="ping-latency"
                sortable
                :custom-sort="sortping"
              >
                {{ props.row.pingLatency }}
              </b-table-column>
              <b-table-column
                v-slot="props"
                :label="$t('operations.name')"
                sortable
                :custom-sort="sortConnections"
                width="300"
              >
                <div class="operate-box">
                  <b-button
                    size="is-small"
                    :icon-left="` github-circle iconfont ${
                      props.row.connected
                        ? 'icon-Link_disconnect'
                        : 'icon-lianjie'
                    }`"
                    :outlined="!props.row.connected"
                    :type="props.row.connected ? 'is-warning' : 'is-warning'"
                    @click="handleClickAboutConnection(props.row, subi)"
                  >
                    {{
                      loadBalanceValid
                        ? props.row.connected
                          ? $t("operations.cancel")
                          : $t("operations.select")
                        : props.row.connected
                        ? $t("operations.disconnect")
                        : $t("operations.connect")
                    }}
                  </b-button>
                  <b-button
                    size="is-small"
                    icon-left=" github-circle iconfont icon-winfo-icon-chakanbaogao"
                    :outlined="!props.row.connected"
                    type="is-info"
                    @click="handleClickViewServer(props.row, subi)"
                  >
                    {{ $t("operations.view") }}
                  </b-button>
                  <b-button
                    size="is-small"
                    icon-left=" github-circle iconfont icon-share"
                    :outlined="!props.row.connected"
                    type="is-success"
                    @click="handleClickShare(props.row, subi)"
                  >
                    {{ $t("operations.share") }}
                  </b-button>
                </div>
              </b-table-column>
            </b-table>
          </b-field>
        </b-tab-item>
      </b-tabs>
    </div>
    <b-loading v-else :is-full-page="true" :active="true">
      <i class="iconfont icon-loading_ico-copy" />
    </b-loading>
    <b-modal
      :active.sync="showModalServer"
      has-modal-card
      trap-focus
      aria-role="dialog"
      aria-modal
    >
      <ModalServer
        :which="which"
        :readonly="modalServerReadOnly"
        @submit="handleModalServerSubmit"
      />
    </b-modal>
    <b-modal
      :active.sync="showModalSubscription"
      has-modal-card
      trap-focus
      aria-role="dialog"
      aria-modal
    >
      <ModalSubscription
        :which="which"
        @submit="handleModalSubscriptionSubmit"
      />
    </b-modal>
    <input
      id="QRCodeImport"
      type="file"
      style="display: none"
      accept="image/*"
    />
    <b-modal
      :active.sync="showModalImport"
      has-modal-card
      trap-focus
      aria-role="dialog"
      aria-modal
      @after-enter="handleModalImportShow"
    >
      <div class="modal-card" style="width: 350px">
        <header class="modal-card-head">
          <p class="modal-card-title">{{ $t("operations.import") }}</p>
        </header>
        <section class="modal-card-body">
          {{ $t("import.message") }}
          <b-input
            ref="importInput"
            v-model="importWhat"
            icon-right=" iconfont icon-camera"
            icon-right-clickable
            @icon-right-click="handleClickImportQRCode"
            @keyup.native="handleImportEnter"
          ></b-input>
        </section>
        <footer class="modal-card-foot">
          <button
            class="button is-link is-light"
            type="button"
            @click="handleClickImportInBatch"
          >
            {{ $t("operations.inBatch") }}
          </button>
          <div
            style="
              display: flex;
              justify-content: flex-end;
              width: -moz-available;
            "
          >
            <button
              class="button"
              type="button"
              @click="showModalImport = false"
            >
              {{ $t("operations.cancel") }}
            </button>
            <button
              class="button is-primary"
              type="button"
              @click="handleClickImportConfirm"
            >
              {{ $t("operations.confirm") }}
            </button>
          </div>
        </footer>
      </div>
    </b-modal>
    <b-modal
      :active.sync="showModalImportInBatch"
      has-modal-card
      trap-focus
      aria-role="dialog"
      aria-modal
      @close="showModalImport = false"
    >
      <div class="modal-card" style="width: 350px">
        <header class="modal-card-head">
          <p class="modal-card-title">{{ $t("operations.import") }}</p>
        </header>
        <section class="modal-card-body">
          {{ $t("import.batchMessage") }}
          <b-input
            ref="importInput"
            v-model="importWhat"
            type="textarea"
            custom-class="horizon-scroll"
          ></b-input>
        </section>
        <footer class="modal-card-foot">
          <div
            style="
              display: flex;
              justify-content: flex-end;
              width: -moz-available;
            "
          >
            <button
              class="button"
              type="button"
              @click="
                () => {
                  showModalImport = false;
                  showModalImportInBatch = false;
                }
              "
            >
              {{ $t("operations.cancel") }}
            </button>
            <button
              class="button is-primary"
              type="button"
              @click="handleClickImportConfirm"
            >
              {{ $t("operations.confirm") }}
            </button>
          </div>
        </footer>
      </div>
    </b-modal>
  </section>
</template>

<script>
import { locateServer, handleResponse } from "@/assets/js/utils";
import CONST from "@/assets/js/const";
import QRCode from "qrcode";
import { Decoder } from "@nuintun/qrcode";
import ClipboardJS from "clipboard";
import { Base64 } from "js-base64";
import ModalServer from "@/components/modalServer";
import ModalSubscription from "@/components/modalSubcription";
import ModalSharing from "@/components/modalSharing";
import { waitingConnected } from "@/assets/js/networkInspect";
import axios from "@/plugins/axios";
import * as dayjs from "dayjs";

export default {
  name: "Node",
  components: { ModalSubscription, ModalServer },
  filters: {
    unix2datetime(x) {
      x = dayjs.unix(x);
      let now = dayjs();
      if (localStorage["_lang"] === "zh") {
        now = now.locale("zh-cn");
      } else if (localStorage["_lang"] === "en") {
        now = now.locale("en");
      }
      return now.to(x);
    },
  },
  props: {
    outbound: {
      type: String,
      default: "proxy",
    },
    observatory: {
      type: Object,
      default() {
        return null;
      },
    },
  },
  data() {
    return {
      enterReducedSidebar: false,
      showSidebar: false,
      importWhat: "",
      showModalImport: false,
      showModalImportInBatch: false,
      currentPage: { servers: 1, subscriptions: 1 },
      tableData: {
        servers: [],
        subscriptions: [],
        connectedServer: [],
      },
      checkedRows: [],
      ready: false,
      tab: 0,
      runningState: {
        running: this.$t("common.checkRunning"),
        connectedServer: null,
        outboundToServerName: {},
      },
      showModalServer: false,
      which: null,
      modalServerReadOnly: false,
      showModalSubscription: false,
      connectedServerInTab: {
        subscriptionServer: Array(100),
        server: false,
      },
      connectedServerInfo: [],
      overHeight: false,
      clipboard: null,
    };
  },
  computed: {
    loadBalanceValid() {
      return localStorage["loadBalanceValid"] === "true";
    },
  },
  watch: {
    "runningState.running"() {
      this.updateConnectView();
    },
    outbound() {
      this.updateConnectView();
    },
    tableData(x) {
      for (const sub of x.subscriptions) {
        sub.status = dayjs(sub.status)
          .tz(dayjs.tz.guess())
          .format("YYYY-MM-DD HH:mm:ss");
      }
    },
    observatory(val) {
      for (const info of val.body.outboundStatus) {
        this.connectedServerInfo.some((x) => {
          if (
            info.which._type === x.which._type &&
            info.which.id === x.which.id &&
            info.which.sub === x.which.sub
          ) {
            for (const k in info) {
              if (k === "which" || !info.hasOwnProperty(k)) {
                continue;
              }
              x.info[k] = info[k];
            }
            return true;
          }
          return false;
        });
      }
      let minDelay = 99999;
      let index = -1;
      this.connectedServerInfo.forEach((x, i) => {
        x.selected = false;
        if (x.info.delay && x.info.delay < minDelay) {
          minDelay = x.info.delay;
          index = i;
        }
      });
      if (index >= 0) {
        this.connectedServerInfo[index].selected = true;
      }
    },
  },
  created() {
    this.$axios({
      url: apiRoot + "/touch",
    }).then((res) => {
      this.refreshTableData(res.data.data.touch, res.data.data.running);
      this.locateTabToConnected();
      this.ready = true;
    });
  },
  beforeDestroy() {
    this.clipboard.destroy();
  },
  mounted() {
    document
      .querySelector("#QRCodeImport")
      .addEventListener("change", this.handleFileChange, false);
    this.clipboard = new ClipboardJS(".sharingAddressTag");
    this.clipboard.on("success", (e) => {
      this.$buefy.toast.open({
        message: this.$t("common.success"),
        type: "is-primary",
        position: "is-top",
        queue: false,
      });
      e.clearSelection();
    });
    this.clipboard.on("error", (e) => {
      this.$buefy.toast.open({
        message: this.$t("common.fail") + ", error:" + e.toLocaleString(),
        type: "is-warning",
        position: "is-top",
        queue: false,
      });
    });
    const that = this;
    let scrollTimer = null;
    window.addEventListener("scroll", (e) => {
      clearTimeout(scrollTimer);
      setTimeout(() => {
        scrollTimer = null;
        that.overHeight = e.target.scrollingElement.scrollTop > 50;
      }, 100);
    });
  },
  methods: {
    refreshTableData(touch, running) {
      touch.servers.forEach((v) => {
        v.connected = false;
      });
      touch.subscriptions.forEach((s) => {
        s.servers.forEach((v) => {
          v.connected = false;
        });
      });
      this.tableData = touch;
      if (running !== undefined) {
        Object.assign(this.runningState, {
          running: running
            ? this.$t("common.isRunning")
            : this.$t("common.notRunning"),
          connectedServer: touch.connectedServer,
        });
      }
    },
    handleClickConnectedServer(which) {
      const that = this;
      this.locateTabToConnected(which);
      let tabIndex = -1;
      if (which._type === "server") {
        tabIndex = 1;
      } else {
        tabIndex = 2 + which.sub;
      }
      let tryCnt = 0;
      const maxTry = 5;
      const tryInterval = 500;

      function waitingAndLocate() {
        if (
          !document
            .querySelector(
              `.main-tabs > .tabs > ul > li:nth-child(${1 + tabIndex})`
            )
            .classList.contains("is-active")
        ) {
          tryCnt++;
          if (tryCnt > maxTry) {
            return;
          }
          setTimeout(waitingAndLocate, tryInterval);
          return;
        }
        console.log("ok");
        that.$nextTick(() => {
          let nodes = document.querySelectorAll(".main-tabs .b-table");
          if (which._type === "subscriptionServer") {
            // solid
            tabIndex = 2;
          }
          nodes = nodes[tabIndex].querySelectorAll("table > tbody > tr");
          const node = Array.from(nodes).find(
            (node) =>
              parseInt(
                node.querySelector('td[data-label="ID"]')?.textContent
              ) === which.id
          );
          if (!node) {
            console.warn("node not found");
            return;
          }
          node.scrollIntoView({ block: "center", inline: "center" });
          let highlightClass = "highlight-row-connected";
          if (that.runningState.running !== that.$t("common.isRunning")) {
            highlightClass = "highlight-row-disconnected";
          }
          node.classList.add(highlightClass);
          setTimeout(() => {
            node.classList.remove(highlightClass);
            setTimeout(() => {
              node.classList.add(highlightClass);
              setTimeout(() => {
                node.classList.remove(highlightClass);
              }, 200);
            }, 50);
          }, 200);
        });
      }

      waitingAndLocate();
    },
    handleClickImportInBatch() {
      this.showModalImportInBatch = true;
    },
    handleModalImportShow() {
      this.$refs.importInput.focus();
    },
    handleImportEnter(event) {
      if (event.keyCode !== 13) {
        return;
      }
      this.handleClickImportConfirm();
    },
    handleFileChange(e) {
      const that = this;
      const file = e.target.files[0];
      let elem = document.querySelector("#QRCodeImport");
      // eslint-disable-next-line no-self-assign
      elem.outerHTML = elem.outerHTML;
      this.$nextTick(() => {
        document
          .querySelector("#QRCodeImport")
          .addEventListener("change", this.handleFileChange, false);
      });
      // console.log(file);
      if (!file.type.match(/image\/.*/)) {
        this.$buefy.toast.open({
          message: this.$t("import.qrcodeError"),
          type: "is-warning",
          position: "is-top",
          queue: false,
        });
        return;
      }
      const reader = new FileReader();
      reader.onload = function (e) {
        // target.result 该属性表示目标对象的DataURL
        // console.log(e.target.result);
        const file = e.target.result;
        const qrcode = new Decoder();
        qrcode
          .scan(file)
          .then((result) => {
            console.log(result);
            that.handleClickImportConfirm(result.data);
          })
          .catch((error) => {
            console.error(error);
            that.$buefy.toast.open({
              message: that.$t("import.qrcodeError"),
              type: "is-warning",
              position: "is-top",
              queue: false,
            });
          });
      };
      reader.readAsDataURL(file);
    },
    sortNumberServers(a, b, isAsc) {
      if (!isAsc) {
        return a.servers.length < b.servers.length ? 1 : -1;
      }
      return a.servers.length > b.servers.length ? 1 : -1;
    },
    sortping(a, b, isAsc) {
      if (isNaN(parseInt(a.pingLatency))) {
        return 1;
      }
      if (isNaN(parseInt(b.pingLatency))) {
        return -1;
      }
      if (!isAsc) {
        return parseInt(a.pingLatency) < parseInt(b.pingLatency) ? 1 : -1;
      } else {
        return parseInt(a.pingLatency) > parseInt(b.pingLatency) ? 1 : -1;
      }
    },
    sortConnections(a, b, isAsc) {
      // when sorted, only connected servers on top
      // desc: error > high ping > low ping > unconnected
      // asc: low ping > high ping > error > unconnected
      if (a.connected && !b.connected) {
        return -1;
      }
      if (!a.connected && b.connected) {
        return 1;
      }
      if (!isAsc) {
        if (isNaN(parseInt(a.pingLatency))) {
          return -1;
        }
        if (isNaN(parseInt(b.pingLatency))) {
          return 1;
        }
        return parseInt(a.pingLatency) < parseInt(b.pingLatency) ? 1 : -1;
      } else {
        if (isNaN(parseInt(a.pingLatency))) {
          return 1;
        }
        if (isNaN(parseInt(b.pingLatency))) {
          return -1;
        }
        return parseInt(a.pingLatency) > parseInt(b.pingLatency) ? 1 : -1;
      }
    },
    filterConnectedServer(servers, outbound = this.outbound) {
      const connectedServers = [];
      if (servers instanceof Array) {
        for (let s of servers) {
          if (s.outbound === outbound) {
            connectedServers.push(s);
          }
        }
        return connectedServers.length ? connectedServers : null;
      }
      return servers;
    },
    updateConnectView() {
      let connectedServer = this.runningState.connectedServer;
      // associate outbounds and servers
      this.runningState.outboundToServerName = {};
      this.runningState.connectedServer?.forEach((cs) => {
        const server = locateServer(this.tableData, cs);
        if (
          this.runningState.outboundToServerName[cs.outbound] &&
          typeof this.runningState.outboundToServerName[cs.outbound] !==
            "number"
        ) {
          this.runningState.outboundToServerName[cs.outbound] = 1;
        }
        if (
          typeof this.runningState.outboundToServerName[cs.outbound] ===
          "number"
        ) {
          this.runningState.outboundToServerName[cs.outbound]++;
        } else {
          this.runningState.outboundToServerName[cs.outbound] = server.name;
        }
      });

      connectedServer = this.filterConnectedServer(connectedServer);
      // clear connected state
      this.tableData.servers.forEach((v) => {
        v.connected && (v.connected = false);
      });
      this.tableData.subscriptions.forEach((s) => {
        s.servers.forEach((v) => {
          v.connected && (v.connected = false);
        });
      });
      if (connectedServer) {
        let server = locateServer(this.tableData, connectedServer);
        if (server instanceof Array) {
          for (const s of server) {
            s.connected = true;
          }
        } else {
          server.connected = true;
          server = [server];
        }
        this.connectedServerInfo = [];
        for (const i in server) {
          let subscription_name = null;
          if (connectedServer[i]._type === "subscriptionServer") {
            subscription_name =
              this.tableData.subscriptions[
                connectedServer[i].sub
              ].host.toUpperCase();
          }
          this.connectedServerInfo.push({
            info: {
              ...server[i],
              subscription_name,
              alive: null,
              delay: null,
              outbound_tag: null,
              last_seen_time: null,
              last_error_reason: null,
              last_try_time: null,
            },
            which: connectedServer[i],
            showContent: true,
            selected: false,
          });
        }
      } else {
        this.connectedServerInfo = [];
      }

      this.connectedServerInfo.sort((x, y) => {
        return x.info.name > y.info.name;
      });

      this.connectedServerInTab.server = false;
      for (const i in this.connectedServerInTab.subscriptionServer) {
        this.connectedServerInTab.subscriptionServer[i] = false;
      }
      if (connectedServer) {
        let servers = connectedServer;
        if (!(connectedServer instanceof Array)) {
          servers = [connectedServer];
        }
        for (const s of servers) {
          if (s._type === "server") {
            this.connectedServerInTab.server = true;
          } else if (s._type === "subscriptionServer") {
            this.connectedServerInTab.subscriptionServer[s.sub] = true;
          }
        }
      }
      this.$emit("input", this.runningState);
    },
    locateTabToConnected(which) {
      let whichServer = which;
      if (!whichServer) {
        whichServer = this.runningState.connectedServer;
      }
      if (!whichServer) {
        return;
      }
      whichServer = this.filterConnectedServer(whichServer);
      if (!whichServer) {
        return;
      }
      if (whichServer instanceof Array) {
        whichServer = whichServer[0];
      }
      let sub = whichServer.sub;
      let subscriptionServersOffset = 2;
      let serversOffset = 1;
      // if (this.tableData.subscriptions.length > 0) {
      //   subscriptionServersOffset++;
      //   serversOffset++;
      // }
      // if (this.tableData.servers.length > 0) {
      //   subscriptionServersOffset++;
      // }
      if (whichServer._type === CONST.SubscriptionServerType) {
        this.tab = sub + subscriptionServersOffset;
      } else if (whichServer._type === CONST.ServerType) {
        this.tab = serversOffset;
      }
    },
    handleClickImportQRCode() {
      document.querySelector("#QRCodeImport").click();
    },
    handleClickImport() {
      this.showModalImport = true;
    },
    handleClickImportConfirm(value) {
      if (typeof value != "string") {
        value = null;
      }
      return this.$axios({
        url: apiRoot + "/import",
        method: "post",
        data: {
          url: value || this.importWhat,
        },
      }).then((res) => {
        if (res.data.code === "SUCCESS") {
          this.refreshTableData(res.data.data.touch, res.data.data.running);
          this.updateConnectView();
          this.$buefy.toast.open({
            message: this.$t("common.success"),
            type: "is-primary",
            position: "is-top",
            queue: false,
          });
          this.showModalImport = false;
          this.showModalImportInBatch = false;
          this.importWhat = "";
        } else {
          this.$buefy.toast.open({
            message: res.data.message,
            type: "is-warning",
            position: "is-top",
            queue: false,
          });
        }
      });
    },
    deleteSelectedServers() {
      this.$axios({
        url: apiRoot + "/touch",
        method: "delete",
        data: {
          touches: this.checkedRows.map((x) => {
            return {
              id: x.id,
              _type: x._type,
            };
          }),
        },
      }).then((res) => {
        if (res.data.code === "SUCCESS") {
          this.refreshTableData(res.data.data.touch, res.data.data.running);
          this.checkedRows = [];
          this.updateConnectView();
        } else {
          this.$buefy.toast.open({
            message: res.data.message,
            type: "is-warning",
            position: "is-top",
            duration: 5000,
            queue: false,
          });
        }
      });
    },
    handleClickDelete() {
      this.$buefy.dialog.confirm({
        title: this.$t("delete.title"),
        message: this.$t("delete.message"),
        confirmText: this.$t("operations.delete"),
        cancelText: this.$t("operations.cancel"),
        type: "is-danger",
        hasIcon: true,
        icon: " iconfont icon-alert",
        onConfirm: () => this.deleteSelectedServers(),
      });
    },
    handleClickAboutConnection(row, sub) {
      let cancel;
      if (!row.connected) {
        //该节点并未处于连接状态，因此进行连接
        waitingConnected(
          this.$axios({
            url: apiRoot + "/connection",
            method: "post",
            data: {
              id: row.id,
              _type: row._type,
              sub: sub,
              outbound: this.outbound,
            },
            cancelToken: new axios.CancelToken(function executor(c) {
              cancel = c;
            }),
          }).then((res) => {
            if (res.data.code === "SUCCESS") {
              Object.assign(this.runningState, {
                running: res.data.data.running
                  ? this.$t("common.isRunning")
                  : this.$t("common.notRunning"),
                connectedServer: res.data.data.touch.connectedServer,
              });
              this.$nextTick(() => {
                this.updateConnectView();
              });
            } else {
              this.$buefy.toast.open({
                message: res.data.message,
                type: "is-warning",
                position: "is-top",
                duration: 5000,
                queue: false,
              });
              this.deleteSelectedServers();
            }
          }),
          3 * 1000,
          cancel
        );
      } else {
        this.$axios({
          url: apiRoot + "/connection",
          method: "delete",
          data: {
            id: row.id,
            _type: row._type,
            sub: sub,
            outbound: this.outbound,
          },
        }).then((res) => {
          if (res.data.code === "SUCCESS") {
            row.connected = false;
            Object.assign(this.runningState, {
              running: res.data.data.running
                ? this.$t("common.isRunning")
                : this.$t("common.notRunning"),
              connectedServer: res.data.data.touch.connectedServer,
            });
            this.updateConnectView();
          } else {
            this.$buefy.toast.open({
              message: res.data.message,
              type: "is-warning",
              position: "is-top",
              duration: 5000,
              queue: false,
            });
          }
        });
      }
    },
    handleClickLatency(ping) {
      let touches = JSON.stringify(
        this.checkedRows.map((x) => {
          //穷举sub
          let sub = this.tableData.subscriptions.findIndex((subscription) =>
            subscription.servers.some((y) => x === y)
          );
          return {
            id: x.id,
            _type: x._type,
            sub: sub === -1 ? null : sub,
          };
        })
      );
      this.checkedRows.forEach((x) => (x.pingLatency = "testing...")); //refresh
      // this.checkedRows = [];
      let timerTip = setTimeout(() => {
        this.$buefy.toast.open({
          message: this.$t("latency.message"),
          type: "is-primary",
          position: "is-top",
          duration: 5000,
          queue: false,
        });
      }, 10 * 1200);
      this.$axios({
        url: apiRoot + (ping ? "/pingLatency" : "/httpLatency"),
        params: {
          whiches: touches,
        },
        timeout: 0,
      })
        .then((res) => {
          handleResponse(
            res,
            this,
            () => {
              res.data.data.whiches.forEach((x) => {
                let server = locateServer(this.tableData, x);
                server.pingLatency = x.pingLatency;
              });
              this.updateConnectView();
            },
            () => {
              this.$buefy.toast.open({
                message: res.data.message,
                type: "is-warning",
                position: "is-top",
                queue: false,
                duration: 5000,
              });
              this.checkedRows.forEach((x) => (x.pingLatency = ""));
            }
          );
        })
        .finally(() => {
          clearTimeout(timerTip);
        });
    },
    // eslint-disable-next-line no-unused-vars
    handleTabsChange(index) {
      this.checkedRows = [];
    },
    isCheckedRowsDeletable() {
      // CONST.SubscriptionServerType is not deletable
      return (
        this.checkedRows.length > 0 &&
        this.checkedRows.every((x) => x._type !== CONST.SubscriptionServerType)
      );
    },
    isCheckedRowsPingable() {
      // CONST.SubscriptionServerType is not deletable
      return (
        this.checkedRows.length > 0 &&
        this.checkedRows.some(
          (x) =>
            x._type === CONST.ServerType ||
            x._type === CONST.SubscriptionServerType
        )
      );
    },
    handleClickShare(row, sub) {
      const TYPE_MAP = {
        [CONST.SubscriptionServerType]: "SERVER",
        [CONST.ServerType]: "SERVER",
        [CONST.SubscriptionType]: "SUBSCRIPTION",
      };
      this.$axios({
        url: apiRoot + "/sharingAddress",
        method: "get",
        params: {
          touch: {
            id: row.id,
            _type: row._type,
            sub,
          },
        },
      }).then((res) => {
        handleResponse(res, this, () => {
          this.$buefy.modal.open({
            width: 500,
            component: ModalSharing,
            props: {
              title: TYPE_MAP[row._type],
              sharingAddress: res.data.data.sharingAddress,
              shortDesc: row.name || row.host || row.address,
              type: row._type,
            },
          });
        });
      });
    },
    handleClickUpdateSubscription(row) {
      this.$axios({
        url: apiRoot + "/subscription",
        method: "put",
        data: {
          id: row.id,
          _type: row._type,
        },
      }).then((res) => {
        handleResponse(res, this, () => {
          this.refreshTableData(res.data.data.touch, res.data.data.running);
          this.updateConnectView();
          this.$buefy.toast.open({
            message: this.$t("common.success"),
            type: "is-primary",
            position: "is-top",
            duration: 5000,
            queue: false,
          });
        });
      });
    },
    handleClickCreate() {
      this.modalServerReadOnly = false;
      this.which = null;
      this.showModalServer = true;
    },
    handleClickModifyServer(row) {
      this.modalServerReadOnly = false;
      this.which = Object.assign({}, row);
      this.which.servers = [];
      this.showModalServer = true;
    },
    handleClickViewServer(row, sub) {
      this.modalServerReadOnly = true;
      this.which = { ...row, sub };
      this.showModalServer = true;
    },
    handleModalServerSubmit(url) {
      this.$axios({
        url: apiRoot + "/import",
        method: "post",
        data: {
          url: url,
          which: this.which,
        },
        timeout: 0,
      }).then((res) => {
        handleResponse(res, this, () => {
          this.$buefy.toast.open({
            message: this.$t("common.success"),
            type: "is-primary",
            position: "is-top",
            duration: 3000,
            queue: false,
          });
          this.showModalServer = false;
          this.refreshTableData(res.data.data.touch, res.data.data.running);
          this.updateConnectView();
        });
      });
    },
    handleClickModifySubscription(row) {
      this.which = Object.assign({}, row);
      this.which.servers = [];
      this.showModalSubscription = true;
    },
    handleModalSubscriptionSubmit(subscription) {
      this.$axios({
        url: apiRoot + "/subscription",
        method: "patch",
        data: {
          subscription,
        },
      }).then((res) => {
        handleResponse(res, this, () => {
          this.$buefy.toast.open({
            message: this.$t("common.success"),
            type: "is-primary",
            position: "is-top",
            duration: 3000,
            queue: false,
          });
          this.showModalSubscription = false;
          this.refreshTableData(res.data.data.touch, res.data.data.running);
          this.updateConnectView();
        });
      });
    },
  },
};
</script>

<style lang="scss" scoped>
td {
  font-size: 0.9em;
}

.node-section {
  margin-top: 1rem;

  .iconfont {
    margin-right: 0.1em;
  }

  .operate-box {
    > * {
      margin-right: 0.5rem;
    }
  }
}

.card {
  max-width: 500px;
  margin: auto;
}

.ping-latency {
  font-size: 0.8em;
}
</style>

<style lang="scss">
@import "~bulma/sass/utilities/all";

#toolbar {
  @media screen and (max-width: 450px) {
    &.float-toolbar {
      top: 4.25rem;
      margin-left: 25px;
      width: calc(100% - 50px);
    }
    .field.is-grouped .field:not(:last-child) {
      margin-right: 0.3rem;
    }
  }

  .field.is-grouped.is-grouped-multiline:last-child {
    margin-bottom: 0;
  }

  padding: 0.75em 0.75em;
  margin-bottom: 1rem;
  position: sticky;
  top: 65px;
  z-index: 2;
  background: rgba(0, 0, 0, 0.05);
  width: 100%;
  border-radius: 3px;
  pointer-events: none;

  * {
    pointer-events: auto;
  }

  .right {
    position: absolute;
    right: 0.75rem;
    top: 0.75em;
    /*max-width: 70%;*/
  }

  transition: all 200ms linear;

  button {
    transition: all 100ms ease-in-out;
  }
}

.tabs {
  .icon + span {
    color: #ff6719; //方案1
  }

  .icon {
    display: none; //方案1
    margin: 0 0 0 -0.5em !important;

    .iconfont {
      font-size: 32px;
      color: coral;
    }
  }
}

tr.is-connected-running {
  $c: #bbdefb;
  background: $c;
  color: findColorInvert($c);
}

tr.is-connected-not-running {
  $c: rgba(255, 69, 58, 0.4);
  background: $c;
  color: findColorInvert($c);
}

@keyframes loading-rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.not-show {
  opacity: 0;
  pointer-events: none !important;
  overflow: hidden;
  width: 0 !important;
  display: inline-block !important;
  padding-left: 0 !important;
  margin-left: 0 !important;
  border-left: 0 !important;
  padding-right: 0 !important;
  margin-right: 0 !important;
  border-right: 0 !important;
}

.not-display {
  display: none;
}

table td,
table th {
  vertical-align: middle !important;
}

.dialog .mdi-.iconfont.icon-alert {
  font-size: 40px;
}

.qrcode#canvas {
  min-height: 300px !important;
  min-width: 300px !important;
}

$coverBackground: rgba(0, 0, 0, 0.6);
.tag-cover {
  height: 100%;
  width: 100%;
  position: absolute;
  top: 0;
  left: 0;
  background-color: $coverBackground !important;
  transition: all 0.5s ease;
  cursor: pointer;
  text-align: center;
  line-height: 22px;
  user-select: none;
}

#tag-cover-text {
  color: findColorInvert($coverBackground);
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: calc(100% - 8px);
  display: none;
  justify-content: center;
  align-items: center;
  z-index: 1;
  font-size: 12px;
  pointer-events: none;
}

.mobile-small {
  @media screen and (max-width: 450px) {
    border-radius: 2px;
    font-size: 0.65rem;
  }
}

.b-sidebar.node-status-sidebar-reduced > .sidebar-content.is-fixed {
  z-index: 1;
  left: 1px;
  top: 4.25rem;
  background-color: white;
  width: unset;
  line-height: 0;
  border-radius: 4px;
}

.b-sidebar.node-status-sidebar > .sidebar-content.is-fixed {
  left: 1px;
  top: 4.25rem;
  background-color: white;
  max-height: calc(100vh - 5rem);
  overflow-y: auto;

  .message {
    cursor: pointer;
  }

  .tabs:not(:last-child),
  .pagination:not(:last-child),
  .message:not(:last-child),
  .level:not(:last-child),
  .breadcrumb:not(:last-child),
  .highlight:not(:last-child),
  .block:not(:last-child),
  .title:not(:last-child),
  .subtitle:not(:last-child),
  .table-container:not(:last-child),
  .table:not(:last-child),
  .progress:not(:last-child),
  .notification:not(:last-child),
  .content:not(:last-child),
  .box:not(:last-child) {
    margin-bottom: 0.25rem;
  }
}

tr.highlight-row-connected {
  transition: background-color 0.05s linear;
  background-color: #a8cff0;
}

tr.highlight-row-disconnected {
  transition: background-color 0.05s linear;
  background-color: rgba(255, 69, 58, 0.55);
}

.click-through {
  pointer-events: none;
}
</style>
