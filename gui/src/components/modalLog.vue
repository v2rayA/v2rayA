<template>
  <!-- TODO: mobile device compatibility -->
  <div class="modal-card" style="width:65rem">
    <header class="modal-card-head">
      <p class="modal-card-title">
        {{ $tc("log.logModalTitle") }}
      </p>
    </header>
    <section ref="section" :class="{ 'modal-card-body': true }">
      <b-field :label="$tc('log.refreshInterval')">
        <b-select v-model="intervalTime" @input="changeInterval">
          <option
            v-for="candidate in intervalCandidate"
            :key="candidate"
            :value="candidate"
          >
            {{ `${candidate} ${$tc("log.seconds")}` }}
          </option>
        </b-select>
      </b-field>
      <b-field label="Logs" style="margin-bottom:2em">
        <RecycleScroller class="log-scroller" :items="items" :item-size="28">
          <template v-slot="{ item }">
            <p class="text">{{ item.text }}</p>
          </template>
        </RecycleScroller>
      </b-field>
    </section>
  </div>
</template>
<script>
export default {
  data() {
    return {
      items: [],
      endOfLine: true,
      currentSkip: 0,
      intervalId: 0,
      intervalTime: 5,
      intervalCandidate: [2, 5, 10, 15]
    };
  },
  created() {
    this.$axios({
      url: apiRoot + "/logger"
    }).then(this.updateLog);
  },
  mounted() {
    this.intervalId = setInterval(() => {
      this.$axios({
        url: apiRoot + `/logger`,
        params: { skip: this.currentSkip }
      }).then(this.updateLog);
    }, this.intervalTime * 1000);
  },
  destroyed() {
    clearInterval(this.intervalId);
  },
  methods: {
    updateLog(logs) {
      if (logs.data.length && logs.data.length !== 0) {
        const baseIndex = this.items.length;
        const items = logs.data
          .split("\n")
          .map((x, i) => ({ text: x, id: baseIndex + i }));
        if (this.endOfLine) {
          this.items = this.items.concat(items);
        } else {
          this.items[this.items.length - 1].text += items[0].text;
          this.items = this.items.concat(items.slice(1));
        }
        this.endOfLine = items[items.length - 1].text.endsWith("\n");
        this.currentSkip += new Blob([logs.data]).size;
      }
    },
    changeInterval(val) {
      this.intervalTime = val;
      clearInterval(this.intervalId);
      this.intervalId = setInterval(() => {
        this.$axios({
          url: apiRoot + `/logger`,
          params: { skip: this.currentSkip }
        }).then(this.updateLog);
      }, this.intervalTime * 1000);
    }
  }
};
</script>
<style scoped>
.text {
  font-size: 18px;
  line-height: 28px;
  white-space: nowrap;
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
