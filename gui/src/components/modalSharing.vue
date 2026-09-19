<template>
  <div class="modal-card" style="max-width: 500px; margin: auto">
    <header class="modal-card-head">
      <p class="modal-card-title has-text-centered">{{ title }}</p>
      <button type="button" class="delete" aria-label="close" @click="$emit('close')"></button>
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

  </div>
</template>

<script>
import QRCode from "qrcode";
import CONST from "@/assets/js/const";
import { Base64 } from "js-base64";
import i18n from "@/plugins/i18n";

export default {
  name: "ModalSharing",
  emits: ["close"],
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

  mounted() {
    let add = this.sharingAddress;
    if (this.type === CONST.SubscriptionType) {
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
};
</script>

<style scoped>
.modal-card-head,
.modal-card-foot {
  border-radius: 0.25rem;
}
.modal-card-head {
  border-bottom-left-radius: 0;
  border-bottom-right-radius: 0;
}
.modal-card-foot {
  border-top-left-radius: 0;
  border-top-right-radius: 0;
}
</style>
