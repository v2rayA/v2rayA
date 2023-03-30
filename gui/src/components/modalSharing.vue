<template>
  <div class="modal-card" style="max-width: 500px; margin: auto">
    <header class="modal-card-head">
      <p class="modal-card-title has-text-centered">{{ title }}</p>
    </header>
    <section class="modal-card-body lazy" style="text-align: center">
      <div><canvas id="canvas" class="qrcode"></canvas></div>
      <div class="tags has-addons is-centered" style="position: relative">
        <span
          class="tag is-rounded is-dark sharingAddressTag"
          style="position: relative"
          :data-clipboard-text="sharingAddress"
        >
          <div class="tag-cover tag is-rounded" style="display: none"></div>
          <span class="has-ellipsis" style="max-width: 10em">
            {{ shortDesc }}
          </span>
        </span>
        <div id="tag-cover-text">{{ $t("operations.copyLink") }}</div>
        <span
          class="tag is-rounded is-primary sharingAddressTag"
          style="position: relative"
          :data-clipboard-text="sharingAddress"
        >
          <span class="has-ellipsis" style="max-width: 25em">
            {{ sharingAddress }}
          </span>
          <div class="tag-cover tag is-rounded" style="display: none"></div>
        </span>
      </div>
    </section>
    <footer class="modal-card-foot" style="justify-content: center">
      <a
        class="is-link"
        href="https://github.com/v2rayA/v2rayA"
        target="_blank"
      >
        <img
          class="leave-right"
          src="https://img.shields.io/github/stars/mzz2017/v2rayA.svg?style=social"
          alt="stars"
        />
        <img
          class="leave-right"
          src="https://img.shields.io/github/forks/mzz2017/v2rayA.svg?style=social"
          alt="forks"
        />
        <img
          class="leave-right"
          src="https://img.shields.io/github/watchers/mzz2017/v2rayA.svg?style=social"
          alt="watchers"
        />
      </a>
    </footer>
  </div>
</template>

<script>
import QRCode from "qrcode";
import { Decoder } from "@nuintun/qrcode";
import ClipboardJS from "clipboard";
import CONST from "@/assets/js/const";
import { Base64 } from "js-base64";
import i18n from "@/plugins/i18n";

export default {
  name: "ModalSharing",
  i18n,
  props: {
    title: {
      type: String,
      required: true,
    },
    sharingAddress: {
      type: String,
      required: true,
    },
    shortDesc: {
      type: String,
      required: true,
    },
    type: {
      type: String,
      required: true,
    },
  },
  beforeDestroy() {
    this.clipboard.destroy();
  },
  mounted() {
    document
      .querySelector("#QRCodeImport")
      .addEventListener("change", this.handleFileChange, false);
    this.clipboard = new ClipboardJS(".sharingAddressTag");
    this.clipboard.on("success", (e) => {
      this.$buefy.toast.open({
        message: this.$t("common.success"),
        type: "is-primary",
        position: "is-top",
        queue: false,
      });
      e.clearSelection();
    });
    this.clipboard.on("error", (e) => {
      this.$buefy.toast.open({
        message: this.$t("common.fail") + ", error:" + e.toLocaleString(),
        type: "is-warning",
        position: "is-top",
        queue: false,
      });
    });

    let add = this.sharingAddress;
    if (this._type === CONST.SubscriptionType) {
      add = "sub://" + Base64.encode(add);
    }
    let canvas = document.getElementById("canvas");
    QRCode.toCanvas(
      canvas,
      add,
      { errorCorrectionLevel: "H" },
      function (error) {
        if (error) console.error(error);
        // console.log("QRCode has been generated successfully!");
      }
    );
    let targets = document.querySelectorAll(".sharingAddressTag");
    let covers = document.querySelectorAll(".tag-cover");
    let coverText = document.querySelector("#tag-cover-text");
    let enter = () => {
      covers.forEach((x) => (x.style.display = "unset"));
      coverText.style.display = "flex";
    };
    let leave = () => {
      covers.forEach((x) => (x.style.display = "none"));
      coverText.style.display = "none";
    };
    targets.forEach((x) => x.addEventListener("mouseenter", enter));
    targets.forEach((x) => x.addEventListener("mouseleave", leave));
  },
  methods: {
    handleFileChange(e) {
      const that = this;
      const file = e.target.files[0];
      let elem = document.querySelector("#QRCodeImport");
      // eslint-disable-next-line no-self-assign
      elem.outerHTML = elem.outerHTML;
      this.$nextTick(() => {
        document
          .querySelector("#QRCodeImport")
          .addEventListener("change", this.handleFileChange, false);
      });
      // console.log(file);
      if (!file.type.match(/image\/.*/)) {
        this.$buefy.toast.open({
          message: this.$t("import.qrcodeError"),
          type: "is-warning",
          position: "is-top",
          queue: false,
        });
        return;
      }
      const reader = new FileReader();
      reader.onload = function (e) {
        // target.result 该属性表示目标对象的DataURL
        // console.log(e.target.result);
        const file = e.target.result;
        const qrcode = new Decoder();
        qrcode
          .scan(file)
          .then((result) => {
            console.log(result);
            that.handleClickImportConfirm(result.data);
          })
          .catch((error) => {
            console.error(error);
            that.$buefy.toast.open({
              message: that.$t("import.qrcodeError"),
              type: "is-warning",
              position: "is-top",
              queue: false,
            });
          });
      };
      reader.readAsDataURL(file);
    },
  },
};
</script>

<style scoped></style>
