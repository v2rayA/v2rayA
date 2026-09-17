import Vue from "vue";
import Buefy from "buefy";
import { ConfigProgrammatic } from "buefy";
import "@/assets/scss/buefy.scss";
import "lucide-static/font/lucide.css";

Vue.use(Buefy);
ConfigProgrammatic.setOptions({
  defaultProgrammaticPromise: true,
  // one notice at a time: repeated clicks on a failing action used to
  // stack a screenful of identical toasts
  defaultNoticeQueue: true,
  // on phones Buefy turns every dropdown into a centred modal with a
  // backdrop; a menu should stay a menu
  defaultDropdownMobileModal: false,
  defaultIconPack: "lucide",
  customIconPacks: {
    lucide: {
      sizes: {
        default: "",
        "is-small": "",
        "is-medium": "icon-md",
        "is-large": "icon-lg",
      },
      iconPrefix: "icon-",
      internalIcons: {
        information: "info",
        check: "check",
        "check-circle": "circle-check",
        alert: "triangle-alert",
        "alert-circle": "circle-alert",
        "arrow-up": "arrow-up",
        "chevron-right": "chevron-right",
        "chevron-left": "chevron-left",
        "chevron-down": "chevron-down",
        eye: "eye",
        "eye-off": "eye-off",
        "menu-down": "chevron-down",
        "menu-up": "chevron-up",
        "close-circle": "circle-x",
        loading: "loader",
      },
    },
  },
});
