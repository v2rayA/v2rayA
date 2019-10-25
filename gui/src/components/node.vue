<template>
  <section class="node-section container hero">
    <div v-if="ready" class="hero-body">
      <b-field
        grouped
        group-multiline
        style="margin-bottom: 1rem;position: relative"
      >
        <button
          v-show="!!checkedRows.length"
          class="button field is-delete"
          @click="handleClickDelete"
        >
          <i class="iconfont icon-delete"></i>
          <span>删除</span>
        </button>
        <div style="position:absolute;right:0;">
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

      <b-tabs v-model="tab" position="is-centered" type="is-toggle-rounded">
        <b-tab-item label="SERVER">
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
                <b-table-column field="net" label="传输协议/加密方式">
                  {{ props.row.net }}
                </b-table-column>
                <b-table-column field="pingLatency" label="Ping时延">
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
        <b-tab-item label="SUBSCRIPTION">
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
        <b-tab-item
          v-for="(sub, subi) of tableData.subscriptions"
          :key="sub.id"
          :label="sub.host.toUpperCase()"
        >
          <b-field :label="`${sub.host.toUpperCase()}(${sub.servers.length})`">
            <b-table
              :data="sub.servers"
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
                <b-table-column field="net" label="传输协议/加密方式">
                  {{ props.row.net }}
                </b-table-column>
                <b-table-column field="pingLatency" label="Ping时延">
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
import { locateServer } from "@/assets/js/utils";
import CONST from "@/assets/const";
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
      if (whichServer._type === "subscription") {
        this.tab = sub + 2;
        console.log("locate to", whichServer);
      } else if (whichServer._type === "server") {
        this.tab = 0;
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
      // this.checkedRows = [];
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
        this.tableData = res.data.data.touch;
        this.checkedRows = [];
      });
    },
    handleClickAboutConnection(node, i) {
      console.log(node);
      if (!node.connected) {
        //该节点并未处于连接状态，因此进行连接
        this.$axios({
          url: apiRoot + "/connection",
          method: "post",
          data: { ...node, sub: i }
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
</style>
