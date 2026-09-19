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
      <i class="lucide icon-panel-left sidebar-handle" :title="$t('common.expand')" />
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
        :key="connectedServerKey(v.which)"
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
        <template #header>
          <div class="node-status-card__header">
            <span class="node-status-card__title">{{ formatServerName(v.info) }}</span>
            <span
              v-if="formatOutboundLabel(v.which)"
              class="node-status-card__group"
            >
              {{ formatOutboundLabel(v.which) }}
            </span>
            <span
              v-if="v.info.subscription_name"
              class="node-status-card__subscription"
            >
              {{ v.info.subscription_name }}
            </span>
          </div>
        </template>
        <div v-if="v.showContent" class="node-status-card__body">
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
    <b-notification
      v-if="ready && coreVersionValid === false"
      type="is-danger"
      role="alert"
      :closable="false"
      class="core-version-error"
    >
      <span class="core-version-error__icon"><i class="lucide icon-triangle-alert" /></span>
      <span>{{ $t("version.coreVersionMismatch", { err: coreVersionErr || "" }) }}</span>
    </b-notification>
    <div v-if="ready" class="hero-body">
      <b-field
        id="toolbar"
        grouped
        group-multiline
        :class="{
          'float-toolbar': overHeight,
          'float-toolbar-active': overHeight && (isCheckedRowsPingable() || isCheckedRowsDeletable()),
        }"
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
            <i class="lucide icon-activity" />
            <span>{{ $t("operations.ping") }}</span>
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
            <i class="lucide icon-activity" />
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
            <i class="lucide icon-trash-2" />
            <span>{{ $t("operations.delete") }}</span>
          </button>
          <b-dropdown
            aria-role="list"
            position="is-bottom-left"
            :class="{
              field: true,
              'mobile-small': true,
              'not-display': !overHeight && !isCheckedRowsExportable(),
            }"
          >
            <template #trigger>
              <button class="button is-info mobile-small" :disabled="!isCheckedRowsExportable()">
                <i class="lucide icon-share-2" />
                <span>{{ $t("operations.export") }}</span>
              </button>
            </template>
            <b-dropdown-item aria-role="listitem" @click="handleClickExportSelected('copy')">
              {{ $t("operations.copySelected") }}
            </b-dropdown-item>
            <b-dropdown-item aria-role="listitem" @click="handleClickExportSelected('download')">
              {{ $t("operations.downloadTxt") }}
            </b-dropdown-item>
          </b-dropdown>
        </div>
        <div class="right">
          <b-button
            class="field mobile-small"
            type="is-primary"
            @click="handleClickCreate"
          >
            <i class="lucide icon-square-plus" />
            <span>{{ $t("operations.create") }}</span>
          </b-button>
          <b-button
            class="field mobile-small"
            type="is-primary"
            @click="handleClickImport"
          >
            <i class="lucide icon-download" />
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
            <b-icon :icon="props.open ? 'chevron-down' : 'chevron-up'"></b-icon>
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
        <b-tab-item :label="$t('subscription.subscription')">
          <b-field :label="`${$t('subscription.subscription')}(${tableData.subscriptions.length})`">
            <b-table
              :data="tableData.subscriptions"
              :checked-rows.sync="checkedRows"
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
                    icon-left="refresh-cw"
                    outlined
                    type="is-warning"
                    @click="handleClickUpdateSubscription(props.row)"
                  >
                    {{ $t("operations.update") }}
                  </b-button>
                  <b-button
                    size="is-small"
                    icon-left="pencil"
                    outlined
                    type="is-info"
                    @click="handleClickModifySubscription(props.row)"
                  >
                    {{ $t("operations.modify") }}
                  </b-button>
                  <b-button
                    size="is-small"
                    icon-left="share-2"
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
          :label="$t('server.server')"
          :header-class="connectedServerInTab['server'] ? 'tab-connected' : ''"
        >
          <b-field :label="`${$t('server.server')}(${tableData.servers.length})`">
            <b-table
              per-page="100"
              :current-page.sync="currentPage.servers"
              :data="tableData.servers"
              :checked-rows.sync="checkedRows"
              checkable
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
                <p class="address-column" :title="props.row.address">
                  {{ props.row.address }}
                </p>
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
                <p
                  :class="{
                    'latency-column': true,
                    'latency-valid': props.row.pingLatency.endsWith('ms'),
                  }"
                  :title="props.row.pingLatency"
                >
                  {{ props.row.pingLatency }}
                </p>
              </b-table-column>
              <b-table-column
                v-slot="props"
                :label="$t('operations.name')"
                sortable
                :custom-sort="sortConnections"
                width="300"
              >
                <div class="operate-box">
                  <b-dropdown
                    v-if="loadBalanceValid"
                    position="is-bottom-left"
                  >
                    <b-button
                      slot="trigger"
                      size="is-small"
                      type="is-primary"
                      icon-right="chevron-down"
                    >
                      {{ $t("operations.addTo") }}
                    </b-button>
                    <b-dropdown-item
                      v-for="group in outbounds"
                      :key="group"
                      @click="toggleNodeInGroup(props.row, undefined, group)"
                    >
                      <span
                        class="node-group-option"
                        :class="{ 'node-group-option--active': isNodeInOutbound(props.row, undefined, group) }"
                      >
                        {{ group.toUpperCase() }}
                      </span>
                    </b-dropdown-item>
                  </b-dropdown>
                  <b-button
                    v-else
                    size="is-small"
                    :icon-left="props.row.connected ? 'unlink' : 'link'"
                    :outlined="!props.row.connected"
                    :type="props.row.connected ? 'is-warning' : 'is-primary'"
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
                    icon-left="pencil"
                    :outlined="!props.row.connected"
                    type="is-info"
                    @click="handleClickModifyServer(props.row)"
                  >
                    {{ $t("operations.modify") }}
                  </b-button>
                  <b-button
                    size="is-small"
                    icon-left="share-2"
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
          :header-class="connectedServerInTab['subscriptionServer'][subi] ? 'tab-connected' : ''"
        >
          <b-field
            v-if="tab === subi + 2"
            :label="`${sub.host.toUpperCase()} (${sub.servers.length})${sub.info ? ' · ' + sub.info : ''}`"
          >
            <b-table
              :current-page.sync="currentPage[sub.id]"
              per-page="100"
              :data="sub.servers"
              :checked-rows.sync="checkedRows"
              checkable
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
                <p class="address-column" :title="props.row.address">
                  {{ props.row.address }}
                </p>
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
                <p
                  :class="{
                    'latency-column': true,
                    'latency-valid': props.row.pingLatency.endsWith('ms'),
                  }"
                  :title="props.row.pingLatency"
                >
                  {{ props.row.pingLatency }}
                </p>
              </b-table-column>
              <b-table-column
                v-slot="props"
                :label="$t('operations.name')"
                sortable
                :custom-sort="sortConnections"
                width="300"
              >
                <div class="operate-box">
                  <b-dropdown
                    v-if="loadBalanceValid"
                    position="is-bottom-left"
                  >
                    <b-button
                      slot="trigger"
                      size="is-small"
                      type="is-primary"
                      icon-right="chevron-down"
                    >
                      {{ $t("operations.addTo") }}
                    </b-button>
                    <b-dropdown-item
                      v-for="group in outbounds"
                      :key="group"
                      @click="toggleNodeInGroup(props.row, subi, group)"
                    >
                      <span
                        class="node-group-option"
                        :class="{ 'node-group-option--active': isNodeInOutbound(props.row, subi, group) }"
                      >
                        {{ group.toUpperCase() }}
                      </span>
                    </b-dropdown-item>
                  </b-dropdown>
                  <b-button
                    v-else
                    size="is-small"
                    :icon-left="props.row.connected ? 'unlink' : 'link'"
                    :outlined="!props.row.connected"
                    :type="props.row.connected ? 'is-warning' : 'is-primary'"
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
                    icon-left="file-text"
                    :outlined="!props.row.connected"
                    type="is-info"
                    @click="handleClickViewServer(props.row, subi)"
                  >
                    {{ $t("operations.view") }}
                  </b-button>
                  <b-button
                    size="is-small"
                    icon-left="share-2"
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
      <i class="lucide icon-loader-circle" />
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
          <b-field>
            <b-radio-button
              v-model="importKind"
              native-value="server"
              type="is-primary is-light"
              size="is-small"
            >
              {{ $t("import.server") }}
            </b-radio-button>
            <b-radio-button
              v-model="importKind"
              native-value="subscription"
              type="is-primary is-light"
              size="is-small"
            >
              {{ $t("import.subscription") }}
            </b-radio-button>
          </b-field>
          {{
            importKind === "subscription"
              ? $t("import.subscriptionMessage")
              : $t("import.serverMessage")
          }}
          <b-input
            ref="importInput"
            v-model="importWhat"
            icon-right="camera"
            icon-right-clickable
            @icon-right-click="handleClickImportQRCode"
            @keyup.native="handleImportEnter"
          ></b-input>
        </section>
        <footer class="modal-card-foot">
          <button
            v-if="importKind === 'server'"
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
              :class="{ 'is-loading': importing }"
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
              :class="{ 'is-loading': importing }"
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
  backendMessage,
  handleResponse,
  locateServer,
} from "@/assets/js/utils";
import CONST from "@/assets/js/const";
import QRCode from "qrcode";
import { Decoder } from "@nuintun/qrcode";
import ClipboardJS from "clipboard";
import { Base64 } from "js-base64";
import ModalServer from "@/components/modalServer";
import ModalSubscription from "@/components/modalSubcription";
import ModalSharing from "@/components/modalSharing";
import ModalPickProxyGroup from "@/components/modalPickProxyGroup";
import { waitingConnected } from "@/assets/js/networkInspect";
import axios from "@/plugins/axios";
import dayjs from "dayjs";
import i18n from "@/plugins/i18n";

// vue-i18n locale -> dayjs locale (all loaded in plugins/dayjs.js)
const DAYJS_LOCALES = { zh: "zh-cn", en: "en", fa: "fa", ru: "ru", pt: "pt-br", ko: "ko" };

export default {
  name: "Node",
  components: { ModalSubscription, ModalServer },
  filters: {
    unix2datetime(x) {
      x = dayjs.unix(x);
      return dayjs().locale(DAYJS_LOCALES[i18n.locale] || "en").to(x);
    },
  },
  props: {
    outbound: {
      type: String,
      default: "proxy",
    },
    outbounds: {
      type: Array,
      default() {
        return ["proxy"];
      },
    },
    // The /version response arrives after this component is created, so the
    // core-version banner used to render the previous page load's state until
    // the next reload. The parent owns it now.
    coreVersionValid: {
      type: Boolean,
      default: true,
    },
    coreVersionErr: {
      type: String,
      default: "",
    },
    loadBalanceValid: {
      type: Boolean,
      default: false,
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
      importing: false,
      importKind: "server",
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
        networkPaused: false,
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
      scrollTimer: null,
    };
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
    if (!localStorage["token"]) return; // Not authenticated yet — skip to avoid spurious 401 modals
    const loadTouch = (retries = 3) => {
      this.$axios({
        url: apiRoot + "/touch",
      }).then((res) => {
        if (res.data && res.data.code === "SUCCESS") {
          this.refreshTableData(res.data.data.touch, res.data.data.running, res.data.data.networkPaused);
          this.updateConnectView();
          this.locateTabToConnected();
          this.ready = true;
        } else if (retries > 0) {
          // Non-SUCCESS response (e.g. server busy) — retry after a short delay
          setTimeout(() => loadTouch(retries - 1), 2000);
        } else {
          // Give up retrying; unblock the UI so the user isn't stuck on a spinner
          this.ready = true;
        }
      }).catch(() => {
        // Network error (can happen during Tun route setup on Windows) — retry
        if (retries > 0) {
          setTimeout(() => loadTouch(retries - 1), 2000);
        } else {
          this.ready = true;
        }
      });
    };
    loadTouch();
  },
  beforeDestroy() {
    this.clipboard.destroy();
    window.removeEventListener("scroll", this.handleWindowScroll);
    clearTimeout(this.scrollTimer);
  },
  mounted() {
    document
      .querySelector("#QRCodeImport")
      .addEventListener("change", this.handleFileChange, false);
    this.clipboard = new ClipboardJS(".sharingAddressTag");
    this.clipboard.on("success", (e) => {
      this.$buefy.toast.open({
        message: this.$t("sharing.copied"),
        type: "is-primary",
        position: "is-top",
      });
      e.clearSelection();
    });
    this.clipboard.on("error", (e) => {
      this.$buefy.toast.open({
        message: this.$t("sharing.copyFailed"),
        type: "is-warning",
        position: "is-top",
      });
    });
    window.addEventListener("scroll", this.handleWindowScroll);

    // if lastNodeTab in the local storage, set it as the current tab.
    const { lastNodeTab } = localStorage;
    if (lastNodeTab !== undefined) {
      this.tab = parseInt(lastNodeTab);
    }
  },
  methods: {
    handleWindowScroll(e) {
      clearTimeout(this.scrollTimer);
      this.scrollTimer = setTimeout(() => {
        this.overHeight = e.target.scrollingElement.scrollTop > 50;
      }, 100);
    },
    getRunningLabel(running, networkPaused = false) {
      if (networkPaused) {
        return this.$t("common.waitingNetwork");
      }
      return running ? this.$t("common.isRunning") : this.$t("common.notRunning");
    },
    connectedServerKey(which = {}) {
      const parts = [
        which._type || "unknown",
        which.sub !== undefined ? which.sub : "na",
        which.id || "0",
        which.outbound || "default",
      ];
      return parts.join("-");
    },
    formatServerName(info = {}) {
      if (info.name) {
        return info.name;
      }
      if (info.address) {
        return info.address;
      }
      return this.$t("server.name");
    },
    formatOutboundLabel(which = {}) {
      if (!which.outbound) {
        return null;
      }
      const outbound = which.outbound;
      const mapping = this.runningState.outboundToServerName
        ? this.runningState.outboundToServerName[outbound]
        : null;
      if (typeof mapping === "number") {
        return `${outbound.toUpperCase()} - ${this.$t("common.loadBalance")} (${mapping})`;
      }
      return outbound.toUpperCase();
    },
    refreshTableData(touch, running, networkPaused = false) {
      touch.servers.forEach((v) => {
        v.connected = false;
      });
      touch.subscriptions.forEach((s) => {
        s.servers.forEach((v) => {
          v.connected = false;
        });
      });
      // re-point the selection at the replacement row objects so a refresh
      // neither drops the user's selection nor leaves stale rows behind
      // identify a row by what it points at, not by its position: a
      // subscription update reorders nodes and reuses the ids, so an
      // id-based key would move the selection onto a different server
      const keyOf = (data, row) => {
        const sub = data.subscriptions.findIndex((s) => s.servers.includes(row));
        return `${row._type}|${sub}|${row.address}|${row.name}|${row.net}`;
      };
      const selected = this.checkedRows.map((x) => keyOf(this.tableData, x));
      this.tableData = touch;
      const byKey = new Map();
      const remember = (v) => {
        const k = keyOf(touch, v);
        // two identical rows: keep the first, so the selection cannot jump
        if (!byKey.has(k)) byKey.set(k, v);
      };
      touch.servers.forEach(remember);
      touch.subscriptions.forEach((s) => s.servers.forEach(remember));
      this.checkedRows = selected.map((k) => byKey.get(k)).filter(Boolean);
      if (running !== undefined) {
        Object.assign(this.runningState, {
          running: this.getRunningLabel(running, networkPaused),
          networkPaused,
          connectedServer: touch.connectedServer,
        });
      }
    },
    async syncLatestNodeOverview(showError = false) {
      try {
        const res = await this.$axios({
          url: apiRoot + "/touch",
        });
        if (res.data.code === "SUCCESS") {
          this.refreshTableData(res.data.data.touch, res.data.data.running, res.data.data.networkPaused);
          this.updateConnectView();
          return true;
        }
        if (showError) {
          this.$buefy.toast.open({
            message: this.$t("server.refreshFailed", {
              message: backendMessage(this, res) || this.$t("common.fail"),
            }),
            type: "is-warning",
            position: "is-top",
            duration: 5000,
          });
        }
      } catch (err) {
        if (showError) {
          this.$buefy.toast.open({
            message: this.$t("server.refreshFailed", {
              message: err?.response?.data?.message || err?.message || this.$t("common.fail"),
            }),
            type: "is-warning",
            position: "is-top",
            duration: 5000,
          });
        }
      }
      return false;
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
          message: this.$t("import.notImage"),
          type: "is-warning",
          position: "is-top",
        });
        return;
      }
      const reader = new FileReader();
      reader.onload = function (e) {
        // target.result property represents the DataURL of the target object
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
    // notifyStopped is called by the parent (App.vue) when a WebSocket
    // running_state message with running=false is received (e.g. the core
    // crashed or transparent proxy paused because the physical network is down).
    // It immediately updates the local running state so the UI reflects the
    // correct status without waiting for the next /touch poll.
    notifyStopped(networkPaused = false) {
      const next = this.getRunningLabel(false, networkPaused);
      if (this.runningState.running !== next || this.runningState.networkPaused !== networkPaused) {
        Object.assign(this.runningState, {
          running: next,
          networkPaused,
        });
        this.$emit("input", this.runningState);
      }
    },
    notifyRunning(networkPaused = false) {
      const next = this.getRunningLabel(true, networkPaused);
      if (this.runningState.running !== next || this.runningState.networkPaused !== networkPaused) {
        Object.assign(this.runningState, {
          running: next,
          networkPaused,
        });
        this.$emit("input", this.runningState);
      }
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
      if (this.importing) {
        // A second Confirm or Enter while a subscription is still being
        // fetched would import it twice.
        return;
      }
      this.importing = true;
      return this.$axios({
        url: apiRoot + "/import",
        method: "post",
        // the backend allows a subscription fetch 90 s; the 60 s default
        // aborted the request client-side while the import still went through
        timeout: 120000,
        data: {
          url: value || this.importWhat,
          // the batch dialog only takes server links
          kind: this.showModalImportInBatch ? "server" : this.importKind,
        },
      }).then((res) => {
        if (res.data.code === "SUCCESS") {
          this.syncLatestNodeOverview();
          this.$buefy.toast.open({
            message: this.$t("import.success"),
            type: "is-primary",
            position: "is-top",
          });
          this.showModalImport = false;
          this.showModalImportInBatch = false;
          this.importWhat = "";
        } else {
          this.$buefy.toast.open({
            message: this.$t("import.failed", {
              message: backendMessage(this, res) || this.$t("common.fail"),
            }),
            type: "is-warning",
            position: "is-top",
          });
        }
      }).catch((err) => {
        // the interceptor reports every other error itself but re-throws
        // client-side timeouts silently
        if (err && err.code === "ECONNABORTED") {
          this.$buefy.toast.open({
            message: this.$t("import.timeout"),
            type: "is-warning",
            position: "is-top",
          });
        }
      }).finally(() => {
        this.importing = false;
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
          this.checkedRows = [];
          this.syncLatestNodeOverview();
        } else {
          this.$buefy.toast.open({
            message: this.$t("delete.failed", {
              message: backendMessage(this, res) || this.$t("common.fail"),
            }),
            type: "is-warning",
            position: "is-top",
            duration: 5000,
          });
        }
      });
    },
    handleClickDelete() {
      this.$buefy.dialog.confirm({
        title: this.$t("delete.title"),
        message: this.$t("delete.message", { n: this.checkedRows.length }),
        confirmText: this.$t("operations.delete"),
        cancelText: this.$t("operations.cancel"),
        type: "is-danger",
        hasIcon: true,
        icon: "triangle-alert",
        onConfirm: () => this.deleteSelectedServers(),
      });
    },
    handleClickAboutConnection(row, sub) {
      if (!row.connected && this.loadBalanceValid) {
        this.openPickProxyGroup(row, sub);
        return;
      }
      if (!row.connected) {
        this.connectToProxyGroup(row, sub, this.outbound);
        return;
      }
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
            running: this.getRunningLabel(res.data.data.running, res.data.data.networkPaused),
            networkPaused: !!res.data.data.networkPaused,
            connectedServer: res.data.data.touch.connectedServer,
          });
          this.updateConnectView();
          this.syncLatestNodeOverview();
        } else {
          this.$buefy.toast.open({
            message: this.$t("connection.disconnectFailed", {
              message: backendMessage(this, res) || this.$t("common.fail"),
            }),
            type: "is-warning",
            position: "is-top",
            duration: 5000,
          });
        }
      });
    },
    normalizeProxyGroups(groups) {
      const seen = new Set();
      const normalized = [];
      if (groups instanceof Array) {
        for (const group of groups) {
          if (typeof group !== "string") {
            continue;
          }
          const name = group.trim();
          if (!name || seen.has(name)) {
            continue;
          }
          seen.add(name);
          normalized.push(name);
        }
      }
      if (!seen.has("proxy")) {
        normalized.unshift("proxy");
      }
      return normalized;
    },
    getConnectedServersInOutbound(outbound) {
      const connectedServers = this.runningState.connectedServer;
      if (!(connectedServers instanceof Array)) {
        return [];
      }
      return connectedServers.filter(
        (which) => (which.outbound || "proxy") === (outbound || "proxy")
      );
    },
    // 判断某节点是否已连接到指定分组，用于下拉菜单黄色高亮
    isNodeInOutbound(row, sub, outboundName) {
      return this.getConnectedServersInOutbound(outboundName).some((which) => {
        if (sub !== undefined) {
          return (
            which._type === "subscriptionServer" &&
            which.id === row.id &&
            which.sub === sub
          );
        }
        return which._type === "server" && which.id === row.id;
      });
    },
    openPickProxyGroup(row, sub) {
      const groups = this.normalizeProxyGroups(this.outbounds);
      if (groups.length === 1) {
        this.connectToProxyGroup(row, sub, groups[0]);
        return;
      }
      this.$buefy.modal.open({
        parent: this,
        component: ModalPickProxyGroup,
        hasModalCard: true,
        canCancel: true,
        props: {
          groups,
          initialGroup: this.outbound || "proxy",
        },
        events: {
          select: (selectedGroup) => {
            this.connectToProxyGroup(row, sub, selectedGroup);
          },
        },
      });
    },
    connectToProxyGroup(row, sub, outbound) {
      let cancel;
      let loading = this.$buefy.loading.open();
      waitingConnected(
        this.$axios({
          url: apiRoot + "/connection",
          method: "post",
          data: {
            id: row.id,
            _type: row._type,
            sub: sub,
            outbound,
          },
          cancelToken: new axios.CancelToken(function executor(c) {
            cancel = c;
          }),
        }).then((res) => {
          loading.close();
          if (res.data.code === "SUCCESS") {
            Object.assign(this.runningState, {
              running: this.getRunningLabel(res.data.data.running, res.data.data.networkPaused),
              networkPaused: !!res.data.data.networkPaused,
              connectedServer: res.data.data.touch.connectedServer,
            });
            this.$nextTick(() => {
              this.updateConnectView();
            });
            this.syncLatestNodeOverview();
          } else {
            this.$buefy.toast.open({
              message: this.$t("connection.connectFailed", {
                message: backendMessage(this, res) || this.$t("common.fail"),
              }),
              type: "is-warning",
              position: "is-top",
              duration: 5000,
            });
          }
        }).catch((err) => {
          loading.close();
          this.$buefy.toast.open({
            message: this.$t("connection.connectFailed", {
              message: err?.response?.data?.message || err?.message || this.$t("common.fail"),
            }),
            type: "is-warning",
            position: "is-top",
            duration: 5000,
          });
        }),
        3 * 1000,
        cancel
      );
    },
    toggleNodeInGroup(row, sub, group) {
      const targetType = row._type;
      const targetSub = targetType === "subscriptionServer" ? sub : 0;
      const targetId = row.id;
      const currentMembers = this.getConnectedServersInOutbound(group).map((w) => ({
        id: w.id,
        _type: w._type,
        sub: w._type === "subscriptionServer" ? w.sub : 0,
        outbound: group,
      }));

      const sameWhich = (w) => {
        if (w._type !== targetType || w.id !== targetId) {
          return false;
        }
        if (targetType === "subscriptionServer") {
          return w.sub === targetSub;
        }
        return true;
      };

      let nextMembers;
      if (this.isNodeInOutbound(row, sub, group)) {
        nextMembers = currentMembers.filter((w) => !sameWhich(w));
      } else {
        nextMembers = currentMembers.concat([{
          id: targetId,
          _type: targetType,
          sub: targetSub,
          outbound: group,
        }]);
      }

      const loading = this.$buefy.loading.open();
      this.$axios({
        url: apiRoot + "/outboundConnections",
        method: "put",
        data: {
          outbound: group,
          touches: nextMembers,
        },
      }).then((res) => {
        loading.close();
        if (res.data.code === "SUCCESS") {
          Object.assign(this.runningState, {
            running: this.getRunningLabel(res.data.data.running, res.data.data.networkPaused),
            networkPaused: !!res.data.data.networkPaused,
            connectedServer: res.data.data.touch.connectedServer,
          });
          this.$nextTick(() => { this.updateConnectView(); });
          this.syncLatestNodeOverview();
        } else {
          this.$buefy.toast.open({
            message: this.$t("proxyGroup.updateFailed", {
              group,
              message: backendMessage(this, res) || this.$t("common.fail"),
            }),
            type: "is-warning",
            position: "is-top",
            duration: 5000,
          });
        }
      }).catch((err) => {
        loading.close();
        this.$buefy.toast.open({
          message: this.$t("proxyGroup.updateFailed", {
            group,
            message: err?.response?.data?.message || err?.message || this.$t("common.fail"),
          }),
          type: "is-warning",
          position: "is-top",
          duration: 5000,
        });
      });
    },
    handleClickLatency(ping) {
      let touches = JSON.stringify(
        this.checkedRows.map((x) => {
          // iterate through subscriptions
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
      this.checkedRows.forEach((x) => (x.pingLatency = this.$t("latency.testing"))); //refresh
      // this.checkedRows = [];
      let timerTip = setTimeout(() => {
        this.$buefy.toast.open({
          message: this.$t("latency.message"),
          type: "is-primary",
          position: "is-top",
          duration: 5000,
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
                message: this.$t("latency.failed", {
                  message: backendMessage(this, res) || this.$t("common.fail"),
                }),
                type: "is-warning",
                position: "is-top",
                duration: 5000,
              });
              this.checkedRows.forEach((x) => (x.pingLatency = ""));
            }
          );
        })
        .catch(() => {
          // network error: the interceptor already reported it; do not
          // leave the rows on "testing..."
          this.checkedRows.forEach((x) => (x.pingLatency = ""));
        })
        .finally(() => {
          clearTimeout(timerTip);
        });
    },
    // eslint-disable-next-line no-unused-vars
    handleTabsChange(index) {
      // store the index in local storage to remember the tab.
      localStorage.lastNodeTab = index;
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
    isCheckedRowsExportable() {
      return this.checkedRows.length > 0;
    },
    getTouchFromCheckedRow(row) {
      const sub = this.tableData.subscriptions.findIndex((subscription) =>
        subscription.servers.some((server) => server === row)
      );
      return {
        id: row.id,
        _type: row._type,
        sub: sub === -1 ? null : sub,
      };
    },
    async collectSelectedSharingAddresses() {
      const touches = this.checkedRows.map((row) => this.getTouchFromCheckedRow(row));
      const requests = touches.map((touch) =>
        this.$axios({
          url: apiRoot + "/sharingAddress",
          method: "get",
          params: { touch },
        })
      );
      const responses = await Promise.all(requests);
      return responses.map((res) => {
        if (!res?.data || res.data.code !== "SUCCESS") {
          throw new Error(
            (res?.data && backendMessage(this, res)) ||
              this.$t("operations.exportEmpty")
          );
        }
        return res.data.data.sharingAddress || "";
      }).filter((address) => !!address);
    },
    async buildSelectedNodesExportText() {
      const addresses = await this.collectSelectedSharingAddresses();
      if (!addresses.length) {
        throw new Error(this.$t("operations.exportEmpty"));
      }
      return addresses.join("\n");
    },
    async copyTextToClipboard(text) {
      if (navigator?.clipboard?.writeText) {
        await navigator.clipboard.writeText(text);
        return;
      }
      const textarea = document.createElement("textarea");
      textarea.value = text;
      textarea.style.position = "fixed";
      textarea.style.left = "-9999px";
      document.body.appendChild(textarea);
      textarea.select();
      document.execCommand("copy");
      document.body.removeChild(textarea);
    },
    downloadTextFile(text) {
      const timestamp = dayjs().format("YYYYMMDD-HHmmss-SSS");
      const blob = new Blob([text], { type: "text/plain;charset=utf-8" });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `${timestamp}.txt`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    },
    async handleClickExportSelected(mode) {
      if (!this.isCheckedRowsExportable()) {
        return;
      }
      try {
        const exportText = await this.buildSelectedNodesExportText();
        if (mode === "copy") {
          await this.copyTextToClipboard(exportText);
        } else if (mode === "download") {
          this.downloadTextFile(exportText);
        }
        this.$buefy.toast.open({
          message: this.$t(
            mode === "copy" ? "operations.copySelectedDone" : "operations.downloadTxtDone"
          ),
          type: "is-primary",
          position: "is-top",
          duration: 2500,
        });
      } catch (err) {
        this.$buefy.toast.open({
          message: this.$t("operations.exportFailed", {
            message: err?.message || this.$t("common.fail"),
          }),
          type: "is-warning",
          position: "is-top",
          duration: 5000,
        });
      }
    },
    handleClickShare(row, sub) {
      const TYPE_MAP = {
        [CONST.SubscriptionServerType]: this.$t("sharing.serverTitle"),
        [CONST.ServerType]: this.$t("sharing.serverTitle"),
        [CONST.SubscriptionType]: this.$t("sharing.subscriptionTitle"),
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
        }, null, "sharing.failed");
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
          this.syncLatestNodeOverview();
          this.$buefy.toast.open({
            message: this.$t("subscription.updated"),
            type: "is-primary",
            position: "is-top",
            duration: 5000,
          });
        }, null, "subscription.updateFailed");
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
          kind: "server",
          which: this.which,
        },
        timeout: 0,
      }).then((res) => {
        handleResponse(res, this, () => {
          this.$buefy.toast.open({
            message: this.$t("server.saved"),
            type: "is-primary",
            position: "is-top",
            duration: 3000,
          });
          this.showModalServer = false;
          this.syncLatestNodeOverview();
        }, null, "server.saveFailed");
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
            message: this.$t("subscription.saved"),
            type: "is-primary",
            position: "is-top",
            duration: 3000,
          });
          this.showModalSubscription = false;
          this.syncLatestNodeOverview();
        }, null, "subscription.saveFailed");
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

  .lucide {
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
@import "bulma/sass/utilities/all.sass";

#toolbar {
  @media screen and (max-width: 450px) {
    &.float-toolbar {
      top: 4.25rem;
      margin-left: 25px;
      width: calc(100% - 50px);
    }
  }

  // Phones: the two groups stack as full-width rows with one gap value, so
  // the buttons line up instead of wrapping into three ragged rows with
  // Bulma's per-field margins.
  @media screen and (max-width: 768px) {
    flex-direction: column;
    align-items: stretch;
    gap: 0.5rem;
    padding: 0.5rem;

    > .field-body,
    > .field-body > .field.is-grouped {
      flex-direction: column;
      align-items: stretch;
      gap: 0.5rem;
      width: 100%;
    }

    > .field-body > .field.is-grouped > div,
    .right {
      display: flex;
      flex-wrap: wrap;
      gap: 0.5rem;
      max-width: 100% !important;
      margin: 0;
    }

    // .right is the create/import pair: keep it right-aligned like on
    // desktop, on its own row under the selection actions
    .right {
      justify-content: flex-end;
      width: 100%;
    }

    .button,
    .field,
    .dropdown {
      margin: 0 !important;
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
  background: transparent;
  width: 100%;
  border-radius: 3px;
  pointer-events: none;

  // While stuck to the top the table scrolls underneath; the bar needs an
  // opaque surface or the rows show through between and behind the
  // (possibly disabled, half-transparent) buttons. Dark values live in
  // dark-theme.scss.
  &.float-toolbar {
    background: #f5f5f5;
    box-shadow: 0 2px 6px rgba(0, 0, 0, 0.08);
  }
  &.float-toolbar-active {
    background: #ececec;
  }

  * {
    pointer-events: auto;
  }

  // Both groups stay in flow: with the buttons absolutely positioned the
  // toolbar collapsed to its padding when the left group was hidden (no
  // rows selected) and the create/import buttons overlapped whatever came
  // next, e.g. the welcome card on an empty page.
  display: flex;
  justify-content: space-between;
  align-items: flex-start;

  // below the tablet breakpoint Buefy wraps the groups in
  // .field-body > .field.is-grouped, neither of which stretches; keep
  // both full-width flex rows so the right group still ends up right
  > .field-body,
  > .field-body > .field.is-grouped {
    display: flex;
    flex: 1;
    width: 100%;
    justify-content: space-between;
    align-items: flex-start;
  }

  .right {
    margin-left: auto;
    display: flex;
    flex-wrap: wrap;
    justify-content: flex-end;
  }

  transition: all 200ms linear;

  button {
    transition: all 100ms ease-in-out;
  }
}

// a tab whose table holds a connected node; set through header-class
// instead of a hidden 32px icon that showed as a stray glyph on some
// phones
.tabs li.tab-connected a span {
  color: #ff6719;
}

.node-group-option {
  display: block;
  width: 100%;
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
}

.node-group-option--active {
  background-color: #ffe08a;
  color: #5f4b00;
  font-weight: 600;
}

body.theme-dark .node-group-option--active {
  background-color: #6a5318;
  color: #ffe29a;
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

.dialog .icon-triangle-alert {
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

// Toolbar buttons on phones. reset.scss scales the root font to 0.8em
// below 768px, so rem values shrink again; use px for a real touch target
// (14px text, ~36px tall) instead of the old 0.65rem (~8px text).
@media screen and (max-width: 768px) {
  #toolbar .button.mobile-small,
  #toolbar .button.field {
    font-size: 14px;
    height: 2.5em;
    padding-left: 0.9em;
    padding-right: 0.9em;
    border-radius: 4px;
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

// The handle that brings the status sidebar back was a 36px glyph in a
// shadowed box, larger than any control on the page; a phone showed it as the
// biggest thing on screen. Keep it a small, evenly padded target.
.sidebar-handle {
  display: block;
  font-size: 20px;
  line-height: 1;
  padding: 6px;
  cursor: pointer;
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

  // Bulma's small message pads the header and the body differently, which
  // left a gap under the title and the body text a step further in.
  .message.is-small .message-header,
  .message.is-small .message-body {
    padding: 0.55rem 0.75rem;
  }

  .message.is-small .message-body {
    padding-top: 0.45rem;
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

// Three things share the header: the name, the group tag and the
// subscription it came from. The tag used to sit beside the name and squeeze
// it into two lines while staying centred against them, and the subscription
// name wrapped wherever the flex line broke. The name takes the row; a tag
// that does not fit beside it drops to its own line; the subscription name
// always sits under both, in a smaller face.
.node-status-card__header {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  column-gap: 0.5rem;
  row-gap: 0.3rem;
  font-weight: 600;
}

.node-status-card__title {
  flex: 1 1 auto;
  min-width: 0;
  font-size: 0.95rem;
  line-height: 1.3;
  overflow-wrap: anywhere;
}

.node-status-card__subscription {
  flex: 0 0 100%;
  font-size: 0.75rem;
  font-weight: 500;
  opacity: 0.75;
  line-height: 1.2;
}

.node-status-card__group {
  font-size: 0.7rem;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  border-radius: 999px;
  padding: 0.1rem 0.6rem;
  border: 1px solid var(--node-status-group-border, rgba(255, 255, 255, 0.6));
  background-color: var(--node-status-group-bg, rgba(255, 255, 255, 0.15));
  color: var(--node-status-group-color, inherit);
  white-space: nowrap;
  flex: 0 0 auto;
  margin-left: auto;
}

.message.is-light .node-status-card__group {
  border-color: var(--node-status-group-light-border, rgba(0, 0, 0, 0.45));
  background-color: var(--node-status-group-light-bg, rgba(0, 0, 0, 0.05));
}

.node-status-card__body {
  font-size: 0.85rem;
  line-height: 1.4;
}

.node-status-card__body p {
  margin-bottom: 0.15rem;
}

.node-status-card__body p:last-child {
  margin-bottom: 0;
}

tr.highlight-row-connected {
  transition: background-color 0.05s linear;
  background-color: var(--node-highlight-connected, #a8cff0);
}

tr.highlight-row-connected > td {
  transition: background-color 0.05s linear;
  background-color: var(--node-highlight-connected, #a8cff0) !important;
}

tr.highlight-row-disconnected {
  transition: background-color 0.05s linear;
  background-color: var(--node-highlight-disconnected, rgba(255, 69, 58, 0.55));
}

tr.highlight-row-disconnected > td {
  transition: background-color 0.05s linear;
  background-color: var(--node-highlight-disconnected, rgba(255, 69, 58, 0.55)) !important;
}

.click-through {
  pointer-events: none;
}

// The core-version notice is the one persistent banner in the app; keep it
// inside the content column instead of edge to edge, and lay the icon out
// as a column so long text does not wrap under it.
.core-version-error {
  margin: 0.75rem auto 0;
  max-width: 1200px;
  width: calc(100% - 1.5rem);
  display: flex;
  border-radius: 6px;
  font-size: 0.9rem;

  ::v-deep .media,
  ::v-deep .media-content {
    align-items: flex-start;
  }

  &__icon {
    margin-right: 0.5rem;
  }
}
.address-column {
  max-width: 350px !important;
  overflow: hidden !important;
  text-overflow: ellipsis !important;
}

// Buefy's mobile cards (below 769px): one td per line with the label on
// the left. Long host names wrapped over three lines and the checkbox
// took a line of its own, so a card was ~350px tall.
@media screen and (max-width: 768px) {
  .b-table .table tbody tr {
    position: relative;
  }
  .b-table .table tbody td.checkbox-cell {
    position: absolute;
    top: 0.45rem;
    right: 0.5rem;
    width: 2.5rem;
    justify-content: flex-end;
    padding: 0;
    border: 0;
  }
  // the ID line shares the row with the checkbox, which sits at its right end
  .b-table .table tbody td.checkbox-cell + td {
    padding-right: 3.5rem !important;
  }
  .b-table .table tbody td .address-column {
    max-width: 60vw !important;
    white-space: nowrap;
    direction: rtl; // keep the distinctive tail of long host names visible
    text-align: right;
  }
}
.latency-column {
  max-width: 120px !important;
  overflow: hidden !important;
  text-overflow: ellipsis !important;
  white-space: nowrap;
}
.latency-valid {
  color: green;
}

@media screen and (max-width: 1920px) {
  .latency-column {
    max-width: 70px !important;
  }
  .address-column {
    max-width: 150px !important;
  }
}
</style>
