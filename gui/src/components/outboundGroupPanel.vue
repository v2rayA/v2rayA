<template>
  <div class="ogp-wrapper">
    <b-tag
      class="pointerTag"
      type="is-info"
      :icon-right="menuOpen ? 'menu-up' : 'menu-down'"
      role="button"
      tabindex="0"
      @click.native.stop="toggleMenu"
      @keydown.native.enter.prevent.stop="toggleMenu"
      @keydown.native.space.prevent.stop="toggleMenu"
    >
      {{ $t("common.proxyGroups") }}: {{ currentOutbound.toUpperCase() }}
    </b-tag>

    <!-- Persistent expandable menu: close only on outside click or manual toggle -->
    <div v-if="menuOpen" class="ogp-panel" @click.stop>
        <div
          v-for="outbound in outbounds"
          :key="outbound"
          class="ogp-group-block"
        >
          <!-- Group header row -->
          <div
            class="ogp-group-row"
            :class="{
              'ogp-group-row--current': outbound === currentOutbound,
              'ogp-group-row--open': expandedGroup === outbound,
            }"
            @click="handleClickGroup(outbound)"
          >
            <span class="ogp-group-label">
              <span
                v-if="outbound === currentOutbound"
                class="mdi mdi-circle has-text-success"
                style="font-size: 0.5em; vertical-align: middle; margin-right: 4px"
              ></span>
              {{ outbound.toUpperCase() }}
            </span>
            <b-tag
              v-if="getGroupCount(outbound) > 0"
              size="is-small"
              type="is-warning"
              rounded
              style="margin-left: auto; margin-right: 4px"
            >{{ getGroupCount(outbound) }}</b-tag>
            <span
              class="mdi"
              :class="expandedGroup === outbound ? 'mdi-chevron-down' : 'mdi-chevron-right'"
              style="color: #aaa"
            ></span>
          </div>

          <!-- Expanded: nodes in this group -->
          <div v-if="expandedGroup === outbound" class="ogp-nodes-inline" @click.stop>
            <div class="ogp-nodes-actions">
              <b-button
                size="is-small"
                type="is-warning"
                outlined
                icon-left="plus"
                @click.stop="openNodePicker(outbound)"
              >{{ $t("operations.addTo") }}</b-button>
              <b-button
                v-if="outbound !== 'proxy'"
                size="is-small"
                type="is-danger"
                outlined
                icon-left="delete"
                style="margin-left: auto"
                @click.stop="deleteGroup(outbound)"
              >{{ $t("operations.delete") }}</b-button>
            </div>
            <div class="ogp-nodes-scroll">
              <div
                v-for="node in getGroupNodes(outbound)"
                :key="node.key"
                class="ogp-node-row"
              >
                <span class="mdi mdi-circle has-text-warning" style="font-size: 0.7em"></span>
                <span class="ogp-node-name" :title="node.name">{{ node.name }}</span>
                <span
                  class="mdi mdi-close ogp-node-del"
                  :title="$t('operations.disconnect')"
                  @click.stop="disconnectNode(node)"
                ></span>
              </div>
              <div v-if="!getGroupNodes(outbound).length" class="ogp-nodes-empty">
                {{ $t("common.empty") || "暂无节点" }}
              </div>
            </div>
          </div>
        </div>

        <hr class="dropdown-divider" style="margin: 4px 0" />
        <div class="ogp-group-row ogp-group-row--add" @click.stop="$emit('add-outbound')">
          <span class="mdi mdi-plus"></span> {{ $t("operations.addOutbound") }}
        </div>
    </div>

    <!-- Node picker modal (outside dropdown to avoid z-index issues) -->
    <b-modal :active.sync="showPicker" has-modal-card trap-focus>
      <div class="modal-card" style="max-width: 520px; margin: auto">
        <header class="modal-card-head">
          <p class="modal-card-title">
            {{ $t("operations.addTo") }}
            <b-tag type="is-info" size="is-medium" style="margin-left: 8px">
              {{ pickerGroup ? pickerGroup.toUpperCase() : "" }}
            </b-tag>
          </p>
        </header>
        <section class="modal-card-body" style="min-height: 200px; max-height: 60vh; overflow-y: auto">
          <b-input
            v-model="nodeSearch"
            placeholder="搜索节点..."
            icon="magnify"
            style="margin-bottom: 0.75rem"
          ></b-input>
          <div v-if="loadingNodes" style="text-align: center; padding: 2rem">
            <b-loading :is-full-page="false" :active="true"></b-loading>
          </div>
          <template v-else>
            <div
              v-for="node in filteredNodes"
              :key="node.key"
              class="ogp-picker-row"
              :class="{ 'ogp-picker-row--highlight': isPickerNodeHighlighted(node) }"
              @click="toggleNode(node)"
            >
              <span
                class="mdi"
                :class="pickerNodeIconClass(node)"
                style="font-size: 1.1em; margin-right: 6px; flex-shrink: 0"
              ></span>
              <span class="ogp-picker-name">{{ node.name }}</span>
              <b-tag
                v-if="isRunningInGroup(node, pickerGroup)"
                size="is-small"
                type="is-warning"
                rounded
                style="margin-left: 6px; flex-shrink: 0"
              >{{ $t("common.isRunning") }}</b-tag>
              <b-tag v-if="node.subName" size="is-small" type="is-light" style="margin-left: auto; flex-shrink: 0">
                {{ node.subName }}
              </b-tag>
            </div>
            <div v-if="!filteredNodes.length && !loadingNodes" style="text-align: center; padding: 1rem; color: #888">
              暂无节点
            </div>
          </template>
        </section>
        <footer class="modal-card-foot flex-end">
          <b-button @click="showPicker = false">{{ $t("operations.close") }}</b-button>
          <b-button type="is-primary" :loading="saving" @click="savePickerChanges">
            {{ $t("operations.save") }}
          </b-button>
        </footer>
      </div>
    </b-modal>
  </div>
</template>

<script>
import i18n from "@/plugins/i18n";

export default {
  name: "OutboundGroupPanel",
  i18n,
  props: {
    outbounds: { type: Array, default: () => ["proxy"] },
    currentOutbound: { type: String, default: "proxy" },
    isMobile: { type: Boolean, default: false },
  },
  data() {
    return {
      menuOpen: false,
      expandedGroup: null,
      touchData: null,
      isCoreRunning: false,
      showPicker: false,
      pickerGroup: null,
      draftSelectionMap: {},
      nodeSearch: "",
      loadingNodes: false,
      saving: false,
    };
  },
  computed: {
    // Map: outbound -> array of {key, name, id, _type, sub, subName}
    groupNodeMap() {
      const map = {};
      if (!this.touchData) return map;
      const connectedServers = this.touchData.connectedServer || [];
      for (const cs of connectedServers) {
        const outbound = cs.outbound || "proxy";
        if (!map[outbound]) map[outbound] = [];
        const serverObj = this.findServer(cs);
        if (serverObj) {
          map[outbound].push({
            key: `${cs._type}-${cs.id}-${cs.sub ?? "na"}-${outbound}`,
            name: serverObj.name || serverObj.address || String(cs.id),
            id: cs.id,
            _type: cs._type,
            sub: cs.sub,
            subName: cs._type === "subscriptionServer" && this.touchData.subscriptions
              ? (this.touchData.subscriptions[cs.sub]?.host || null)
              : null,
          });
        }
      }
      return map;
    },
    // Flat list of all nodes for the picker
    allNodes() {
      if (!this.touchData) return [];
      const result = [];
      // regular servers
      for (const s of this.touchData.servers || []) {
        result.push({
          key: `server-${s.id}`,
          name: s.name || s.address || String(s.id),
          id: s.id,
          _type: "server",
          sub: undefined,
          subName: null,
        });
      }
      // subscription servers
      const subs = this.touchData.subscriptions || [];
      for (let i = 0; i < subs.length; i++) {
        const sub = subs[i];
        for (const s of sub.servers || []) {
          result.push({
            key: `subsrv-${i}-${s.id}`,
            name: s.name || s.address || String(s.id),
            id: s.id,
            _type: "subscriptionServer",
            sub: i,
            subName: sub.host || null,
          });
        }
      }
      return result;
    },
    filteredNodes() {
      const q = this.nodeSearch.trim().toLowerCase();
      if (!q) return this.allNodes;
      return this.allNodes.filter(
        (n) =>
          n.name.toLowerCase().includes(q) ||
          (n.subName && n.subName.toLowerCase().includes(q))
      );
    },
    runningGroupNodeMap() {
      const map = {};
      if (!this.isCoreRunning || !this.touchData) {
        return map;
      }
      const connectedServers = this.touchData.connectedServer || [];
      for (const cs of connectedServers) {
        const outbound = cs.outbound || "proxy";
        if (!map[outbound]) {
          map[outbound] = new Set();
        }
        map[outbound].add(this.whichKey(cs));
      }
      return map;
    },
  },
  methods: {
    async requestSuccess(config, fallbackMessage) {
      const res = await this.$axios(config);
      if (!res || !res.data || res.data.code !== "SUCCESS") {
        throw new Error((res && res.data && res.data.message) || fallbackMessage || "request failed");
      }
      return res;
    },
    whichKey(which) {
      if (which._type === "subscriptionServer") {
        return `subscriptionServer-${which.sub ?? "na"}-${which.id}`;
      }
      return `${which._type}-${which.id}`;
    },
    initDraftSelection(outbound) {
      const nextMap = {};
      const nodes = this.getGroupNodes(outbound);
      for (const node of nodes) {
        nextMap[this.whichKey(node)] = true;
      }
      this.draftSelectionMap = nextMap;
    },
    isDraftSelected(node) {
      return !!this.draftSelectionMap[this.whichKey(node)];
    },
    isRunningInGroup(node, outbound) {
      if (!outbound) {
        return false;
      }
      const set = this.runningGroupNodeMap[outbound];
      return !!(set && set.has(this.whichKey(node)));
    },
    isPickerNodeHighlighted(node) {
      return this.isDraftSelected(node) || this.isRunningInGroup(node, this.pickerGroup);
    },
    pickerNodeIconClass(node) {
      if (this.isPickerNodeHighlighted(node)) {
        return "mdi-check-circle has-text-warning";
      }
      return "mdi-circle-outline has-text-grey-light";
    },
    findServer(cs) {
      if (!this.touchData) return null;
      if (cs._type === "server") {
        return (this.touchData.servers || []).find((s) => s.id === cs.id) || null;
      }
      if (cs._type === "subscriptionServer") {
        const sub = (this.touchData.subscriptions || [])[cs.sub];
        if (!sub) return null;
        return (sub.servers || []).find((s) => s.id === cs.id) || null;
      }
      return null;
    },
    getGroupCount(outbound) {
      return (this.groupNodeMap[outbound] || []).length;
    },
    getGroupNodes(outbound) {
      return this.groupNodeMap[outbound] || [];
    },
    isInGroup(node, outbound) {
      if (!outbound) return false;
      return (this.groupNodeMap[outbound] || []).some((n) => {
        if (n._type !== node._type || n.id !== node.id) return false;
        // For subscriptionServer, also match sub index; for plain server, sub is irrelevant
        if (n._type === "subscriptionServer") return n.sub === node.sub;
        return true;
      });
    },
    async fetchTouchData() {
      try {
        const res = await this.$axios({ url: apiRoot + "/touch" });
        if (res.data.code === "SUCCESS") {
          this.touchData = res.data.data.touch;
          this.isCoreRunning = !!res.data.data.running;
        }
      } catch (_) {}
    },
    toggleMenu() {
      this.menuOpen = !this.menuOpen;
      if (this.menuOpen) {
        this.fetchTouchData();
        this.expandedGroup = this.currentOutbound;
      }
    },
    handleDocumentClick(event) {
      if (!this.$el.contains(event.target)) {
        this.menuOpen = false;
      }
    },
    handleClickGroup(outbound) {
      this.$emit("select", outbound);
      this.expandedGroup = this.expandedGroup === outbound ? null : outbound;
    },
    openNodePicker(outbound) {
      this.pickerGroup = outbound;
      this.nodeSearch = "";
      this.loadingNodes = true;
      this.showPicker = true;
      this.menuOpen = false;
      this.fetchTouchData().finally(() => {
        this.initDraftSelection(outbound);
        this.loadingNodes = false;
      });
    },
    toggleNode(node) {
      if (!this.pickerGroup || this.saving) return;
      const key = this.whichKey(node);
      this.$set(this.draftSelectionMap, key, !this.draftSelectionMap[key]);
    },
    async savePickerChanges() {
      if (!this.pickerGroup || this.saving) {
        return;
      }
      const previousSet = new Set(
        this.getGroupNodes(this.pickerGroup).map((node) => this.whichKey(node))
      );
      const draftSet = new Set(
        Object.keys(this.draftSelectionMap).filter((key) => this.draftSelectionMap[key])
      );
      const toAdd = [];
      const toRemove = [];
      for (const node of this.allNodes) {
        const key = this.whichKey(node);
        if (!previousSet.has(key) && draftSet.has(key)) {
          toAdd.push(node);
        }
        if (previousSet.has(key) && !draftSet.has(key)) {
          toRemove.push(node);
        }
      }
      if (!toAdd.length && !toRemove.length) {
        this.showPicker = false;
        return;
      }
      const restartAfterSave = this.isCoreRunning;
      this.saving = true;
      try {
        const selectedNodes = this.allNodes.filter((node) => {
          const key = this.whichKey(node);
          return draftSet.has(key);
        });
        const touches = selectedNodes.map((node) => ({
          id: node.id,
          _type: node._type,
          sub: node._type === "subscriptionServer" ? node.sub : 0,
          outbound: this.pickerGroup,
        }));
        await this.requestSuccess({
          url: apiRoot + "/outboundConnections",
          method: "put",
          data: {
            outbound: this.pickerGroup,
            touches,
          },
        }, this.$t("common.fail"));
        if (restartAfterSave) {
          await this.requestSuccess({
            url: apiRoot + "/v2ray",
            method: "delete",
          }, this.$t("common.fail"));
          await this.requestSuccess({
            url: apiRoot + "/v2ray",
            method: "post",
          }, this.$t("common.fail"));
        }
        await this.fetchTouchData();
        this.$emit("changed");
        this.showPicker = false;
        this.$buefy.toast.open({
          message: this.$t("common.success"),
          type: "is-success",
          position: "is-top",
          duration: 2000,
          queue: false,
        });
      } catch (err) {
        this.$buefy.toast.open({
          message: err?.response?.data?.message || err?.message || "Save failed",
          type: "is-warning",
          position: "is-top",
          duration: 5000,
          queue: false,
        });
      } finally {
        this.saving = false;
      }
    },
    async disconnectNode(node) {
      try {
        await this.$axios({
          url: apiRoot + "/connection",
          method: "delete",
          data: {
            id: node.id,
            _type: node._type,
            sub: node.sub,
            outbound: this.expandedGroup,
          },
        });
        await this.fetchTouchData();
        this.$emit("changed");
      } catch (_) {}
    },
    async deleteGroup(outbound) {
      try {
        const res = await this.$axios({
          url: apiRoot + "/outbound",
          method: "delete",
          data: { outbound },
        });
        if (res.data.code === "SUCCESS") {
          if (this.expandedGroup === outbound) {
            this.expandedGroup = null;
          }
          this.$emit("group-deleted", res.data.data.outbounds);
        } else {
          // Show error message from server (e.g., bound custom inbounds)
          this.$buefy.toast.open({
            message: res.data.message || this.$t("common.fail"),
            type: "is-warning",
            position: "is-top",
            duration: 8000,
            queue: false,
          });
        }
      } catch (err) {
        // Show error from server response
        const msg = err?.response?.data?.message || err?.message || this.$t("common.fail");
        this.$buefy.toast.open({
          message: msg,
          type: "is-warning",
          position: "is-top",
          duration: 8000,
          queue: false,
        });
      }
    },
  },
  mounted() {
    document.addEventListener("click", this.handleDocumentClick, true);
  },
  beforeDestroy() {
    document.removeEventListener("click", this.handleDocumentClick, true);
  },
};
</script>

<style lang="scss">
.ogp-wrapper {
  display: inline-block;
  position: relative;
}

.ogp-panel {
  position: absolute;
  top: calc(100% + 6px);
  left: 0;
  z-index: 35;
  border: 1px solid #dbdbdb;
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 8px 20px rgba(10, 10, 10, 0.12);
  min-width: 220px;
  max-width: 280px;
  padding: 4px 0;
}

.ogp-group-block {
  display: flex;
  flex-direction: column;
}

.pointerTag {
  cursor: pointer;
}

.ogp-group-row {
  display: flex;
  align-items: center;
  padding: 6px 12px;
  cursor: pointer;
  user-select: none;
  white-space: nowrap;
  transition: background 0.1s;

  &:hover,
  &--open {
    background: #f5f5f5;
  }

  &--current {
    font-weight: 600;
  }

  &--add {
    color: #666;
    font-size: 0.9em;

    &:hover {
      color: #4a9eff;
    }
  }
}

.ogp-group-label {
  flex: 1;
  margin-right: 4px;
}

.ogp-nodes-inline {
  background: #fafafa;
  border-top: 1px solid #f0f0f0;
  border-bottom: 1px solid #f0f0f0;
  margin-bottom: 2px;
}

.ogp-nodes-actions {
  display: flex;
  align-items: center;
  padding: 5px 10px;
  border-bottom: 1px solid #eee;
}

.ogp-nodes-scroll {
  overflow-y: auto;
  max-height: 250px;
  padding: 4px 0;
}

.ogp-node-row {
  display: flex;
  align-items: center;
  padding: 4px 10px;
  font-size: 0.85em;
  gap: 4px;
  background: #fff5cf;
  border: 1px solid #ffe08a;
  border-radius: 6px;
  margin: 3px 8px;
}

.ogp-node-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #856404;
}

.ogp-node-del {
  cursor: pointer;
  color: #bbb;
  flex-shrink: 0;

  &:hover {
    color: #e55;
  }
}

.ogp-nodes-empty {
  padding: 12px 10px;
  color: #aaa;
  font-size: 0.85em;
  text-align: center;
}

.ogp-picker-row {
  display: flex;
  align-items: center;
  padding: 6px 8px;
  border-radius: 4px;
  cursor: pointer;
  transition: background 0.1s;

  &:hover {
    background: #f5f5f5;
  }

  &--highlight {
    background: #fff8e1;

    &:hover {
      background: #fff0c0;
    }
  }
}

.ogp-picker-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

body.theme-dark {
  .ogp-panel {
    background: var(--md-surface-container);
    border-color: var(--md-surface-variant);
    box-shadow: 0 10px 24px rgba(0, 0, 0, 0.45);
  }

  .ogp-group-row {
    color: var(--md-on-surface);

    &:hover,
    &--open {
      background: #2b2930;
    }

    &--add {
      color: var(--md-on-surface-variant);

      &:hover {
        color: var(--md-primary);
      }
    }
  }

  .ogp-nodes-inline {
    background: #26242b;
    border-top-color: var(--md-surface-variant);
    border-bottom-color: var(--md-surface-variant);
  }

  .ogp-nodes-actions {
    border-bottom-color: var(--md-surface-variant);
  }

  .ogp-node-row {
    background: #4a3d21;
    border-color: #7f6420;
  }

  .ogp-node-row:hover {
    background: #564724;
  }

  .ogp-node-name {
    color: #ffdf9d;
  }

  .ogp-nodes-empty {
    color: var(--md-on-surface-variant);
  }

  .ogp-picker-row {
    color: var(--md-on-surface);

    &:hover {
      background: #2b2930;
    }

    &--highlight {
      background: #3b311a;

      &:hover {
        background: #4a3d21;
      }
    }
  }
}
</style>
