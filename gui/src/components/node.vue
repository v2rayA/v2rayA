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
      this.locateTabToConnected(val.connectedServer);
      this.$emit("input", this.runningState);
    }
  },
  created() {
    this.$axios({
      url: apiRoot + "/touch"
    })
      .then(res => {
        this.tableData = res.data.data.touch;
        // this.$store.commit("CONNECTED_SERVER", this.tableData.connectedServer);
        this.runningState = {
          running: res.data.data.running ? CONST.IS_RUNNING : CONST.NOT_RUNNING,
          connectedServer: this.tableData.connectedServer,
          lastConnectedServer: null
        };
        this.locateTabToConnected();
        this.ready = true;
      })
      .catch(err => {
        this.$buefy.snackbar.open({
          message: err,
          type: "is-warning",
          position: "is-top"
        });
        if (err.message === "Network Error") {
          console.log("todo"); //TODO
        }
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
                this.$buefy.snackbar.open({
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
      console.log(this.checkedRows);
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
            running: this.tableData.running
              ? CONST.IS_RUNNING
              : CONST.NOT_RUNNING,
            connectedServer: this.tableData.connectedServer,
            lastConnectedServer: null
          });
          this.updateConnectView();
        } else {
          this.$buefy.snackbar.open({
            message: res.data.message,
            type: "is-warning",
            position: "is-top"
          });
        }
      });
    },
    handleClickAboutConnection(node, sub) {
      console.log(node);
      if (!node.connected) {
        //该节点并未处于连接状态，因此进行连接
        this.$axios({
          url: apiRoot + "/connection",
          method: "post",
          data: { ...node, sub: sub }
        }).then(res => {
          if (res.data.code === "SUCCESS") {
            Object.assign(this.runningState, {
              running: CONST.IS_RUNNING,
              connectedServer: res.data.data.connectedServer,
              lastConnectedServer: res.data.data.lastConnectedServer
            });
            this.updateConnectView();
          } else {
            this.$buefy.snackbar.open({
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
            node.connected = false;
            Object.assign(this.runningState, {
              running: CONST.NOT_RUNNING,
              connectedServer: null,
              lastConnectedServer: res.data.data.lastConnectedServer
            });
            this.updateConnectView();
          } else {
            this.$buefy.snackbar.open({
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
      this.checkedRows = [];
      this.$axios({
        url: apiRoot + "/pingLatency",
        params: {
          touches
        }
      }).then(res => {
        handleResponse(res, this, () => {
          res.data.data.touches.forEach(x => {
            let server = locateServer(this.tableData, x);
            server.pingLatency = x.pingLatency;
          });
          this.updateConnectView();
        });
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
.icon-loading_ico-copy {
  font-size: 2.5rem;
  color: rgba(0, 0, 0, 0.45);
  animation: loading-rotate 2s infinite linear;
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
table td, table th{
  vertical-align: middle !important;
}
</style>
