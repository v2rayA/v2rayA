<template>
  <div class="modal-card" style="max-width: 640px; margin: auto">
    <header class="modal-card-head">
      <p class="modal-card-title">{{ $t("customInbound.title") }}</p>
      <button type="button" class="delete" aria-label="close" @click="$emit('close')"></button>
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
          <i
            v-if="props.row.username"
            class="lucide icon-lock"
            :title="$t('customInbound.authEnabled')"
          />
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
            icon-left="trash-2"
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
        <div class="inbound-form">
          <b-field :label="$t('customInbound.tag')" label-position="on-border" class="inbound-form__tag">
            <b-input
              v-model="form.tag"
              :placeholder="$t('customInbound.tagPlaceholder')"
            ></b-input>
          </b-field>
          <b-field :label="$t('customInbound.protocol')" label-position="on-border" class="inbound-form__protocol">
            <b-select v-model="form.protocol" expanded>
              <option value="socks">SOCKS</option>
              <option value="http">HTTP</option>
            </b-select>
          </b-field>
          <b-field :label="$t('customInbound.port')" label-position="on-border" class="inbound-form__port">
            <b-input
              v-model.number="form.port"
              type="number"
              min="1"
              max="65535"
              :placeholder="$t('customInbound.portPlaceholder')"
            ></b-input>
          </b-field>
          <b-field :label="$t('customInbound.outbound')" label-position="on-border" class="inbound-form__outbound">
            <b-select v-model="form.outbound" expanded>
              <option
                v-for="ob in outbounds"
                :key="ob"
                :value="ob"
              >{{ ob }}</option>
            </b-select>
          </b-field>
          <b-field :label="$t('customInbound.outboundType')" label-position="on-border" class="inbound-form__mode">
            <b-select v-model="form.outboundType" expanded>
              <option value="direct">{{ $t("customInbound.outboundTypeDirect") }}</option>
              <option value="routingA">{{ $t("customInbound.outboundTypeRoutingA") }}</option>
            </b-select>
          </b-field>
          <b-field :label="$t('customInbound.username')" label-position="on-border" class="inbound-form__user">
            <b-input
              v-model="form.username"
              :placeholder="$t('customInbound.authOptional')"
              autocomplete="off"
            ></b-input>
          </b-field>
          <b-field :label="$t('customInbound.password')" label-position="on-border" class="inbound-form__pass">
            <b-input
              v-model="form.password"
              type="password"
              password-reveal
              :placeholder="$t('customInbound.authOptional')"
              autocomplete="off"
            ></b-input>
          </b-field>
          <div class="inbound-form__add">
            <b-button type="is-primary" expanded :loading="adding" @click="handleAdd">
              {{ $t("operations.add") }}
            </b-button>
          </div>
        </div>

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
  emits: ["close"],
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
      username: "",
      password: "",
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
      this.$axios({ url: apiRoot + "/outbounds" }).then((res) => {
        if (res.data.code === "SUCCESS") {
          this.outbounds = res.data.data.outbounds || [];
          // The select renders blank with no value, which reads as an empty
          // list; every inbound needs a group anyway, so preselect the first.
          if (!this.form.outbound && this.outbounds.length > 0) {
            this.form.outbound = this.outbounds[0];
          }
        }
      });
    },
    handleAdd() {
      if (!this.form.tag || !this.form.port) {
        this.$buefy.toast.open({
          message: this.$t("customInbound.fillAll"),
          type: "is-warning",
          position: "is-top",
        });
        return;
      }
      if (!this.form.outbound) {
        this.$buefy.toast.open({
          message: this.$t("customInbound.outboundRequired"),
          type: "is-warning",
          position: "is-top",
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
          username: this.form.username.trim(),
          password: this.form.password,
        },
      })
        .then((res) => {
          handleResponse(res, this, () => {
            this.inbounds = res.data.data.inbounds || [];
            this.form = {
              tag: "",
              protocol: "socks",
              port: "",
              outbound: this.outbounds[0] || "",
              outboundType: "direct",
              routingARules: "",
              username: "",
              password: "",
            };
          }, null, "customInbound.saveFailed");
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
            }, null, "customInbound.deleteFailed");
          });
        },
      });
    },
  },
};
</script>

<style lang="scss" scoped>
// A grid keeps the labels on their own borders and the columns aligned; the
// previous grouped fields put the outbound select on a row of its own with the
// Add button wedged against it.
.inbound-form {
  display: grid;
  grid-template-columns: repeat(12, 1fr);
  gap: 1.25rem 0.75rem;
  align-items: end;

  .field {
    margin-bottom: 0;
  }
}

.inbound-form__tag {
  grid-column: span 6;
}

.inbound-form__protocol,
.inbound-form__port {
  grid-column: span 3;
}

.inbound-form__outbound,
.inbound-form__mode {
  grid-column: span 6;
}

.inbound-form__user,
.inbound-form__pass,
.inbound-form__add {
  grid-column: span 4;
}

@media screen and (max-width: 640px) {
  .inbound-form {
    grid-template-columns: repeat(2, 1fr);
  }

  .inbound-form__tag,
  .inbound-form__outbound,
  .inbound-form__mode,
  .inbound-form__add {
    grid-column: span 2;
  }

  .inbound-form__protocol,
  .inbound-form__port,
  .inbound-form__user,
  .inbound-form__pass {
    grid-column: span 1;
  }

  .inbound-form__add {
    grid-column: span 2;
  }
}
</style>
