// The destinations the drawer, the rail and the bottom bar offer; the
// app bar titles the current one. About lives at the bottom of the
// settings page.
import {
  mdiCogOutline,
  mdiCog,
  mdiScriptTextOutline,
  mdiScriptText,
  mdiServerNetworkOutline,
  mdiServerNetwork,
  mdiViewDashboardOutline,
  mdiViewDashboard,
} from "@mdi/js";
import type { View } from "@/stores/app";

export interface Destination {
  view: View;
  /** locale key of the label and the page title */
  label: string;
  icon: string;
  /** the filled variant, shown when active */
  activeIcon: string;
}

export const destinations: Destination[] = [
  {
    view: "dashboard",
    label: "common.dashboard",
    icon: mdiViewDashboardOutline,
    activeIcon: mdiViewDashboard,
  },
  {
    view: "proxies",
    label: "common.proxies",
    icon: mdiServerNetworkOutline,
    activeIcon: mdiServerNetwork,
  },
  {
    view: "settings",
    label: "common.setting",
    icon: mdiCogOutline,
    activeIcon: mdiCog,
  },
  {
    view: "logs",
    label: "common.log",
    icon: mdiScriptTextOutline,
    activeIcon: mdiScriptText,
  },
];
