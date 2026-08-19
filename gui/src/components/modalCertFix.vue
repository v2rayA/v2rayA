<template>
  <div class="modal-card" style="max-width: 560px; margin: auto">
    <header class="modal-card-head">
      <p class="modal-card-title">
        {{ phase === "select" ? $t("certfix.detectTitle") : $t("certfix.fixTitle") }}
      </p>
    </header>
    <section class="modal-card-body">
      <template v-if="phase === 'select'">
        <p class="block">
          {{ $t("certfix.description") }}
        </p>
        <div
          v-if="candidates.length === 0"
          class="has-text-centered has-text-grey"
        >
          {{ $t("certfix.noCandidates") }}
        </div>
        <div v-else class="certfix-candidate-list">
          <div
            v-for="c in candidates"
            :key="candidateKey(c)"
            class="certfix-candidate"
          >
            <b-checkbox v-model="selected" :native-value="c">
              <span class="certfix-candidate-name">{{ c.name }}</span>
              <span class="certfix-candidate-reason tag is-light is-small">
                {{ c.reason }}
              </span>
            </b-checkbox>
          </div>
        </div>
      </template>

      <template v-else>
        <b-progress
          :value="progressPercent"
          type="is-primary"
          show-value
          size="is-medium"
        >
          {{ processed }}/{{ total }}
        </b-progress>

        <div class="level is-mobile certfix-stats">
          <div class="level-item has-text-centered">
            <div>
              <p class="heading">{{ $t("certfix.pinned") }}</p>
              <p class="title is-5">{{ pinnedCount }}</p>
            </div>
          </div>
          <div class="level-item has-text-centered">
            <div>
              <p class="heading">{{ $t("certfix.trusted") }}</p>
              <p class="title is-5">{{ trustedCount }}</p>
            </div>
          </div>
          <div class="level-item has-text-centered">
            <div>
              <p class="heading">{{ $t("certfix.failed") }}</p>
              <p class="title is-5 has-text-danger">{{ failedCount }}</p>
            </div>
          </div>
        </div>

        <div v-if="perNode.length" class="certfix-node-list">
          <div
            v-for="n in perNode"
            :key="candidateKey(n)"
            class="certfix-node"
          >
            <span class="certfix-node-name">{{ n.name }}</span>
            <b-tag :type="statusType(n.status)" size="is-small">
              {{ statusLabel(n.status) }}
            </b-tag>
            <span
              v-if="n.error"
              class="certfix-node-error has-text-danger is-size-7"
              :title="n.error"
            >
              {{ n.error }}
            </span>
          </div>
        </div>

        <div v-if="phase === 'summary'" class="certfix-summary has-text-centered">
          <p>{{ summaryText }}</p>
        </div>

        <div class="certfix-logs">
          <div
            v-for="(line, i) in logsTail"
            :key="i"
            class="certfix-log-line is-size-7"
          >
            {{ line }}
          </div>
        </div>
      </template>
    </section>

    <footer class="modal-card-foot" style="justify-content: flex-end">
      <button
        v-if="phase === 'select'"
        class="button"
        type="button"
        @click="close"
      >
        {{ $t("operations.cancel") }}
      </button>
      <button
        v-if="phase === 'select'"
        class="button is-primary"
        type="button"
        :disabled="selected.length === 0 || starting"
        @click="handleProcess"
      >
        {{ $t("certfix.process") }}
      </button>
      <button
        v-if="phase === 'processing'"
        class="button"
        type="button"
        :disabled="cancelling"
        @click="handleCancelJob"
      >
        {{ $t("certfix.cancel") }}
      </button>
      <button
        v-if="phase === 'summary' && failedCount > 0"
        class="button is-primary"
        type="button"
        @click="handleRetryFailed"
      >
        {{ $t("certfix.retry") }}
      </button>
      <button
        v-if="phase === 'summary'"
        class="button"
        type="button"
        @click="close"
      >
        {{ $t("operations.confirm") }}
      </button>
    </footer>
  </div>
</template>

<script>
const TERMINAL_STATUSES = ["completed", "completed_with_failures", "cancelled"];

export default {
  name: "ModalCertFix",
  props: {
    candidates: {
      type: Array,
      default() {
        return [];
      },
    },
  },
  data() {
    return {
      phase: "select",
      selected: [],
      jobId: null,
      status: "",
      total: 0,
      processed: 0,
      succeeded: 0,
      failed: 0,
      results: [],
      logs: [],
      pollId: null,
      starting: false,
      cancelling: false,
    };
  },
  computed: {
    progressPercent() {
      if (!this.total) {
        return 0;
      }
      return Math.round((this.processed / this.total) * 100);
    },
    pinnedCount() {
      return this.results.filter((r) => r.status === "pinned").length;
    },
    trustedCount() {
      return this.results.filter((r) => r.status === "trusted").length;
    },
    failedCount() {
      return this.results.filter((r) => r.status === "failed").length;
    },
    perNode() {
      return this.results;
    },
    logsTail() {
      return this.logs.slice(-80);
    },
    summaryText() {
      const key = this.failedCount > 0 ? "certfix.summaryFailed" : "certfix.summarySuccess";
      return this.$t(key, {
        processed: this.processed,
        total: this.total,
        pinned: this.pinnedCount,
        trusted: this.trustedCount,
        failed: this.failedCount,
      });
    },
  },
  created() {
    this.selected = this.candidates.slice();
  },
  mounted() {
    this.$root.$on("certfix-message", this.handleCertFixMessage);
  },
  beforeDestroy() {
    this.stopPolling();
    this.$root.$off("certfix-message", this.handleCertFixMessage);
  },
  methods: {
    candidateKey(c) {
      const w = c.which || {};
      return `${w._type || "unknown"}-${w.id || 0}-${w.sub ?? "na"}`;
    },
    statusType(status) {
      switch (status) {
        case "trusted":
          return "is-success";
        case "pinned":
          return "is-info";
        case "failed":
          return "is-danger";
        case "skipped":
          return "is-warning";
        default:
          return "is-light";
      }
    },
    statusLabel(status) {
      const key = `certfix.status.${status}`;
      const translated = this.$t(key);
      // vue-i18n returns the key itself when missing, which is acceptable
      return translated === key ? status : translated;
    },
    handleProcess() {
      if (this.selected.length === 0) {
        return;
      }
      this.starting = true;
      this.$axios({
        url: apiRoot + "/certfix",
        method: "post",
        data: {
          candidates: this.selected,
        },
      })
        .then((res) => {
          if (res.data?.code === "SUCCESS") {
            const data = res.data.data || {};
            this.jobId = data.jobId;
            this.status = data.status || "running";
            this.total = this.selected.length;
            this.processed = 0;
            this.results = [];
            this.logs = [];
            this.phase = "processing";
            this.fetchJob();
            this.startPolling();
          } else {
            this.$buefy.toast.open({
              message: res.data?.message || this.$t("common.fail"),
              type: "is-warning",
              position: "is-top",
              queue: false,
              duration: 5000,
            });
          }
        })
        .catch((err) => {
          this.$buefy.toast.open({
            message: err?.response?.data?.message || err?.message || this.$t("common.fail"),
            type: "is-warning",
            position: "is-top",
            queue: false,
            duration: 5000,
          });
        })
        .finally(() => {
          this.starting = false;
        });
    },
    startPolling() {
      this.stopPolling();
      this.pollId = setInterval(() => this.fetchJob(), 1500);
    },
    stopPolling() {
      if (this.pollId) {
        clearInterval(this.pollId);
        this.pollId = null;
      }
    },
    fetchJob() {
      if (!this.jobId) {
        return;
      }
      this.$axios({
        url: apiRoot + `/certfix/${this.jobId}`,
        method: "get",
      })
        .then((res) => {
          if (res.data?.code === "SUCCESS") {
            this.updateFromJob(res.data.data);
          }
        })
        .catch((err) => {
          console.error("certfix poll failed", err);
        });
    },
    updateFromJob(data) {
      if (!data) {
        return;
      }
      this.status = data.status || this.status;
      this.total = data.total ?? this.total;
      this.processed = data.processed ?? this.processed;
      this.succeeded = data.succeeded ?? this.succeeded;
      this.failed = data.failed ?? this.failed;
      if (data.results) {
        this.results = data.results;
      }
      if (data.logs) {
        this.logs = data.logs;
      }
      if (TERMINAL_STATUSES.includes(this.status)) {
        this.stopPolling();
        this.phase = "summary";
      }
    },
    handleCertFixMessage(body) {
      if (!body || body.jobId !== this.jobId) {
        return;
      }
      this.updateFromJob(body);
    },
    handleCancelJob() {
      this.cancelling = true;
      this.stopPolling();
      if (this.jobId) {
        this.$axios({
          url: apiRoot + `/certfix/${this.jobId}`,
          method: "delete",
          data: { id: this.jobId },
        })
          .then(() => {
            this.fetchJob();
          })
          .catch((err) => {
            console.error("cancel certfix failed", err);
          })
          .finally(() => {
            this.cancelling = false;
          });
      } else {
        this.cancelling = false;
        this.close();
      }
    },
    handleRetryFailed() {
      const failed = this.results
        .filter((r) => r.status === "failed")
        .map((r) => ({
          which: r.which,
          name: r.name,
          reason: r.error || "failed",
        }));
      if (failed.length === 0) {
        return;
      }
      this.selected = failed;
      this.handleProcess();
    },
    close() {
      this.stopPolling();
      this.$emit("close");
    },
  },
};
</script>

<style lang="scss" scoped>
.certfix-candidate-list,
.certfix-node-list {
  max-height: 280px;
  overflow-y: auto;
  border: 1px solid rgba(0, 0, 0, 0.1);
  border-radius: 4px;
  padding: 0.5rem;
  margin-bottom: 1rem;
}

.certfix-candidate,
.certfix-node {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.35rem 0;
  border-bottom: 1px solid rgba(0, 0, 0, 0.05);

  &:last-child {
    border-bottom: none;
  }
}

.certfix-candidate-name,
.certfix-node-name {
  margin-left: 0.25rem;
  margin-right: 0.5rem;
  word-break: break-all;
}

.certfix-candidate-reason {
  margin-left: 0.5rem;
}

.certfix-node-error {
  margin-left: 0.5rem;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.certfix-stats {
  margin-bottom: 1rem;
}

.certfix-logs {
  max-height: 200px;
  overflow-y: auto;
  background: rgba(0, 0, 0, 0.03);
  border-radius: 4px;
  padding: 0.5rem;
  font-family: monospace;
  white-space: pre-wrap;
}

.certfix-log-line {
  line-height: 1.4;
}

.certfix-summary {
  margin: 1rem 0;
}

body.theme-dark {
  .certfix-candidate-list,
  .certfix-node-list {
    border-color: rgba(255, 255, 255, 0.15);
  }

  .certfix-candidate,
  .certfix-node {
    border-color: rgba(255, 255, 255, 0.08);
  }

  .certfix-logs {
    background: rgba(255, 255, 255, 0.05);
  }
}
</style>
