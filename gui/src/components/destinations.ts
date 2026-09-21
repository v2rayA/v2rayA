// The destinations the drawer, the rail and the bottom bar offer; the
// app bar titles the current one. About lives at the bottom of the
// settings page. The bottom bar holds five, Material's limit, so on a
// phone the docs move to the app bar's menu.
import {
  mdiBookOpenPageVariant,
  mdiBookOpenPageVariantOutline,
  mdiCogOutline,
  mdiCog,
  mdiRss,
  mdiRssBox,
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

/** the five the compact window's bottom bar shows */
export const barDestinations = () =>
  destinations.filter((d) => d.view !== "docs");

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
    view: "subscriptions",
    label: "common.subscriptions",
    icon: mdiRss,
    activeIcon: mdiRssBox,
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
  {
    view: "docs",
    label: "common.docs",
    icon: mdiBookOpenPageVariantOutline,
    activeIcon: mdiBookOpenPageVariant,
  },
];
