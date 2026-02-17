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
        <RecycleScroller
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
                  {{ option.value === 'all' ? $t(option.label) : option.label }}
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
      sourceFilter: "all",
      sourceOptions: [
        { value: "all", label: "log.sources.all" },
      ],
    };
  },
  computed: {
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

    this.$axios({
      url: apiRoot + "/logger",
    }).then(this.updateLog);
  },
  mounted() {
    this.intervalId = setInterval(() => {
      this.$axios({
        url: apiRoot + `/logger`,
        params: { skip: this.currentSkip },
      }).then(this.updateLog);
    }, this.intervalTime * 1000);
  },
  destroyed() {
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
      // 匹配 [xxx.go:123] 或 [xxxService] 格式
      const match = text.match(/\[([^\]]+\.go|[A-Za-z]+Service|[A-Za-z]+\.[A-Za-z]+)(?::\d+)?\]/);
      if (match) {
        return match[1];
      }
      return "other";
    },
    addSourceOption(source) {
      if (source && !this.sourceOptions.find(opt => opt.value === source)) {
        this.sourceOptions.push({
          value: source,
          label: source
        });
      }
    },
    updateLog(logs) {
      if (logs.data.length && logs.data.length !== 0) {
        const baseIndex = this.items.length;
        const items = logs.data
          .split("\n")
          .map((x, i) => {
            const source = this.detectSource(x);
            this.addSourceOption(source);
            return {
              text: x,
              id: baseIndex + i,
              level: this.detectLevel(x),
              source: source,
            };
          });
        if (this.endOfLine) {
          this.items = this.items.concat(items);
        } else {
          this.items[this.items.length - 1].text += items[0].text;
          this.items = this.items.concat(items.slice(1));
        }
        this.endOfLine = items[items.length - 1].text.endsWith("\n");
        this.currentSkip += new Blob([logs.data]).size;
        if (this.autoScoll && this.filteredItems.length > 0) {
          this.$nextTick(() => {
            if (this.$refs.logScroller && this.filteredItems.length > 0) {
              this.$refs.logScroller.scrollToItem(this.filteredItems.length - 1);
            }
          });
        }
      }
    },
    changeInterval(val) {
      this.intervalTime = val;
      clearInterval(this.intervalId);
      this.intervalId = setInterval(() => {
        this.$axios({
          url: apiRoot + `/logger`,
          params: { skip: this.currentSkip },
        }).then(this.updateLog);
      }, this.intervalTime * 1000);
    },
    changeScoll(val) {
      localStorage.setItem("log.autoScoll", val);
    },
  },
};
</script>

<style scoped>
.log-modal {
  width: 65rem;
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

.log-line-number {
  min-width: 3.5rem;
  text-align: right;
  color: #9aa4b2;
  font-size: 14px;
  user-select: none;
}

.log-footer {
  display: flex;
  align-items: flex-start;
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
  align-items: flex-start;
  margin-left: auto;
}

.log-footer-item {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.5rem;
  min-width: 160px;
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
.log-scroller {
  height: 50vh;
  max-height: 600px;

  .vue-recycle-scroller__item-wrapper {
    overflow-x: auto;
  }
}

@media screen and (max-width: 768px) {
  .log-scroller {
    height: 40vh;
    max-height: 400px;
  }
}

@media screen and (max-width: 480px) {
  .log-scroller {
    height: 35vh;
    max-height: 300px;
  }
}
</style>
