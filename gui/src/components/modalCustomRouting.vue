<template>
  <div
    class="modal-card modal-configure-pac"
    style="max-width: 550px; height: 700px; margin: auto"
  >
    <header class="modal-card-head">
      <p class="modal-card-title">{{ $t("customRouting.title") }}</p>
      <button type="button" class="delete" aria-label="close" @click="$emit('close')"></button>
    </header>
    <section class="modal-card-body rules">
      <b-message type="is-info" class="after-line-dot5">
        <p
          v-html="
            $t('customRouting.messages.0', {
              V2RayLocationAsset,
            })
          "
        />
        <p
          v-html="
            $t('customRouting.messages.1', {
              V2RayLocationAsset,
            })
          "
        ></p>
      </b-message>
      <b-message type="is-success" class="after-line-dot5">
        <p>{{ $t("customRouting.messages.2") }}</p>
      </b-message>
      <b-collapse class="card">
        <template #trigger="props">
        <div
          class="card-header"
          role="button"
        >
          <p class="card-header-title">
            {{ $t("customRouting.defaultRoutingRule") }}
          </p>
          <a class="card-header-icon">
            <b-icon :icon="props.open ? 'chevron-down' : 'chevron-up'"> </b-icon>
          </a>
        </div>
        </template>
        <div class="card-content">
          <b-field
            :label="$t('customRouting.defaultRoutingRule')"
            label-position="on-border"
          >
            <b-select v-model="customPac.defaultProxyMode" expanded>
              <option value="direct">{{ $t("customRouting.direct") }}</option>
              <option value="proxy">{{ $t("customRouting.proxy") }}</option>
              <option value="block">{{ $t("customRouting.block") }}</option>
            </b-select>
          </b-field>
        </div>
      </b-collapse>
      <b-collapse
        v-for="(rule, index) of customPac.routingRules"
        :key="rule.value"
        class="card"
      >
        <template #trigger="props">
        <div
          class="card-header"
          role="button"
        >
          <p class="card-header-title" style="position: relative">
            <span>{{ `${$t("customRouting.rule")}${index + 1}` }}</span>
            <b-button
              type="is-text"
              size="is-small"
              style="position: absolute; right: 0"
              @click="handleClickDeleteRule(...arguments, index)"
              >{{ $t("operations.delete") }}</b-button
            >
          </p>
          <a class="card-header-icon">
            <b-icon :icon="props.open ? 'chevron-down' : 'chevron-up'"></b-icon>
          </a>
        </div>
        </template>
        <div class="card-content">
          <b-field
            :label="$t('customRouting.domainFile')"
            label-position="on-border"
          >
            <b-select v-model="rule.filename" expanded>
              <option
                v-for="file of siteDatFiles"
                :key="file.filename"
                :value="file.filename"
              >
                {{ file.filename }}
              </option>
            </b-select>
          </b-field>
          <b-field :label="$t('customRouting.tags')" label-position="on-border">
            <b-select
              v-model="rule.tags"
              multiple
              :native-size="Math.min(16, tagsOf(rule.filename).length)"
              size="is-small"
              expanded
            >
              <option
                v-for="tag of tagsOf(rule.filename)"
                :key="tag"
                :value="tag"
              >
                {{ tag }}
              </option>
            </b-select>
          </b-field>
          <p class="content" style="font-size: 0.8em; margin-left: 0.5em">
            {{ $t("customRouting.tags") }}: {{ rule.tags }}
          </p>
          <b-field
            :label="$t('customRouting.typeRule')"
            label-position="on-border"
          >
            <b-select v-model="rule.ruleType" expanded>
              <option value="direct">
                {{
                  $t("customRouting.direct") +
                  (customPac.defaultProxyMode === "direct"
                    ? `(${$t("customRouting.sameAsDefaultRule")})`
                    : "")
                }}
              </option>
              <option value="proxy">
                {{
                  $t("customRouting.proxy") +
                  (customPac.defaultProxyMode === "proxy"
                    ? `(${$t("customRouting.sameAsDefaultRule")})`
                    : "")
                }}
              </option>
              <option value="block">
                {{
                  $t("customRouting.block") +
                  (customPac.defaultProxyMode === "block"
                    ? `(${$t("customRouting.sameAsDefaultRule")})`
                    : "")
                }}
              </option>
            </b-select>
          </b-field>
        </div>
      </b-collapse>
    </section>
    <footer class="modal-card-foot">
      <div
        style="
          position: relative;
          display: flex;
          justify-content: flex-end;
          width: 100%;
        "
      >
        <button class="button btn-new" type="button" @click="handleNew">
          {{ $t("customRouting.appendRule") }}
        </button>
        <button class="button" type="button" @click="$emit('close')">
          {{ $t("operations.cancel") }}
        </button>
        <button class="button is-primary" @click="handleClickSubmit">
          {{ $t("operations.save") }}
        </button>
      </div>
    </footer>
  </div>
</template>

<script>
import { handleResponse } from "@/assets/js/utils";
export default {
  name: "ModalCustomRouting",
  emits: ["close"],
  data: () => ({
    customPac: {
      defaultProxyMode: "",
      routingRules: [],
    },
    siteDatFiles: [],
    firstSiteDatFilename: "",
    V2RayLocationAsset: "",
  }),
  mounted() {
    (async () => {
      let customPac = this.customPac;
      await this.$axios({
        url: apiRoot + "/customPac",
      }).then((res) => {
        handleResponse(res, this, () => {
          customPac = res.data.data.customPac;
          this.V2RayLocationAsset = res.data.data.V2RayLocationAsset;
        });
      });
      let closing = false;
      await this.$axios({
        url: apiRoot + "/siteDatFiles",
      }).then((res) => {
        handleResponse(res, this, () => {
          if (
            res.data.data.siteDatFiles &&
            res.data.data.siteDatFiles.length > 0
          ) {
            //将数组转换为map
            this.siteDatFiles = {};
            res.data.data.siteDatFiles.forEach((x) => {
              this.siteDatFiles[x.filename] = x;
              x.tags.sort(); //对tags进行排序，方便查找
            });
            this.firstSiteDatFilename = res.data.data.siteDatFiles[0].filename;
          } else {
            this.$buefy.toast.open({
              message: this.$t("customRouting.messages.noSiteDatFileFound", {
                V2RayLocationAsset: this.V2RayLocationAsset,
              }),
              type: "is-warning",
              position: "is-top",
              duration: 5000,
            });
            closing = true;
          }
        });
      });
      this.customPac = customPac;
      if (closing) {
        this.$emit("close");
      }
    })();
  },
  methods: {
    tagsOf(filename) {
      // a saved rule may name a .dat file that no longer exists
      const f = this.siteDatFiles[filename];
      return f && f.tags ? f.tags : [];
    },
    handleNew() {
      this.customPac.routingRules.push({
        filename: this.firstSiteDatFilename,
        tags: [],
        matchType: "domain",
        ruleType:
          this.customPac.defaultProxyMode === "direct" ? "proxy" : "direct",
      });
      let target = document.querySelector(".modal-configure-pac .rules");
      this.$nextTick(() => {
        target.scrollTop = target.scrollHeight - target.clientHeight;
      });
    },
    handleClickDeleteRule(event, index) {
      event.stopImmediatePropagation();
      delete this.customPac.routingRules.splice(index, 1);
    },
    handleClickSubmit() {
      if (this.customPac.routingRules.some((x) => x.tags.length <= 0)) {
        this.$buefy.toast.open({
          message: this.$t("customRouting.messages.emptyRuleNotPermitted"),
          type: "is-warning",
          position: "is-top",
          duration: 3000,
        });
        return;
      }
      this.$axios({
        url: apiRoot + "/customPac",
        method: "put",
        data: {
          customPac: this.customPac,
        },
      }).then((res) => {
        handleResponse(res, this, () => {
          this.$emit("close");
        }, null, "customRouting.saveFailed");
      });
    },
  },
};
</script>

<style lang="scss" scoped>
.btn-new {
  position: absolute;
  left: 0;
  top: 0;
}
</style>
<style lang="scss">
.icon-tag {
  font-size: 24px !important;
}
.after-line-dot5 {
  font-size: 14px;
  p {
    font-size: 14px;
  }
}
</style>
