<template>
  <div class="modal-card" style="max-width: 400px; margin: auto">
    <header class="modal-card-head">
      <p class="modal-card-title">{{ $t("configureSubscription.title") }}</p>
      <button type="button" class="delete" aria-label="close" @click="$emit('close')"></button>
    </header>
    <section class="modal-card-body">
      <b-field :label="$t('subscription.subscription')">
        <b-input
          v-model="which.address"
          type="textarea"
          :placeholder="$t('subscription.subscription')"
        />
      </b-field>
      <b-field :label="$t('subscription.remarks')">
        <b-input
          v-model="which.remarks"
          :placeholder="$t('subscription.remarks')"
        />
      </b-field>
      <b-field :label="$t('subscription.autoSelectLabel')">
        <b-checkbox
	  v-model="which.autoSelect"
	  >{{ $t("subscription.autoSelect") }}
	</b-checkbox>
      </b-field>
    </section>
    <footer class="modal-card-foot flex-end">
      <button class="button" type="button" @click="$emit('close')">
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
  emits: ["submit", "close"],
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
