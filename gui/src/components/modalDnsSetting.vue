<template>
  <div class="modal-card dns-setting-modal" style="width: auto; min-width: 680px; max-width: 95vw; margin: auto">
    <header class="modal-card-head">
      <p class="modal-card-title">{{ $t("dns.title") }}</p>
      <a
        class="help-link"
        href="https://www.v2fly.org/config/dns.html"
        target="_blank"
        rel="noopener noreferrer"
        :title="$t('dns.helpTooltip')"
      >
        <b-icon icon=" iconfont icon-help-circle-outline" size="is-small" />
        {{ $t("dns.help") }}
      </a>
    </header>
    <section class="modal-card-body">
      <!-- DNS rules table -->
      <div class="dns-table">
        <!-- Header row -->
        <div class="dns-row dns-header">
          <div class="col-server">{{ $t("dns.colServer") }}</div>
          <div class="col-domains">{{ $t("dns.colDomains") }}</div>
          <div class="col-outbound">{{ $t("dns.colOutbound") }}</div>
          <div class="col-actions"></div>
        </div>

        <!-- Rule rows -->
        <div
          v-for="(rule, index) in rules"
          :key="index"
          class="dns-row dns-data-row"
        >
          <div class="col-server">
            <b-input
              v-model="rule.server"
              size="is-small"
              :placeholder="$t('dns.serverPlaceholder')"
              class="code-font"
            />
          </div>
          <div class="col-domains">
            <b-input
              v-model="rule.domains"
              type="textarea"
              size="is-small"
              :placeholder="$t('dns.domainsPlaceholder')"
              class="code-font dns-domains-input"
              rows="2"
            />
          </div>
          <div class="col-outbound">
            <b-select v-model="rule.outbound" size="is-small" expanded>
              <option value="direct">direct</option>
              <option
                v-for="out in outbounds"
                :key="out"
                :value="out"
              >{{ out }}</option>
            </b-select>
          </div>
          <div class="col-actions">
            <b-button
              size="is-small"
              type="is-danger"
              icon-left=" iconfont icon-delete"
              @click="removeRule(index)"
            />
          </div>
        </div>
      </div>

      <div class="dns-add-row">
        <b-button
          size="is-small"
          type="is-primary"
          @click="addRule"
        >+ {{ $t("dns.addRule") }}</b-button>
        <b-button
          size="is-small"
          @click="resetDefault"
          style="margin-left: 8px"
        >{{ $t("dns.resetDefault") }}</b-button>
      </div>
    </section>
    <footer class="modal-card-foot flex-end">
      <button class="button" @click="$emit('close')">
        {{ $t("operations.cancel") }}
      </button>
      <button class="button is-primary" @click="handleClickSubmit">
        {{ $t("operations.save") }}
      </button>
    </footer>
  </div>
</template>

<script>
import { handleResponse } from "@/assets/js/utils";

const DEFAULT_RULES = [
  { server: "localhost", domains: "geosite:private", outbound: "direct" },
  { server: "223.5.5.5", domains: "geosite:cn", outbound: "direct" },
  { server: "8.8.8.8", domains: "", outbound: "proxy" },
];

export default {
  name: "ModalDnsSetting",
  data: () => ({
    rules: DEFAULT_RULES.map((r) => ({ ...r })),
    outbounds: ["proxy"],
  }),
  created() {
    // Load available outbounds
    this.$axios({ url: apiRoot + "/outbounds" }).then((res) => {
      if (res.data && res.data.data && res.data.data.outbounds) {
        this.outbounds = res.data.data.outbounds;
      }
    });
    // Load current DNS rules
    this.$axios({ url: apiRoot + "/dnsRules" }).then((res) => {
      handleResponse(res, this, () => {
        if (res.data.data && res.data.data.rules && res.data.data.rules.length > 0) {
          this.rules = res.data.data.rules.map((r) => ({
            server: r.server || "",
            domains: r.domains || "",
            outbound: r.outbound || "direct",
          }));
        }
      });
    });
  },
  methods: {
    addRule() {
      this.rules.push({ server: "", domains: "", outbound: "direct" });
    },
    removeRule(index) {
      this.rules.splice(index, 1);
    },
    resetDefault() {
      this.rules = DEFAULT_RULES.map((r) => ({ ...r }));
    },
    handleClickSubmit() {
      const validRules = this.rules.filter((r) => r.server.trim() !== "");
      if (validRules.length === 0) {
        this.$buefy.toast.open({
          message: this.$t("dns.errNoRules"),
          type: "is-danger",
          position: "is-top",
        });
        return;
      }
      this.$axios({
        url: apiRoot + "/dnsRules",
        method: "put",
        data: validRules,
      }).then((res) => {
        handleResponse(res, this, () => {
          this.$emit("close");
        });
      });
    },
  },
};
</script>

<style lang="scss" scoped>
.dns-setting-modal {
  .dns-info-msg {
    margin-bottom: 12px;
  }

  .modal-card-head {
    display: flex;
    align-items: center;
  }

  .help-link {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 13px;
    color: #888;
    margin-left: auto;
    text-decoration: none;
    white-space: nowrap;
    &:hover {
      color: #3273dc;
    }
  }

  .dns-table {
    border: 1px solid #dbdbdb;
    border-radius: 4px;
    overflow: hidden;
  }

  .dns-row {
    display: grid;
    grid-template-columns: 220px 1fr 120px 42px;
    gap: 0;
    align-items: start;
    border-bottom: 1px solid #f0f0f0;

    &:last-child {
      border-bottom: none;
    }

    > div {
      padding: 6px 8px;
    }
  }

  .dns-header {
    background: #f8f8f8;
    font-size: 12px;
    font-weight: 600;
    color: #555;
    align-items: center;

    > div {
      padding: 8px 10px;
    }
  }

  .dns-data-row {
    background: #fff;

    &:hover {
      background: #fafafa;
    }
  }

  .col-actions {
    display: flex;
    align-items: flex-start;
    padding-top: 8px;
    justify-content: center;
  }

  .dns-domains-input ::v-deep textarea {
    min-height: 52px;
    resize: vertical;
    font-family: monospace;
    font-size: 12px;
  }

  .dns-add-row {
    margin-top: 12px;
    display: flex;
    align-items: center;
  }
}
</style>
