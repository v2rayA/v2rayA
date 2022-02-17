import Vue from "vue";
import Buefy from "buefy";
import { ConfigProgrammatic } from "buefy";
import "@/assets/scss/buefy.scss";
import "@mdi/font/css/materialdesignicons.css";

Vue.use(Buefy);
ConfigProgrammatic.setOptions({
  defaultProgrammaticPromise: true
});
