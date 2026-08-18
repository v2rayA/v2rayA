<template>
  <div class="modal-card" style="max-width: 420px; margin: auto">
    <header class="modal-card-head">
      <p class="modal-card-title">{{ $t("proxyGroup.pickTitle") }}</p>
    </header>
    <section class="modal-card-body">
      <b-field :label="$t('proxyGroup.group')" label-position="on-border">
        <b-select v-model="selectedGroup" expanded>
          <option v-for="group in groups" :key="group" :value="group">
            {{ group.toUpperCase() }}
          </option>
        </b-select>
      </b-field>
      <p class="help is-info">{{ $t("proxyGroup.pickMessage") }}</p>
    </section>
    <footer class="modal-card-foot flex-end">
      <b-button @click="$parent.close()">{{ $t("operations.cancel") }}</b-button>
      <b-button type="is-primary" @click="handleClickConfirm">{{ $t("operations.confirm") }}</b-button>
    </footer>
  </div>
</template>

<script>
export default {
  name: "ModalPickProxyGroup",
  props: {
    groups: {
      type: Array,
      default: () => ["proxy"],
    },
    initialGroup: {
      type: String,
      default: "proxy",
    },
  },
  data() {
    return {
      selectedGroup: this.initialGroup,
    };
  },
  methods: {
    handleClickConfirm() {
      if (!this.selectedGroup) {
        return;
      }
      this.$emit("select", this.selectedGroup);
      this.$parent.close();
    },
  },
};
</script>
