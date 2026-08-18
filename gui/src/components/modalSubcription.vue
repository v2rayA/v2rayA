<template>
  <div class="modal-card" style="max-width: 400px; margin: auto">
    <header class="modal-card-head">
      <p class="modal-card-title">{{ $t("configureSubscription.title") }}</p>
    </header>
    <section class="modal-card-body">
      <b-field label="SUBSCRIPTION">
        <b-input
          v-model="which.address"
          type="textarea"
          :placeholder="$t('subscription.subscription')"
        />
      </b-field>
      <b-field label="REMARKS">
        <b-input
          v-model="which.remarks"
          :placeholder="$t('subscription.remarks')"
        />
      </b-field>
      <b-field label="AUTO-SELECT">
        <b-checkbox
	  v-model="which.autoSelect"
	  >{{ $t("subscription.autoSelect") }}
	</b-checkbox>
      </b-field>
      <b-field :label="$t('setting.autoUpdateSub')">
        <b-select v-model="which.autoUpdateMode" expanded>
          <option value="none">{{ $t("setting.options.off") }}</option>
          <option value="auto_update">
            {{ $t("setting.options.updateSubWhenStart") }}
          </option>
          <option value="auto_update_at_intervals">
            {{ $t("setting.options.updateSubAtIntervals") }}
          </option>
        </b-select>
      </b-field>
      <b-field
        v-if="which.autoUpdateMode === 'auto_update_at_intervals'"
        :label="$t('setting.options.updateSubAtIntervals')"
      >
        <b-input
          ref="autoUpdateIntervalInput"
          v-model.number="which.autoUpdateIntervalHour"
          type="number"
          min="1"
          required
        />
      </b-field>
    </section>
    <footer class="modal-card-foot flex-end">
      <button class="button" type="button" @click="$parent.close()">
        {{ $t("operations.cancel") }}
      </button>
      <button class="button is-primary" @click="handleClickSubmit">
        {{ $t("operations.saveApply") }}
      </button>
    </footer>
  </div>
</template>

<script>
export default {
  name: "ModalSubscription",
  props: {
    which: {
      type: Object,
      default() {
        return null;
      },
    },
  },
  methods: {
    handleClickSubmit() {
      if (
        this.which.autoUpdateMode === "auto_update_at_intervals" &&
        !this.$refs.autoUpdateIntervalInput.checkHtml5Validity()
      ) {
        return;
      }
      if (this.which.autoUpdateMode !== "auto_update_at_intervals") {
        this.which.autoUpdateIntervalHour = 0;
      }
      this.$emit("submit", this.which);
    },
  },
};
</script>

<style lang="scss">
.is-twitter .is-active a {
  color: #4099ff !important;
}
.readonly {
  pointer-events: none;
}
.same-width-5 li {
  width: 5em;
}
</style>
