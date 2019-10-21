<template>
  <section class="node-section container hero">
    <div class="hero-body">
      <b-field
        grouped
        group-multiline
        style="margin-bottom: 1rem;position: relative"
      >
        <button
          class="button field is-danger"
          @click="checkedRows = []"
          :disabled="!checkedRows.length"
        >
          <i class="iconfont icon-delete"></i>
          <span>删除</span>
        </button>
        <div style="position:absolute;right:0;">
          <b-button class="field" type="is-primary" @click="checkedRows = []">
            <i class="iconfont icon-chuangjiangongdan1"></i>
            <span>创建</span>
          </b-button>
          <b-button class="field" type="is-primary" @click="checkedRows = []">
            <i class="iconfont icon-daoruzupu-xianxing"></i>
            <span>导入</span>
          </b-button>
        </div>
      </b-field>
      <b-field label="SERVER">
        <b-table
          :data="tableData.server"
          :checked-rows.sync="checkedRows"
          checkable
          :row-class="(row, index) => row.selected && 'is-connected'"
        >
          <template slot-scope="props">
            <b-table-column field="id" label="ID" width="40" numeric>
              {{ props.row.id }}
            </b-table-column>
            <b-table-column field="name" label="节点名" width="150">
              {{ props.row.name }}
            </b-table-column>
            <b-table-column field="address" label="节点地址" width="250">
              {{ props.row.address }}
            </b-table-column>
            <b-table-column field="type" label="传输协议/加密方式" width="160">
              {{ props.row.type }}
            </b-table-column>
            <b-table-column field="pingLatency" label="Ping时延" width="100">
              {{ props.row.pingLatency }}
            </b-table-column>
            <b-table-column field="httpLatency" label="HTTP时延" width="100">
              {{ props.row.httpLatency }}
            </b-table-column>
            <b-table-column label="操作">
              <div class="operate-box">
                <b-button
                  size="is-small"
                  icon-left=" github-circle iconfont icon-wendangxiugai"
                  :outlined="!props.row.selected"
                  :type="props.row.selected ? 'is-dark' : ''"
                >
                  修改
                </b-button>
                <b-button
                  size="is-small"
                  :icon-left="
                    ` github-circle iconfont ${
                      props.row.selected
                        ? 'icon-Link_disconnect'
                        : 'icon-lianjie'
                    }`
                  "
                  :outlined="!props.row.selected"
                  :type="props.row.selected ? 'is-warning' : 'is-success'"
                >
                  {{ props.row.selected ? "断开" : "连接" }}
                </b-button>
                <b-button
                  size="is-small"
                  icon-left=" github-circle iconfont icon-share"
                  :outlined="!props.row.selected"
                  type="is-danger"
                >
                  分享
                </b-button>
              </div>
            </b-table-column>
          </template>
        </b-table>
      </b-field>
      <b-field label="SUBSCRIPTION">
        <b-table
          :data="tableData.subscription"
          :checked-rows.sync="checkedRows"
          checkable
        >
          <template slot-scope="props">
            <b-table-column field="id" label="ID" width="40" numeric>
              {{ props.row.id }}
            </b-table-column>
            <b-table-column field="host" label="域名" width="150">
              {{ props.row.host }}
            </b-table-column>
            <b-table-column field="status" label="更新状态" width="250">
              {{ props.row.status }}
            </b-table-column>
            <b-table-column label="节点数" width="360">
              {{ props.row.servers.length }}
            </b-table-column>
            <b-table-column label="操作">
              <div class="operate-box">
                <b-button
                  size="is-small"
                  icon-left=" github-circle iconfont icon-wendangxiugai"
                  :outlined="!props.row.selected"
                  :type="props.row.selected ? 'is-dark' : ''"
                >
                  修改
                </b-button>
                <b-button
                  size="is-small"
                  icon-left=" github-circle iconfont icon-sync"
                  outlined
                  type="is-info"
                >
                  更新
                </b-button>
                <b-button
                  size="is-small"
                  icon-left=" github-circle iconfont icon-share"
                  outlined
                  type="is-danger"
                >
                  分享
                </b-button>
              </div>
            </b-table-column>
          </template>
        </b-table>
      </b-field>
      <b-field
        v-for="sub of tableData.subscription"
        :key="sub.id"
        :label="sub.host.toUpperCase()"
      >
        <b-table :data="sub.servers">
          <template slot-scope="props">
            <b-table-column field="id" label="ID" width="92" numeric>
              {{ props.row.id }}
            </b-table-column>
            <b-table-column field="name" label="节点名" width="150">
              {{ props.row.name }}
            </b-table-column>
            <b-table-column field="address" label="节点地址" width="250">
              {{ props.row.address }}
            </b-table-column>
            <b-table-column field="type" label="传输协议/加密方式" width="160">
              {{ props.row.type }}
            </b-table-column>
            <b-table-column field="pingLatency" label="Ping时延" width="100">
              {{ props.row.pingLatency }}
            </b-table-column>
            <b-table-column field="httpLatency" label="HTTP时延" width="100">
              {{ props.row.httpLatency }}
            </b-table-column>
            <b-table-column label="操作">
              <div class="operate-box">
                <b-button
                  size="is-small"
                  icon-left=" github-circle iconfont icon-winfo-icon-chakanbaogao"
                  :outlined="!props.row.selected"
                  :type="props.row.selected ? 'is-dark' : ''"
                >
                  查看
                </b-button>
                <b-button
                  size="is-small"
                  :icon-left="
                    ` github-circle iconfont ${
                      props.row.selected
                        ? 'icon-Link_disconnect'
                        : 'icon-lianjie'
                    }`
                  "
                  :outlined="!props.row.selected"
                  :type="props.row.selected ? 'is-danger' : 'is-success'"
                >
                  {{ props.row.selected ? "断开" : "连接" }}
                </b-button>
                <b-button
                  size="is-small"
                  icon-left=" github-circle iconfont icon-share"
                  :outlined="!props.row.selected"
                  type="is-danger"
                >
                  分享
                </b-button>
              </div>
            </b-table-column>
          </template>
        </b-table>
      </b-field>
    </div>
  </section>
</template>

<script>
export default {
  name: "node",
  data() {
    return {
      tableData: {
        server: [
          {
            id: 1,
            name: "大猪蹄",
            address: "192.168.50.111:4677",
            type: "kcp",
            selected: true
          }
        ],
        subscription: [
          {
            id: 1,
            host: "hardss.net",
            status: "上次更新：2019-10-21 19:40:47",
            servers: [
              {
                id: 1,
                name: "我好hard啊",
                address: "71.222.50.111:443",
                type: "tcp"
              },
              {
                id: 2,
                name: "我也好hard啊",
                address: "171.112.32.77:443",
                type: "tcp"
              }
            ]
          }
        ]
      },
      checkedRows: []
    };
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
@import "~bulma/sass/utilities/_all";
tr.is-connected {
  $c: #23d160;
  background: $c;
  color: findColorInvert($c);
}
</style>
