<template>
  <!-- TODO: mobile device compatibility -->
  <div class="modal-card" style="width: 65rem">
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
          <div class="log-footer-label">{{ $tc("log.autoScoll") }}</div>
          <div class="log-footer-control">
            <b-switch v-model="autoScoll" @input="changeScoll" />
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
    };
  },
  computed: {
    filteredItems() {
      if (this.levelFilter === "all") {
        return this.items;
      }
      return this.items.filter((item) => item.level === this.levelFilter);
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
    updateLog(logs) {
      if (logs.data.length && logs.data.length !== 0) {
        const baseIndex = this.items.length;
        const items = logs.data
          .split("\n")
          .map((x, i) => ({
            text: x,
            id: baseIndex + i,
            level: this.detectLevel(x),
          }));
        if (this.endOfLine) {
          this.items = this.items.concat(items);
        } else {
          this.items[this.items.length - 1].text += items[0].text;
          this.items = this.items.concat(items.slice(1));
        }
        this.endOfLine = items[items.length - 1].text.endsWith("\n");
        this.currentSkip += new Blob([logs.data]).size;
        if (this.autoScoll && this.filteredItems.length > 0) {
          this.$refs.logScroller.scrollToItem(this.filteredItems.length - 1);
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
  justify-content: flex-start;
  gap: 1.5rem;
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
</style>

<style lang="scss">
.log-scroller {
  height: 50vh;

  .vue-recycle-scroller__item-wrapper {
    overflow-x: auto;
  }
}
</style>
