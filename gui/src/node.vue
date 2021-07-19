<template>
  <section id="node-section" class="node-section container hero">
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
            : 'transparent'
        }"
      >
        <div style="max-width: 60%">
          <button
            :class="{
              button: true,
              field: true,
              'is-info': true,
              'mobile-small': true,
              'not-display': !overHeight && !isCheckedRowsPingable()
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
              'not-display': !overHeight && !isCheckedRowsPingable()
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
              'not-display': !overHeight && !isCheckedRowsDeletable()
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
              'not-show': true
            }"
          >
            <i class="iconfont icon-delete" />
            <span>placeholder</span>
          </button>
          <span class="field not-show">placeholder</span>
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
        class="card"
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
        @change="handleTabsChange"
      >
        <b-tab-item label="SUBSCRIPTION">
          <b-field :label="`SUBSCRIPTION(${tableData.subscriptions.length})`">
            <b-table
              :data="tableData.subscriptions"
              :checked-rows.sync="checkedRows"
              :row-class="(row, index) => row.connected && 'is-connected'"
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
          :icon="
            `${connectedServer._type === 'server' ? ' iconfont icon-dian' : ''}`
          "
        >
          <b-field :label="`SERVER(${tableData.servers.length})`">
            <b-table
              :paginated="tableData.servers.length > 150"
              per-page="100"
              :current-page.sync="currentPage.servers"
              :data="tableData.servers"
              :checked-rows.sync="checkedRows"
              checkable
              :row-class="(row, index) => row.connected && 'is-connected'"
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
                width="300"
              >
                <div class="operate-box">
                  <b-button
                    size="is-small"
                    :icon-left="
                      ` github-circle iconfont ${
                        props.row.connected
                          ? 'icon-Link_disconnect'
                          : 'icon-lianjie'
                      }`
                    "
                    :outlined="!props.row.connected"
                    :type="props.row.connected ? 'is-warning' : 'is-warning'"
                    @click="handleClickAboutConnection(props.row)"
                  >
                    {{
                      props.row.connected
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
          :icon="
            `${
              connectedServer._type === 'subscriptionServer' &&
              connectedServer.sub === subi
                ? ' iconfont icon-dian'
                : ''
            }`
          "
        >
          <b-field
            v-if="tab === subi + 2"
            :label="
              `${sub.host.toUpperCase()}(${sub.servers.length}${
                sub.info ? ') (' : ''
              }${sub.info})`
            "
          >
            <b-table
              :paginated="sub.servers.length >= 150"
              :current-page.sync="currentPage[sub.id]"
              per-page="100"
              :data="sub.servers"
              :checked-rows.sync="checkedRows"
              checkable
              :row-class="(row, index) => row.connected && 'is-connected'"
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
                style="font-size:0.9em"
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
                width="300"
              >
                <div class="operate-box">
                  <b-button
                    size="is-small"
                    :icon-left="
                      ` github-circle iconfont ${
                        props.row.connected
                          ? 'icon-Link_disconnect'
                          : 'icon-lianjie'
                      }`
                    "
                    :outlined="!props.row.connected"
                    :type="props.row.connected ? 'is-warning' : 'is-warning'"
                    @click="handleClickAboutConnection(props.row, subi)"
                  >
                    {{
                      props.row.connected
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
            style="display:flex;justify-content:flex-end;width: -moz-available;"
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
            style="display:flex;justify-content:flex-end;width: -moz-available;"
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
import {
  locateServer,
  handleResponse,
  isVersionGreaterEqual
} from "@/assets/js/utils";
import CONST from "@/assets/js/const";
import QRCode from "qrcode";
import jsqrcode from "./assets/js/jsqrcode";
import ClipboardJS from "clipboard";
import { Base64 } from "js-base64";
import ModalServer from "@/components/modalServer";
import ModalSubscription from "@/components/modalSuscription";
import { waitingConnected } from "@/assets/js/networkInspect";
import axios from "@/plugins/axios";

export default {
  name: "Node",
  components: { ModalSubscription, ModalServer },
  data() {
    return {
      importWhat: "",
      showModalImport: false,
      showModalImportInBatch: false,
      currentPage: { servers: 1, subscriptions: 1 },
      tableData: {
        servers: [],
        subscriptions: [],
        connectedServer: {}
      },
      checkedRows: [],
      ready: false,
      tab: 0,
      runningState: {
        running: this.$t("common.checkRunning"),
        connectedServer: null,
        lastConnectedServer: null
      },
      showModalServer: false,
      which: null,
      modalServerReadOnly: false,
      showModalSubscription: false,
      connectedServer: {
        _type: "",
        id: 0,
        sub: 0
      },
      overHeight: false,
      clipboard: null
    };
  },
  watch: {
    "runningState.running"() {
      let val = this.runningState;
      this.updateConnectView(val);
      this.$emit("input", this.runningState);
    }
  },
  created() {
    this.$axios({
      url: apiRoot + "/touch"
    }).then(res => {
      this.tableData = res.data.data.touch;
      // this.$store.commit("CONNECTED_SERVER", this.tableData.connectedServer);
      this.runningState = {
        running: res.data.data.running
          ? this.$t("common.isRunning")
          : this.$t("common.notRunning"),
        connectedServer: this.tableData.connectedServer,
        lastConnectedServer: null
      };
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
    this.clipboard.on("success", e => {
      this.$buefy.toast.open({
        message: this.$t("common.success"),
        type: "is-primary",
        position: "is-top",
        queue: false
      });
      e.clearSelection();
    });
    this.clipboard.on("error", e => {
      this.$buefy.toast.open({
        message: this.$t("common.fail") + ", error:" + e.toLocaleString(),
        type: "is-warning",
        position: "is-top",
        queue: false
      });
    });
    const that = this;
    let scrollTimer = null;
    window.addEventListener("scroll", e => {
      clearTimeout(scrollTimer);
      setTimeout(() => {
        scrollTimer = null;
        that.overHeight = e.target.scrollingElement.scrollTop > 50;
      }, 100);
    });
  },
  methods: {
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
          queue: false
        });
        return;
      }
      const reader = new FileReader();
      reader.onload = function(e) {
        // target.result 该属性表示目标对象的DataURL
        // console.log(e.target.result);
        const file = e.target.result;
        jsqrcode.callback = result => {
          console.log(result);
          if (result !== "error decoding QR Code") {
            that.handleClickImportConfirm(result);
          } else {
            that.$buefy.toast.open({
              message: that.$t("import.qrcodeError"),
              type: "is-warning",
              position: "is-top",
              queue: false
            });
          }
        };
        jsqrcode.decode(file);
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
    updateConnectView(runningState) {
      if (!runningState) {
        runningState = this.runningState;
      }
      if (runningState.lastConnectedServer) {
        let server = locateServer(
          this.tableData,
          runningState.lastConnectedServer
        );
        if (server.connected) {
          server.connected = false;
        } else {
          //否则广播
          this.tableData.servers.some(v => {
            v.connected && (v = false);
            return v.connected;
          }) ||
            this.tableData.subscriptions.some(s => {
              return s.servers.some(v => {
                v.connected && (v = false);
                return v.connected;
              });
            });
        }
      }
      if (runningState.connectedServer) {
        let server = locateServer(this.tableData, runningState.connectedServer);
        server.connected = true;
      }
      this.connectedServer = Object.assign(
        this.connectedServer,
        runningState.connectedServer
      );
      if (!runningState.connectedServer) {
        this.connectedServer._type = "";
      }
    },
    locateTabToConnected(whichServer) {
      if (!whichServer) {
        whichServer = this.tableData.connectedServer;
      }
      if (!whichServer) {
        return;
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
          url: value || this.importWhat
        }
      }).then(res => {
        if (res.data.code === "SUCCESS") {
          this.tableData = res.data.data.touch;
          this.runningState = {
            running: res.data.data.running
              ? this.$t("common.isRunning")
              : this.$t("common.notRunning"),
            connectedServer: this.tableData.connectedServer,
            lastConnectedServer: null
          };
          this.updateConnectView();
          this.$buefy.toast.open({
            message: this.$t("common.success"),
            type: "is-primary",
            position: "is-top",
            queue: false
          });
          this.showModalImport = false;
          this.showModalImportInBatch = false;
          this.importWhat = "";
        } else {
          this.$buefy.toast.open({
            message: res.data.message,
            type: "is-warning",
            position: "is-top",
            queue: false
          });
        }
      });
    },
    syncConnectedServer() {
      this.$axios({
        url: apiRoot + "/touch",
        method: "delete",
        data: {
          touches: this.checkedRows.map(x => {
            return {
              id: x.id,
              _type: x._type
            };
          })
        }
      }).then(res => {
        if (res.data.code === "SUCCESS") {
          this.tableData = res.data.data.touch;
          this.checkedRows = [];
          Object.assign(this.runningState, {
            running: res.data.data.running
              ? this.$t("common.isRunning")
              : this.$t("common.notRunning"),
            connectedServer: this.tableData.connectedServer,
            lastConnectedServer: null
          });
          this.updateConnectView();
        } else {
          this.$buefy.toast.open({
            message: res.data.message,
            type: "is-warning",
            position: "is-top",
            duration: 5000,
            queue: false
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
        onConfirm: () => this.syncConnectedServer()
      });
    },
    handleClickAboutConnection(row, sub) {
      const that = this;
      let cancel;
      if (!row.connected) {
        //该节点并未处于连接状态，因此进行连接
        waitingConnected(
          this.$axios({
            url: apiRoot + "/connection",
            method: "post",
            data: { id: row.id, _type: row._type, sub: sub },
            cancelToken: new axios.CancelToken(function executor(c) {
              cancel = c;
            })
          }).then(res => {
            if (res.data.code === "SUCCESS") {
              Object.assign(this.runningState, {
                running: this.$t("common.isRunning"),
                connectedServer: res.data.data.connectedServer,
                lastConnectedServer: res.data.data.lastConnectedServer
              });
              this.updateConnectView();
            } else {
              this.$buefy.toast.open({
                message: res.data.message,
                type: "is-warning",
                position: "is-top",
                duration: 5000,
                queue: false
              });
              this.syncConnectedServer();
            }
          }),
          3 * 1000,
          cancel
        );
      } else {
        this.$axios({
          url: apiRoot + "/connection",
          method: "delete"
        }).then(res => {
          if (res.data.code === "SUCCESS") {
            row.connected = false;
            Object.assign(this.runningState, {
              running: that.$t("common.notRunning"),
              connectedServer: null,
              lastConnectedServer: res.data.data.lastConnectedServer
            });
            this.updateConnectView();
          } else {
            this.$buefy.toast.open({
              message: res.data.message,
              type: "is-warning",
              position: "is-top",
              duration: 5000,
              queue: false
            });
          }
        });
      }
    },
    handleClickLatency(ping) {
      let touches = JSON.stringify(
        this.checkedRows.map(x => {
          //穷举sub
          let sub = this.tableData.subscriptions.findIndex(subscription =>
            subscription.servers.some(y => x === y)
          );
          return {
            id: x.id,
            _type: x._type,
            sub: sub === -1 ? null : sub
          };
        })
      );
      this.checkedRows.forEach(x => (x.pingLatency = "testing...")); //refresh
      // this.checkedRows = [];
      let timerTip = setTimeout(() => {
        this.$buefy.toast.open({
          message: this.$t("latency.message"),
          type: "is-primary",
          position: "is-top",
          duration: 5000,
          queue: false
        });
      }, 10 * 1200);
      this.$axios({
        url: apiRoot + (ping ? "/pingLatency" : "/httpLatency"),
        params: {
          whiches: touches
        },
        timeout: 0
      })
        .then(res => {
          handleResponse(
            res,
            this,
            () => {
              res.data.data.whiches.forEach(x => {
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
                duration: 5000
              });
              this.checkedRows.forEach(x => (x.pingLatency = ""));
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
        this.checkedRows.every(x => x._type !== CONST.SubscriptionServerType)
      );
    },
    isCheckedRowsPingable() {
      // CONST.SubscriptionServerType is not deletable
      return (
        this.checkedRows.length > 0 &&
        this.checkedRows.some(
          x =>
            x._type === CONST.ServerType ||
            x._type === CONST.SubscriptionServerType
        )
      );
    },
    handleClickShare(row, sub) {
      const TYPE_MAP = {
        [CONST.SubscriptionServerType]: "SERVER",
        [CONST.ServerType]: "SERVER",
        [CONST.SubscriptionType]: "SUBSCRIPTION"
      };
      this.$axios({
        url: apiRoot + "/sharingAddress",
        method: "get",
        params: {
          touch: {
            id: row.id,
            _type: row._type,
            sub
          }
        }
      }).then(res => {
        handleResponse(res, this, () => {
          this.$buefy.modal.open({
            width: 500,
            content: `
<div class="modal-card" style="max-width: 500px;margin:auto">
                    <header class="modal-card-head">
                        <p class="modal-card-title has-text-centered">${
                          TYPE_MAP[row._type]
                        }</p>
                    </header>
                    <section class="modal-card-body lazy" style="text-align: center">
                        <div><canvas id="canvas" class="qrcode"></canvas></div>
                        <div class="tags has-addons is-centered" style="position: relative">
                            <span class="tag is-rounded is-dark sharingAddressTag" style="position: relative" data-clipboard-text="${
                              res.data.data.sharingAddress
                            }">
                                <div class="tag-cover tag is-rounded" style="display: none;"></div>
                                <span class="has-ellipsis" style="max-width:10em">
                                    ${row.name || row.host || row.address}
                                </span>
                            </span>
                            <div id="tag-cover-text">${this.$t(
                              "operations.copyLink"
                            )}</div>
                            <span class="tag is-rounded is-primary sharingAddressTag" style="position: relative" data-clipboard-text="${
                              res.data.data.sharingAddress
                            }">
                                <span class="has-ellipsis" style="max-width:25em">
                                  ${res.data.data.sharingAddress}
                                </span>
                                <div class="tag-cover tag is-rounded" style="display: none;"></div>
                            </span>
                        </div>
                    </section>
                    <footer class="modal-card-foot" style="justify-content: center">
                        <a class="is-link" href="https://github.com/v2rayA/v2rayA" target="_blank">
                          <img class="leave-right" src="https://img.shields.io/github/stars/mzz2017/v2rayA.svg?style=social" alt="stars">
                          <img class="leave-right" src="https://img.shields.io/github/forks/mzz2017/v2rayA.svg?style=social" alt="forks">
                          <img class="leave-right" src="https://img.shields.io/github/watchers/mzz2017/v2rayA.svg?style=social" alt="watchers">
                        </a>
                    </footer>
                </div>
`
          });
          this.$nextTick(() => {
            let add = res.data.data.sharingAddress;
            if (row._type === CONST.SubscriptionType) {
              add = "sub://" + Base64.encode(add);
            }
            let canvas = document.getElementById("canvas");
            QRCode.toCanvas(
              canvas,
              add,
              { errorCorrectionLevel: "H" },
              function(error) {
                if (error) console.error(error);
                // console.log("QRCode has been generated successfully!");
              }
            );
            let targets = document.querySelectorAll(".sharingAddressTag");
            let covers = document.querySelectorAll(".tag-cover");
            let coverText = document.querySelector("#tag-cover-text");
            let enter = () => {
              covers.forEach(x => (x.style.display = "unset"));
              coverText.style.display = "flex";
            };
            let leave = () => {
              covers.forEach(x => (x.style.display = "none"));
              coverText.style.display = "none";
            };
            targets.forEach(x => x.addEventListener("mouseenter", enter));
            targets.forEach(x => x.addEventListener("mouseleave", leave));
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
          _type: row._type
        }
      }).then(res => {
        handleResponse(res, this, () => {
          this.tableData = res.data.data.touch;
          this.runningState = {
            running: res.data.data.running
              ? this.$t("common.isRunning")
              : this.$t("common.notRunning"),
            connectedServer: this.tableData.connectedServer,
            lastConnectedServer: null
          };
          this.updateConnectView();
          this.$buefy.toast.open({
            message: this.$t("common.success"),
            type: "is-primary",
            position: "is-top",
            duration: 5000,
            queue: false
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
          which: this.which
        },
        timeout: 0
      }).then(res => {
        handleResponse(res, this, () => {
          this.$buefy.toast.open({
            message: this.$t("common.success"),
            type: "is-primary",
            position: "is-top",
            duration: 3000,
            queue: false
          });
          this.showModalServer = false;
          this.tableData = res.data.data.touch;
          this.runningState = {
            running: res.data.data.running
              ? this.$t("common.isRunning")
              : this.$t("common.notRunning"),
            connectedServer: this.tableData.connectedServer,
            lastConnectedServer: null
          };
          this.updateConnectView();
        });
      });
    },
    handleClickModifySubscription(row) {
      if (!isVersionGreaterEqual(localStorage["version"], "0.5.0")) {
        this.$buefy.snackbar.open({
          message: this.$t("version.higherVersionNeeded"),
          type: "is-warning",
          queue: false,
          position: "is-top",
          duration: 3000,
          actionText: this.$t("operations.helpManual"),
          onAction: () => {
            window.open(
              "https://github.com/v2rayA/v2rayA#%E4%BD%BF%E7%94%A8",
              "_blank"
            );
          }
        });
        return;
      }
      this.which = Object.assign({}, row);
      this.which.servers = [];
      this.showModalSubscription = true;
    },
    handleModalSubscriptionSubmit(subscription) {
      this.$axios({
        url: apiRoot + "/subscription",
        method: "patch",
        data: {
          subscription
        }
      }).then(res => {
        handleResponse(res, this, () => {
          this.$buefy.toast.open({
            message: this.$t("common.success"),
            type: "is-primary",
            position: "is-top",
            duration: 3000,
            queue: false
          });
          this.showModalSubscription = false;
          this.tableData = res.data.data.touch;
          this.runningState = {
            running: res.data.data.running
              ? this.$t("common.isRunning")
              : this.$t("common.notRunning"),
            connectedServer: this.tableData.connectedServer,
            lastConnectedServer: null
          };
          this.updateConnectView();
        });
      });
    }
  }
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
  .field.is-grouped .field:not(:last-child) {
    @media screen and (max-width: 450px) {
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

tr.is-connected {
  //$c: #23d160;
  $c: #bbdefb;
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
    font-size: 0.75rem;
  }
}
</style>
