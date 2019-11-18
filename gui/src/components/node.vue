<template>
  <section class="node-section container hero">
    <div v-if="ready" class="hero-body">
      <b-field
        grouped
        group-multiline
        style="margin-bottom: 1rem;position: relative"
      >
        <div>
          <button
            :class="{
              button: true,
              field: true,
              'is-info': true,
              'not-display': !isCheckedRowsPingable()
            }"
            @click="handleClickPing"
          >
            <i class="iconfont icon-wave"></i>
            <span>Ping</span>
          </button>
          <button
            :class="{
              button: true,
              field: true,
              'is-delete': true,
              'not-display': !isCheckedRowsDeletable()
            }"
            @click="handleClickDelete"
          >
            <i class="iconfont icon-delete"></i>
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
            <i class="iconfont icon-delete"></i>
            <span>placeholder</span>
          </button>
          <span class="field not-show">placeholder</span>
        </div>
        <div style="position:absolute;right:0;top:0">
          <b-button class="field" type="is-primary" @click="checkedRows = []">
            <i class="iconfont icon-chuangjiangongdan1"></i>
            <span>创建</span>
          </b-button>
          <b-button class="field" type="is-primary" @click="handleClickImport">
            <i class="iconfont icon-daoruzupu-xianxing"></i>
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
              我们支持以ss、vmess地址，或者订阅地址的方式导入，也支持手动创建节点，快来试试吧！
            </p>
          </div>
        </div>
        <footer class="card-footer">
          <a class="card-footer-item">创建</a>
          <a class="card-footer-item" @click="handleClickImport">导入</a>
        </footer>
      </b-collapse>

      <b-tabs
        v-model="tab"
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
                <b-table-column field="status" label="更新状态">
                  {{ props.row.status }}
                </b-table-column>
                <b-table-column label="节点数">
                  {{ props.row.servers.length }}
                </b-table-column>
                <b-table-column label="操作">
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
                      :outlined="!props.row.connected"
                      type="is-info"
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
        <b-tab-item v-if="!!tableData.servers.length" label="SERVER">
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
                  label="Ping时延"
                  class="ping-latency"
                >
                  {{ props.row.pingLatency }}
                </b-table-column>
                <!--            <b-table-column field="httpLatency" label="HTTP时延" width="100">-->
                <!--              {{ props.row.httpLatency }}-->
                <!--            </b-table-column>-->
                <b-table-column label="操作">
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
          :label="sub.host.toUpperCase()"
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
                  label="Ping时延"
                  class="ping-latency"
                >
                  {{ props.row.pingLatency }}
                </b-table-column>
                <!--            <b-table-column field="httpLatency" label="HTTP时延" width="100">-->
                <!--              {{ props.row.httpLatency }}-->
                <!--            </b-table-column>-->
                <b-table-column label="操作">
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
      <i class="iconfont icon-loading_ico-copy"></i>
    </b-loading>
  </section>
</template>

<script>
import { locateServer, handleResponse } from "@/assets/js/utils";
import CONST from "@/assets/js/const";
import QRCode from "qrcode";
import ClipboardJS from "clipboard";
import { Base64 } from "js-base64";

export default {
  name: "Node",
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
      }
    };
  },
  watch: {
    "runningState.running"() {
      console.log("watch runningState.running:", this.runningState);
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
        console.log(server);
      }
      if (runningState.connectedServer) {
        console.log(
          "updateConnectView",
          this.tableData,
          runningState.connectedServer
        );
        let server = locateServer(this.tableData, runningState.connectedServer);
        server.connected = true;
        console.log(server);
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
        console.log("locate to", whichServer);
      } else if (whichServer._type === CONST.ServerType) {
        this.tab = serversOffset;
      }
    },
    handleClickImport() {
      const that = this;
      this.$buefy.dialog.prompt({
        message: `填入ss/vmess/订阅地址`,
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
      console.log(row);
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
    handleClickPing() {
      console.log(this.checkedRows);
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
      this.checkedRows.forEach(x => x.pingLatency && (x.pingLatency = "")); //refresh
      this.checkedRows = [];
      this.$axios({
        url: apiRoot + "/pingLatency",
        params: {
          whiches: touches
        }
      }).then(res => {
        handleResponse(res, this, () => {
          res.data.data.whiches.forEach(x => {
            let server = locateServer(this.tableData, x);
            server.pingLatency = x.pingLatency;
          });
          this.updateConnectView();
        });
      });
    },
    // eslint-disable-next-line no-unused-vars
    handleTabsChange(index) {
      // this.checkedRows = [];
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
                                <span>
                                    ${row.name || row.host}
                                </span>
                            </span>
                            <div id="tag-cover-text">点击复制</div>
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
          console.log(row);
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
          this.$buefy.toast.open({
            message: "更新完成",
            type: "is-primary",
            position: "is-top",
            duration: 5000
          });
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
  pointer-events: none;
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
