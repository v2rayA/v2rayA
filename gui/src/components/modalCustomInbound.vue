<template>
  <div class="modal-card" style="max-width: 640px; margin: auto">
    <header class="modal-card-head">
      <p class="modal-card-title">{{ $t("customInbound.title") }}</p>
    </header>
    <section class="modal-card-body">
      <!-- Existing custom inbounds list -->
      <b-table
        :data="inbounds"
        :mobile-cards="false"
        bordered
        narrowed
        style="margin-bottom: 1rem"
      >
        <b-table-column v-slot="props" :label="$t('customInbound.tag')" width="120">
          <code>{{ props.row.tag }}</code>
        </b-table-column>
        <b-table-column v-slot="props" :label="$t('customInbound.protocol')" width="70">
          <b-tag :type="props.row.protocol === 'socks' ? 'is-info' : 'is-success'" size="is-small">
            {{ props.row.protocol.toUpperCase() }}
          </b-tag>
        </b-table-column>
        <b-table-column v-slot="props" :label="$t('customInbound.port')" width="70">
          {{ props.row.port }}
        </b-table-column>
        <b-table-column v-slot="props" :label="$t('customInbound.outbound')" width="140">
          <span v-if="props.row.outbound">
            <b-tag size="is-small" type="is-warning">{{ props.row.outbound }}</b-tag>
            <span v-if="props.row.outboundType === 'routingA'" class="is-size-7 has-text-grey"> (RoutingA)</span>
          </span>
          <span v-else class="is-size-7 has-text-grey">—</span>
        </b-table-column>
        <b-table-column v-slot="props" :label="$t('operations.name')" width="60">
          <b-button
            size="is-small"
            type="is-danger"
            icon-left="delete"
            @click="handleDelete(props.row.tag)"
          ></b-button>
        </b-table-column>
        <template #empty>
          <div style="text-align: center; padding: 1rem; color: #888">
            {{ $t("customInbound.empty") }}
          </div>
        </template>
      </b-table>

      <!-- Add new inbound form -->
      <div class="box" style="padding: 0.75rem">
        <p class="is-size-6 has-text-weight-semibold" style="margin-bottom: 0.5rem">
          {{ $t("customInbound.addNew") }}
        </p>
        <b-field grouped group-multiline>
          <b-field :label="$t('customInbound.tag')" expanded label-position="on-border">
            <b-input
              v-model="form.tag"
              :placeholder="$t('customInbound.tagPlaceholder')"
            ></b-input>
          </b-field>
          <b-field :label="$t('customInbound.protocol')" label-position="on-border">
            <b-select v-model="form.protocol">
              <option value="socks">SOCKS</option>
              <option value="http">HTTP</option>
            </b-select>
          </b-field>
          <b-field :label="$t('customInbound.port')" label-position="on-border">
            <b-input
              v-model.number="form.port"
              type="number"
              min="1"
              max="65535"
              style="width: 100px"
              :placeholder="$t('customInbound.portPlaceholder')"
            ></b-input>
          </b-field>
        </b-field>

        <!-- Outbound binding -->
        <b-field grouped group-multiline>
          <b-field :label="$t('customInbound.outbound')" expanded label-position="on-border">
            <b-select v-model="form.outbound" expanded :placeholder="$t('customInbound.outboundPlaceholder')">
              <option
                v-for="ob in outbounds"
                :key="ob"
                :value="ob"
              >{{ ob }}</option>
            </b-select>
          </b-field>
          <b-field :label="$t('customInbound.outboundType')" label-position="on-border">
            <b-select v-model="form.outboundType">
              <option value="direct">{{ $t("customInbound.outboundTypeDirect") }}</option>
              <option value="routingA">{{ $t("customInbound.outboundTypeRoutingA") }}</option>
            </b-select>
          </b-field>
          <b-field label=" " label-position="on-border">
            <b-button type="is-primary" :loading="adding" @click="handleAdd">
              {{ $t("operations.add") }}
            </b-button>
          </b-field>
        </b-field>

        <!-- RoutingA rules editor (shown when outboundType is routingA) -->
        <b-field v-if="form.outboundType === 'routingA'" :label="$t('customInbound.routingARules')" label-position="on-border">
          <b-input
            v-model="form.routingARules"
            type="textarea"
            :placeholder="$t('customInbound.routingARulesPlaceholder')"
            rows="6"
          ></b-input>
        </b-field>

        <b-message type="is-info" size="is-small" class="after-line-dot5">
          {{ $t("customInbound.hint") }}
        </b-message>
      </div>
    </section>
    <footer class="modal-card-foot flex-end">
      <b-button @click="$emit('close')">{{ $t("operations.close") }}</b-button>
    </footer>
  </div>
</template>

<script>
import { handleResponse } from "@/assets/js/utils";
import i18n from "@/plugins/i18n";

export default {
  name: "ModalCustomInbound",
  i18n,
  data: () => ({
    inbounds: [],
    outbounds: [],
    form: {
      tag: "",
      protocol: "socks",
      port: "",
      outbound: "",
      outboundType: "direct",
      routingARules: "",
    },
    adding: false,
  }),
  created() {
    this.fetchInbounds();
    this.fetchOutbounds();
  },
  methods: {
    fetchInbounds() {
      this.$axios({ url: apiRoot + "/customInbound" }).then((res) => {
        if (res.data.code === "SUCCESS") {
          this.inbounds = res.data.data.inbounds || [];
        }
      });
    },
    fetchOutbounds() {
      this.$axios({ url: apiRoot + "/outbound" }).then((res) => {
        if (res.data.code === "SUCCESS") {
          this.outbounds = res.data.data.outbounds || [];
        }
      });
    },
    handleAdd() {
      if (!this.form.tag || !this.form.port) {
        this.$buefy.toast.open({
          message: this.$t("customInbound.fillAll"),
          type: "is-warning",
          position: "is-top",
          queue: false,
        });
        return;
      }
      if (!this.form.outbound) {
        this.$buefy.toast.open({
          message: this.$t("customInbound.outboundRequired"),
          type: "is-warning",
          position: "is-top",
          queue: false,
        });
        return;
      }
      this.adding = true;
      this.$axios({
        url: apiRoot + "/customInbound",
        method: "post",
        data: {
          tag: this.form.tag.trim(),
          protocol: this.form.protocol,
          port: Number(this.form.port),
          outbound: this.form.outbound,
          outboundType: this.form.outboundType,
          routingARules: this.form.outboundType === "routingA" ? this.form.routingARules : "",
        },
      })
        .then((res) => {
          handleResponse(res, this, () => {
            this.inbounds = res.data.data.inbounds || [];
            this.form = { tag: "", protocol: "socks", port: "", outbound: "", outboundType: "direct", routingARules: "" };
          });
        })
        .finally(() => {
          this.adding = false;
        });
    },
    handleDelete(tag) {
      this.$buefy.dialog.confirm({
        message: this.$t("customInbound.deleteConfirm", { tag }),
        type: "is-danger",
        confirmText: this.$t("operations.delete"),
        cancelText: this.$t("operations.cancel"),
        onConfirm: () => {
          this.$axios({
            url: apiRoot + "/customInbound",
            method: "delete",
            data: { tag },
          }).then((res) => {
            handleResponse(res, this, () => {
              this.inbounds = res.data.data.inbounds || [];
            });
          });
        },
      });
    },
  },
};
</script>
