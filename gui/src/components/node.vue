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
        <div style="max-width: 50%">
          <button
            :class="{
              button: true,
              field: true,
              'is-info': true,
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
              'not-display': !overHeight && !isCheckedRowsDeletable()
            }"
            :disabled="!isCheckedRowsDeletable()"
            @click="handleClickDelete"
          >
            <i class="iconfont icon-delete" />
            <span>删除</span>
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
          <b-button class="field" type="is-primary" @click="handleClickCreate">
            <i class="iconfont icon-chuangjiangongdan1" />
            <span>创建</span>
          </b-button>
          <b-button class="field" type="is-primary" @click="handleClickImport">
            <i class="iconfont icon-daoruzupu-xianxing" />
            <span>导入</span>
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
            初来乍到，请多关照
          </p>
          <a class="card-header-icon">
            <b-icon :icon="props.open ? 'menu-down' : 'menu-up'"> </b-icon>
          </a>
        </div>
        <div class="card-content">
          <div class="content">
            <p>我们发现你还没有创建或导入任何节点、订阅。</p>
            <p>
              我们支持以vmess、ss、ssr地址，或者订阅地址的方式导入，也支持手动创建节点，快来试试吧！
            </p>
          </div>
        </div>
        <footer class="card-footer">
          <a class="card-footer-item" @click="handleClickCreate">创建</a>
          <a class="card-footer-item" @click="handleClickImport">导入</a>
        </footer>
      </b-collapse>

      <b-tabs
        :value="tab"
        position="is-centered"
        type="is-toggle-rounded"
        @change="handleTabsChange"
      >
        <b-tab-item
          v-if="!!tableData.subscriptions.length"
          label="SUBSCRIPTION"
        >
          <b-field :label="`SUBSCRIPTION(${tableData.subscriptions.length})`">
            <b-table
              :data="tableData.subscriptions"
              :checked-rows.sync="checkedRows"
              :row-class="(row, index) => row.connected && 'is-connected'"
              checkable
            >
              <template slot-scope="props">
                <b-table-column field="id" label="ID" numeric>
                  {{ props.row.id }}
                </b-table-column>
                <b-table-column field="host" label="域名">
                  {{ props.row.host }}
                </b-table-column>
                <b-table-column field="remarks" label="别名">
                  {{ props.row.remarks }}
                </b-table-column>
                <b-table-column field="status" label="更新状态" width="260">
                  {{ props.row.status }}
                </b-table-column>
                <b-table-column label="节点数" centered>
                  {{ props.row.servers.length }}
                </b-table-column>
                <b-table-column label="操作" width="250">
                  <div class="operate-box">
                    <b-button
                      size="is-small"
                      icon-left=" github-circle iconfont icon-sync"
                      outlined
                      type="is-warning"
                      @click="handleClickUpdateSubscription(props.row)"
                    >
                      更新
                    </b-button>
                    <b-button
                      size="is-small"
                      icon-left=" github-circle iconfont icon-wendangxiugai"
                      outlined
                      type="is-info"
                      @click="handleClickModifySubscription(props.row)"
                    >
                      修改
                    </b-button>
                    <b-button
                      size="is-small"
                      icon-left=" github-circle iconfont icon-share"
                      outlined
                      type="is-success"
                      @click="handleClickShare(props.row)"
                    >
                      分享
                    </b-button>
                  </div>
                </b-table-column>
              </template>
            </b-table>
          </b-field>
        </b-tab-item>
        <b-tab-item
          v-if="!!tableData.servers.length"
          label="SERVER"
          :icon="
            `${connectedServer._type === 'server' ? ' iconfont icon-dian' : ''}`
          "
        >
          <b-field :label="`SERVER(${tableData.servers.length})`">
            <b-table
              :data="tableData.servers"
              :checked-rows.sync="checkedRows"
              checkable
              :row-class="(row, index) => row.connected && 'is-connected'"
            >
              <template slot-scope="props">
                <b-table-column field="id" label="ID" numeric>
                  {{ props.row.id }}
                </b-table-column>
                <b-table-column field="name" label="节点名">
                  {{ props.row.name }}
                </b-table-column>
                <b-table-column field="address" label="节点地址">
                  {{ props.row.address }}
                </b-table-column>
                <b-table-column field="net" label="协议">
                  {{ props.row.net }}
                </b-table-column>
                <b-table-column
                  field="pingLatency"
                  label="时延"
                  class="ping-latency"
                >
                  {{ props.row.pingLatency }}
                </b-table-column>
                <!--            <b-table-column field="httpLatency" label="HTTP时延" width="100">-->
                <!--              {{ props.row.httpLatency }}-->
                <!--            </b-table-column>-->
                <b-table-column label="操作" width="250">
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
                      {{ props.row.connected ? "断开" : "连接" }}
                    </b-button>
                    <b-button
                      size="is-small"
                      icon-left=" github-circle iconfont icon-wendangxiugai"
                      :outlined="!props.row.connected"
                      type="is-info"
                      @click="handleClickModifyServer(props.row)"
                    >
                      修改
                    </b-button>
                    <b-button
                      size="is-small"
                      icon-left=" github-circle iconfont icon-share"
                      :outlined="!props.row.connected"
                      type="is-success"
                      @click="handleClickShare(props.row)"
                    >
                      分享
                    </b-button>
                  </div>
                </b-table-column>
              </template>
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
          <b-field :label="`${sub.host.toUpperCase()}(${sub.servers.length})`">
            <b-table
              :data="sub.servers"
              :checked-rows.sync="checkedRows"
              checkable
              :row-class="(row, index) => row.connected && 'is-connected'"
            >
              <template slot-scope="props">
                <b-table-column field="id" label="ID" numeric>
                  {{ props.row.id }}
                </b-table-column>
                <b-table-column field="name" label="节点名">
                  {{ props.row.name }}
                </b-table-column>
                <b-table-column field="address" label="节点地址">
                  {{ props.row.address }}
                </b-table-column>
                <b-table-column field="net" label="协议">
                  {{ props.row.net }}
                </b-table-column>
                <b-table-column
                  field="pingLatency"
                  label="时延"
                  class="ping-latency"
                >
                  {{ props.row.pingLatency }}
                </b-table-column>
                <b-table-column label="操作" width="250">
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
                      {{ props.row.connected ? "断开" : "连接" }}
                    </b-button>
                    <b-button
                      size="is-small"
                      icon-left=" github-circle iconfont icon-winfo-icon-chakanbaogao"
                      :outlined="!props.row.connected"
                      type="is-info"
                      @click="handleClickViewServer(props.row, subi)"
                    >
                      查看
                    </b-button>
                    <b-button
                      size="is-small"
                      icon-left=" github-circle iconfont icon-share"
                      :outlined="!props.row.connected"
                      type="is-success"
                      @click="handleClickShare(props.row, subi)"
                    >
                      分享
                    </b-button>
                  </div>
                </b-table-column>
              </template>
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
import ClipboardJS from "clipboard";
import { Base64 } from "js-base64";
import ModalServer from "@/components/modalServer";
import ModalSubscription from "./modalSuscription";

export default {
  name: "Node",
  components: { ModalSubscription, ModalServer },
  data() {
    return {
      tableData: {
        servers: [],
        subscriptions: [],
        connectedServer: {}
      },
      checkedRows: [],
      ready: false,
      tab: 0,
      runningState: {
        running: CONST.INSPECTING_RUNNING,
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
      overHeight: false
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
        running: res.data.data.running ? CONST.IS_RUNNING : CONST.NOT_RUNNING,
        connectedServer: this.tableData.connectedServer,
        lastConnectedServer: null
      };
      this.locateTabToConnected();
      this.ready = true;
    });
  },
  mounted() {
    let clipboard = new ClipboardJS(".sharingAddressTag");
    clipboard.on("success", e => {
      this.$buefy.toast.open({
        message: "复制成功",
        type: "is-primary",
        position: "is-top"
      });
      e.clearSelection();
    });
    clipboard.on("error", e => {
      this.$buefy.toast.open({
        message: "复制失败，error:" + e.toLocaleString(),
        type: "is-warning",
        position: "is-top"
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
    updateConnectView(runningState) {
      if (!runningState) {
        runningState = this.runningState;
      }
      if (runningState.lastConnectedServer) {
        let server = locateServer(
          this.tableData,
          runningState.lastConnectedServer
        );
        server.connected = false;
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
      let subscriptionServersOffset = 0;
      let serversOffset = 0;
      if (this.tableData.subscriptions.length > 0) {
        subscriptionServersOffset++;
        serversOffset++;
      }
      if (this.tableData.servers.length > 0) {
        subscriptionServersOffset++;
      }
      if (whichServer._type === CONST.SubscriptionServerType) {
        this.tab = sub + subscriptionServersOffset;
      } else if (whichServer._type === CONST.ServerType) {
        this.tab = serversOffset;
      }
    },
    handleClickImport() {
      const that = this;
      this.$buefy.dialog.prompt({
        message: `填入ss/ssr/vmess/订阅地址`,
        inputAttrs: {
          type: "text",
          value: ""
        },
        trapFocus: true,
        onConfirm: value => {
          return that
            .$axios({
              url: apiRoot + "/import",
              method: "post",
              data: {
                url: value
              }
            })
            .then(res => {
              if (res.data.code === "SUCCESS") {
                this.tableData = res.data.data.touch;
                this.runningState = {
                  running: res.data.data.running
                    ? CONST.IS_RUNNING
                    : CONST.NOT_RUNNING,
                  connectedServer: this.tableData.connectedServer,
                  lastConnectedServer: null
                };
                this.updateConnectView();
                this.$buefy.toast.open({
                  message: "导入成功",
                  type: "is-primary",
                  position: "is-top",
                  queue: false
                });
              } else {
                this.$buefy.toast.open({
                  message: res.data.message,
                  type: "is-warning",
                  position: "is-top"
                });
              }
            });
        }
      });
    },
    handleClickDelete() {
      this.$buefy.dialog.confirm({
        title: "删除节点/订阅",
        message: "确定要<b>删除</b>这些节点/订阅吗？注意，该操作是不可逆的。",
        confirmText: "删除",
        cancelText: "取消",
        type: "is-danger",
        hasIcon: true,
        icon: " iconfont icon-alert",
        onConfirm: () =>
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
                  ? CONST.IS_RUNNING
                  : CONST.NOT_RUNNING,
                connectedServer: this.tableData.connectedServer,
                lastConnectedServer: null
              });
              this.updateConnectView();
            } else {
              this.$buefy.toast.open({
                message: res.data.message,
                type: "is-warning",
                position: "is-top"
              });
            }
          })
      });
    },
    handleClickAboutConnection(row, sub) {
      if (!row.connected) {
        //该节点并未处于连接状态，因此进行连接
        this.$axios({
          url: apiRoot + "/connection",
          method: "post",
          data: { id: row.id, _type: row._type, sub: sub }
        }).then(res => {
          if (res.data.code === "SUCCESS") {
            Object.assign(this.runningState, {
              running: CONST.IS_RUNNING,
              connectedServer: res.data.data.connectedServer,
              lastConnectedServer: res.data.data.lastConnectedServer
            });
            this.updateConnectView();
          } else {
            this.$buefy.toast.open({
              message: res.data.message,
              type: "is-warning",
              position: "is-top"
            });
          }
        });
      } else {
        this.$axios({
          url: apiRoot + "/connection",
          method: "delete"
        }).then(res => {
          if (res.data.code === "SUCCESS") {
            row.connected = false;
            Object.assign(this.runningState, {
              running: CONST.NOT_RUNNING,
              connectedServer: null,
              lastConnectedServer: res.data.data.lastConnectedServer
            });
            this.updateConnectView();
          } else {
            this.$buefy.toast.open({
              message: res.data.message,
              type: "is-warning",
              position: "is-top"
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
          message: "时延测试往往需要花费较长时间，请耐心等待",
          type: "is-primary",
          position: "is-top",
          duration: 5000
        });
      }, 10 * 1200);
      this.$axios({
        url: apiRoot + (ping ? "/pingLatency" : "/httpLatency"),
        params: {
          whiches: touches
        }
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
                            <div id="tag-cover-text">复制链接</div>
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
                        <a class="is-link" href="https://github.com/mzz2017/V2RayA" target="_blank">
                          <img class="leave-right" src="https://img.shields.io/github/stars/mzz2017/V2RayA.svg?style=social" alt="stars">
                          <img class="leave-right" src="https://img.shields.io/github/forks/mzz2017/V2RayA.svg?style=social" alt="forks">
                          <img class="leave-right" src="https://img.shields.io/github/watchers/mzz2017/V2RayA.svg?style=social" alt="watchers">
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
                console.log("QRCode has been generated successfully!");
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
              ? CONST.IS_RUNNING
              : CONST.NOT_RUNNING,
            connectedServer: this.tableData.connectedServer,
            lastConnectedServer: null
          };
          this.updateConnectView();
          this.$buefy.toast.open({
            message: "更新完成",
            type: "is-primary",
            position: "is-top",
            duration: 5000
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
        }
      }).then(res => {
        handleResponse(res, this, () => {
          this.$buefy.toast.open({
            message: "操作成功",
            type: "is-primary",
            position: "is-top",
            duration: 3000
          });
          this.showModalServer = false;
          this.tableData = res.data.data.touch;
          this.runningState = {
            running: res.data.data.running
              ? CONST.IS_RUNNING
              : CONST.NOT_RUNNING,
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
          message: "修改订阅别名需要V2RayA版本高于0.5.0",
          type: "is-warning",
          queue: false,
          position: "is-top",
          duration: 3000,
          actionText: "查看帮助",
          onAction: () => {
            window.open(
              "https://github.com/mzz2017/V2RayA#%E4%BD%BF%E7%94%A8",
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
            message: "操作成功",
            type: "is-primary",
            position: "is-top",
            duration: 3000
          });
          this.showModalSubscription = false;
          this.tableData = res.data.data.touch;
          this.runningState = {
            running: res.data.data.running
              ? CONST.IS_RUNNING
              : CONST.NOT_RUNNING,
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
@import "../../node_modules/bulma/sass/utilities/all";
#toolbar {
  padding: 0.75em 0.75em 0;
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
    max-width: 50%;
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
</style>
