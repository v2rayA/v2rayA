<template>
  <div class="modal-card log-modal">
    <header class="modal-card-head">
      <p class="modal-card-title">
        {{ $tc("log.logModalTitle") }}
      </p>
    </header>
    <section ref="section" :class="{ 'modal-card-body': true }">
      <div class="log-title">{{ $t("log.logsLabel") }}</div>
      <div class="log-content" tabindex="0" @keydown="handleLogKeydown">
        <!-- A virtual scroller needs one fixed row height, so a phone would
             only ever show the left edge of each line. There the tail is
             rendered plainly instead, with the lines wrapped. -->
        <div
          v-if="narrow"
          ref="narrowLog"
          class="log-scroller log-scroller--narrow"
        >
          <div v-if="tailSkipped > 0" class="log-tail-note">
            {{ $t("log.tailOnly", { count: tailLimit, skipped: tailSkipped }) }}
          </div>
          <div
            v-for="(item, index) in narrowItems"
            :key="item.id || index"
            class="log-row log-row--wrap"
          >
            <span class="log-line-number">{{
              filteredItems.length - narrowItems.length + index + 1
            }}</span>
            <hightlight-log class="text" :text="item.text"></hightlight-log>
          </div>
        </div>
        <RecycleScroller
          v-else
          ref="logScroller"
          v-slot="{ item, index }"
          class="log-scroller"
          :items="filteredItems"
          :item-size="itemSize"
          :grid-items="1"
          :buffer="1000"
        >
          <div class="log-row">
            <span class="log-line-number">{{ index + 1 }}</span>
            <hightlight-log class="text" :text="item.text"></hightlight-log>
          </div>
        </RecycleScroller>
      </div>
      <div class="log-footer">
        <div class="log-footer-left">
          <div class="log-footer-item">
            <div class="log-footer-label">{{ $tc("log.refreshInterval") }}</div>
            <div class="log-footer-control">
              <b-select v-model="intervalTime" @input="changeInterval">
                <option
                  v-for="candidate in intervalCandidate"
                  :key="candidate"
                  :value="candidate"
                >
                  {{ `${candidate} ${$tc("log.seconds")}` }}
                </option>
              </b-select>
            </div>
          </div>
          <div class="log-footer-item">
            <div class="log-footer-label">{{ $tc("log.category") }}</div>
            <div class="log-footer-control">
              <b-select v-model="levelFilter">
                <option
                  v-for="option in levelOptions"
                  :key="option.value"
                  :value="option.value"
                >
                  {{ $t(option.label) }}
                </option>
              </b-select>
            </div>
          </div>
          <div class="log-footer-item">
            <div class="log-footer-label">{{ $tc("log.source") }}</div>
            <div class="log-footer-control">
              <b-select v-model="sourceFilter">
                <option
                  v-for="option in sourceOptions"
                  :key="option.value"
                  :value="option.value"
                >
                  {{ option.value === "all" ? $t(option.label) : option.label }}
                </option>
              </b-select>
            </div>
          </div>
        </div>
        <div class="log-footer-right">
          <div class="log-footer-item">
            <div class="log-footer-label">{{ $tc("log.autoShowNew") }}</div>
            <div class="log-footer-control">
              <b-switch v-model="autoScoll" @input="changeScoll" />
            </div>
          </div>
          <div class="log-footer-item">
            <div class="log-footer-label" aria-hidden="true">&nbsp;</div>
            <div class="log-footer-control">
              <b-button
                icon-left="download"
                :disabled="filteredItems.length === 0"
                @click="handleExport"
              >
                {{ $t("log.export") }}
              </b-button>
            </div>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script>
import HightlightLog from "@/components/highlightLog";
export default {
  components: { HightlightLog },
  data() {
    return {
      items: [],
      endOfLine: true,
      currentSkip: 0,
      fetching: false,
      intervalId: 0,
      intervalTime: 5,
      intervalCandidate: [2, 5, 10, 15],
      itemSize: 28,
      autoScoll: true,
      levelFilter: "all",
      levelOptions: [
        { value: "all", label: "log.categories.all" },
        { value: "error", label: "log.categories.error" },
        { value: "warn", label: "log.categories.warn" },
        { value: "info", label: "log.categories.info" },
        { value: "debug", label: "log.categories.debug" },
        { value: "trace", label: "log.categories.trace" },
        { value: "other", label: "log.categories.other" },
      ],
      narrow: false,
      tailLimit: 300,
      sourceFilter: "all",
      sourceOptions: [{ value: "all", label: "log.sources.all" }],
    };
  },
  computed: {
    narrowItems() {
      const items = this.filteredItems;
      return items.length > this.tailLimit
        ? items.slice(-this.tailLimit)
        : items;
    },
    tailSkipped() {
      return Math.max(0, this.filteredItems.length - this.tailLimit);
    },
    filteredItems() {
      let filtered = this.items;

      if (this.levelFilter !== "all") {
        filtered = filtered.filter((item) => item.level === this.levelFilter);
      }

      if (this.sourceFilter !== "all") {
        filtered = filtered.filter((item) => item.source === this.sourceFilter);
      }

      return filtered;
    },
  },
  created() {
    this.autoScoll = !(localStorage.getItem("log.autoScoll") === "false");
    this.narrowQuery = window.matchMedia("(max-width: 768px)");
    this.narrow = this.narrowQuery.matches;
    this.onNarrowChange = (e) => {
      this.narrow = e.matches;
    };
    if (this.narrowQuery.addEventListener) {
      this.narrowQuery.addEventListener("change", this.onNarrowChange);
    } else {
      this.narrowQuery.addListener(this.onNarrowChange);
    }

    this.fetchLog();
  },
  mounted() {
    this.intervalId = setInterval(() => {
      this.fetchLog();
    }, this.intervalTime * 1000);
  },
  destroyed() {
    if (this.narrowQuery) {
      if (this.narrowQuery.removeEventListener) {
        this.narrowQuery.removeEventListener("change", this.onNarrowChange);
      } else {
        this.narrowQuery.removeListener(this.onNarrowChange);
      }
    }
    clearInterval(this.intervalId);
  },
  methods: {
    handleLogKeydown(event) {
      const scroller = this.$refs.logScroller;
      const el = scroller && scroller.$el ? scroller.$el : null;
      if (!el) {
        return;
      }
      const pageStep = Math.max(1, Math.floor(el.clientHeight * 0.9));
      switch (event.key) {
        case "Home":
          el.scrollTop = 0;
          event.preventDefault();
          break;
        case "End":
          el.scrollTop = el.scrollHeight;
          event.preventDefault();
          break;
        case "PageUp":
          el.scrollTop = Math.max(0, el.scrollTop - pageStep);
          event.preventDefault();
          break;
        case "PageDown":
          el.scrollTop = Math.min(el.scrollHeight, el.scrollTop + pageStep);
          event.preventDefault();
          break;
        default:
          break;
      }
    },
    detectLevel(text) {
      const lower = text.toLowerCase();
      if (lower.includes("[e]") || lower.includes(" error ")) {
        return "error";
      }
      if (lower.includes("[w]") || lower.includes(" warn")) {
        return "warn";
      }
      if (lower.includes("[d]") || lower.includes(" debug")) {
        return "debug";
      }
      if (lower.includes("[t]") || lower.includes(" trace")) {
        return "trace";
      }
      if (lower.includes("[i]") || lower.includes(" info")) {
        return "info";
      }
      return "other";
    },
    detectSource(text) {
      // TinyTun logs are prefixed with [tinytun] by the v2rayA log system.
      if (text.includes("[tinytun]")) {
        return "tinytun";
      }
      // 匹配 [xxx.go:123] 或 [xxxService] 格式
      const match = text.match(
        /\[([^\]]+\.go|[A-Za-z]+Service|[A-Za-z]+\.[A-Za-z]+)(?::\d+)?\]/,
      );
      if (match) {
        return match[1];
      }
      return "other";
    },
    addSourceOption(source) {
      if (source && !this.sourceOptions.find((opt) => opt.value === source)) {
        this.sourceOptions.push({
          value: source,
          label: source,
        });
      }
    },
    fetchLog() {
      if (this.fetching) {
        return;
      }
      this.fetching = true;
      return this.$axios({
        url: apiRoot + "/logger",
        params: { skip: this.currentSkip },
      })
        .then(this.updateLog)
        .finally(() => {
          this.fetching = false;
        });
    },
    updateLog(logs) {
      if (logs.data.length && logs.data.length !== 0) {
        const lines = logs.data.split("\n");
        const endsWithNewline = logs.data.endsWith("\n");
        if (endsWithNewline) {
          lines.pop();
        }
        if (!this.endOfLine && this.items.length) {
          const lastItem = this.items[this.items.length - 1];
          lastItem.text += lines.shift() || "";
          lastItem.level = this.detectLevel(lastItem.text);
          lastItem.source = this.detectSource(lastItem.text);
          this.addSourceOption(lastItem.source);
        }
        const baseIndex = this.items.length;
        const items = lines.map((x, i) => {
          const source = this.detectSource(x);
          this.addSourceOption(source);
          return {
            text: x,
            id: baseIndex + i,
            level: this.detectLevel(x),
            source,
          };
        });
        this.items = this.items.concat(items);
        this.endOfLine = endsWithNewline;
        this.currentSkip += new Blob([logs.data]).size;
        if (this.autoScoll && this.narrow) {
          this.$nextTick(() => {
            const el = this.$refs.narrowLog;
            if (el) {
              el.scrollTop = el.scrollHeight;
            }
          });
        }
        if (this.autoScoll && !this.narrow && this.filteredItems.length > 0) {
          this.$nextTick(() => {
            if (this.$refs.logScroller && this.filteredItems.length > 0) {
              this.$refs.logScroller.scrollToItem(
                this.filteredItems.length - 1,
              );
            }
          });
        }
      }
    },
    changeInterval(val) {
      this.intervalTime = val;
      clearInterval(this.intervalId);
      this.intervalId = setInterval(() => {
        this.fetchLog();
      }, this.intervalTime * 1000);
    },
    handleExport() {
      // Export what the dialog shows, filters included, so the file matches
      // what the user was looking at.
      const text = this.filteredItems.map((item) => item.text).join("\n");
      const stamp = new Date().toISOString().replace(/[:T]/g, "-").slice(0, 19);
      const blob = new Blob([text + "\n"], {
        type: "text/plain;charset=utf-8",
      });
      const url = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = url;
      a.download = `v2raya-log-${stamp}.txt`;
      document.body.appendChild(a);
      a.click();
      a.remove();
      setTimeout(() => URL.revokeObjectURL(url), 1000);
    },
    changeScoll(val) {
      localStorage.setItem("log.autoScoll", val);
    },
  },
};
</script>

<style scoped>
/* 60rem is 960px at the 16px root font and never wider than the wrapper
   Buefy sizes (max-width 960px). At 65rem, on any viewport wider than 1400px
   (a 200% display scaling or a zoomed-out window reaches that), the card ran
   80px past the wrapper and the close button, which sits on the wrapper
   corner, ended up in the middle of the header. */
.log-modal {
  width: 60rem;
  max-width: 95vw;
}

.text {
  font-size: 16px;
  line-height: 30px;
  white-space: nowrap;
}

.log-title {
  font-weight: 600;
  margin-bottom: 0.75rem;
}

.log-content {
  margin-bottom: 1.5rem;
}

.log-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

/* On a phone the line has to wrap: the message is the part worth reading and
   it used to be clipped, leaving a column of timestamps. */
.log-row--wrap {
  align-items: flex-start;
}

.log-row--wrap .text {
  white-space: pre-wrap;
  word-break: break-word;
  flex: 1;
  min-width: 0;
}

.log-row--wrap .log-line-number {
  padding-top: 2px;
}

.log-tail-note {
  font-size: 12px;
  color: #9aa4b2;
  padding: 0 0 0.5rem;
}

.log-line-number {
  min-width: 3.5rem;
  text-align: right;
  color: #9aa4b2;
  font-size: 14px;
  user-select: none;
}

.log-footer {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 1.5rem;
}

.log-footer-left {
  display: flex;
  align-items: flex-start;
  gap: 1.5rem;
  flex: 1;
  flex-wrap: wrap;
}

.log-footer-right {
  display: flex;
  align-items: flex-end;
  gap: 1.5rem;
  margin-left: auto;
}

.log-footer-right .log-footer-item {
  min-width: 0;
}

.log-footer-item {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: flex-end;
  gap: 0.5rem;
  min-width: 160px;
}

/* Every control sits on one baseline: a select is 2.5em tall, a switch and a
   button are not, so the row is given the height and centres what it holds. */
.log-footer-control {
  display: flex;
  align-items: center;
  min-height: 2.5em;
}

.log-footer-label {
  font-weight: 600;
  min-height: 1.5rem;
  font-size: 14px;
}

.log-footer-control ::v-deep .control,
.log-footer-control ::v-deep .select,
.log-footer-control ::v-deep select {
  width: 100%;
}

.log-footer-control ::v-deep .control {
  min-height: 40px;
  display: flex;
  align-items: center;
}

/* 移动端适配 */
@media screen and (max-width: 768px) {
  .log-modal {
    width: 100%;
    max-width: 100%;
    margin: 0;
  }

  .log-modal ::v-deep .modal-card-head {
    padding: 1rem;
  }

  .log-modal ::v-deep .modal-card-body {
    padding: 1rem;
  }

  .log-title {
    font-size: 14px;
    margin-bottom: 0.5rem;
  }

  .log-content {
    margin-bottom: 1rem;
  }

  .text {
    font-size: 12px;
    line-height: 22px;
  }

  .log-line-number {
    min-width: 2.5rem;
    font-size: 11px;
  }

  .log-row {
    gap: 8px;
  }

  .log-footer {
    flex-direction: column;
    gap: 1rem;
  }

  .log-footer-left {
    width: 100%;
    flex-direction: column;
    gap: 1rem;
  }

  .log-footer-right {
    width: 100%;
    margin-left: 0;
  }

  .log-footer-item {
    width: 100%;
    min-width: 100%;
    gap: 0.4rem;
  }

  .log-footer-label {
    font-size: 13px;
    min-height: auto;
  }

  .log-footer-control ::v-deep .control {
    min-height: 36px;
  }

  .log-footer-control ::v-deep select,
  .log-footer-control ::v-deep .select select {
    font-size: 14px;
  }
}

@media screen and (max-width: 480px) {
  .log-modal ::v-deep .modal-card-head {
    padding: 0.75rem;
  }

  .log-modal ::v-deep .modal-card-body {
    padding: 0.75rem;
  }

  .log-modal ::v-deep .modal-card-title {
    font-size: 16px;
  }

  .text {
    font-size: 11px;
    line-height: 20px;
  }

  .log-line-number {
    min-width: 2rem;
    font-size: 10px;
  }
}
</style>

<style lang="scss">
.log-scroller--narrow {
  overflow-y: auto;
  overflow-x: hidden;
}

// The log area gives way to the footer: at 50vh a short window pushed the
// filters and the export button below the fold, and only a scrollbar hinted
// that they existed. The subtraction is the dialog chrome plus the footer.
.log-scroller {
  height: clamp(140px, calc(100vh - 360px), 600px);

  .vue-recycle-scroller__item-wrapper {
    overflow-x: auto;
  }
}

@media screen and (max-width: 768px) {
  .log-scroller {
    height: clamp(140px, calc(100vh - 420px), 400px);
  }
}

@media screen and (max-width: 480px) {
  .log-scroller {
    height: clamp(120px, calc(100vh - 480px), 300px);
  }
}

body.theme-dark {
  .log-modal {
    .log-title {
      color: var(--md-on-surface-variant);
    }

    .log-footer-label {
      color: var(--md-on-surface-variant);
    }

    .log-row .log-line-number {
      color: var(--md-surface-variant);
    }

    /* highlight.js github亮色主题覆盖 */
    .hljs {
      background: transparent !important;
      color: #c9d1d9 !important;
    }

    .hljs-number {
      color: #79c0ff !important;
    }

    .hljs-string {
      color: #a5d6ff !important;
    }

    .hljs-keyword,
    .hljs-selector-tag,
    .hljs-meta {
      color: #7ee787 !important;
    }

    .hljs-attr,
    .hljs-attribute {
      color: #ffa657 !important;
    }

    .hljs-comment,
    .hljs-quote {
      color: #8b949e !important;
    }

    .hljs-name,
    .hljs-tag {
      color: #7ee787 !important;
    }

    .hljs-title,
    .hljs-section {
      color: #d2a8ff !important;
    }

    .hljs-literal,
    .hljs-type,
    .hljs-built_in {
      color: #79c0ff !important;
    }

    .hljs-bullet,
    .hljs-symbol,
    .hljs-variable {
      color: #ffa657 !important;
    }

    .hljs-emphasis {
      color: #c9d1d9 !important;
      font-style: italic;
    }

    .hljs-strong {
      color: #c9d1d9 !important;
      font-weight: bold;
    }

    .hljs-link {
      color: #a5d6ff !important;
      text-decoration: underline;
    }

    .hljs-deletion {
      color: #ffa198 !important;
    }

    .hljs-addition {
      color: #7ee787 !important;
    }
  }
}
</style>
