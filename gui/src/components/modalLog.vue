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
        <b-input
          ref="logBox"
          readonly="readonly"
          type="textarea"
          rows="20"
          :value="logText"
        ></b-input>
      </b-field>
    </section>
  </div>
</template>
<script>
export default {
  data() {
    return {
      logText: "",
      currentSkip: 0,
      intervalId: 0,
      intervalTime: 5,
      intervalCandidate: [2, 5, 10, 15]
    };
  },
  created() {
    this.$axios({
      url: apiRoot + "/log"
    }).then(this.updateLog);
  },
  mounted() {
    this.intervalId = setInterval(() => {
      this.$axios({
        url: apiRoot + `/log?skip=${this.currentSkip}`
      }).then(this.updateLog);
    }, this.intervalTime * 1000);
  },
  destroyed() {
    clearInterval(this.intervalId);
  },
  methods: {
    updateLog(logs) {
      if (logs.data.length !== 0) {
        this.logText += logs.data;
        this.currentSkip += new Blob([logs.data]).size;
        this.$refs.logBox.$refs.textarea.scrollTop = this.$refs.logBox.$refs.textarea.scrollHeight;
      }
    },
    changeInterval(val) {
      this.intervalTime = val;
      clearInterval(this.intervalId);
      this.intervalId = setInterval(() => {
        this.$axios({
          url: apiRoot + `/log?skip=${this.currentSkip}`
        }).then(this.updateLog);
      }, this.intervalTime * 1000);
    }
  }
};
</script>
