import "@/plugins/apiRoot";
import "@/plugins/buefy";
import "@/plugins/axios";
import "@/plugins/backendPort";
import "@/plugins/dayjs";
import "@/plugins/virtual-scroll";
import "normalize.css";
import "pace-js";
import "pace-js/themes/blue/pace-theme-corner-indicator.css";

import { buildApp } from "@/plugins/session";

buildApp().mount("#app");
