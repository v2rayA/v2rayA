// The choice lists two pages share: the transparent proxy mode and the
// rule port's splitting mode, as Vuetify select items.
export type T = (key: string) => string;

export const transparentModes = (t: T) => {
  const on = (label: string) => `${t("setting.options.on")}: ${label}`;
  return [
    { value: "close", title: t("setting.options.off") },
    { value: "proxy", title: on(t("setting.options.global")) },
    { value: "whitelist", title: on(t("setting.options.whitelistCn")) },
    { value: "gfwlist", title: on(t("setting.options.gfwlist")) },
    { value: "pac", title: on(t("setting.options.sameAsPacMode")) },
  ];
};

export const pacModes = (t: T) => [
  { value: "whitelist", title: t("setting.options.whitelistCn") },
  { value: "gfwlist", title: t("setting.options.gfwlist") },
  { value: "routingA", title: "RoutingA" },
];

export const subscriptionUpdateModes = (t: T) => [
  { value: "none", title: t("setting.options.off") },
  { value: "auto_update", title: t("setting.options.updateSubWhenStart") },
  {
    value: "auto_update_at_intervals",
    title: t("setting.options.updateSubAtIntervals"),
  },
];

/** the proxy an update runs through; "direct" reads as "follows the transparent proxy" when one is on */
export const updateProxyModes = (t: T, transparentOff: boolean) => [
  {
    value: "direct",
    title: transparentOff
      ? t("setting.options.direct")
      : t("setting.options.dependTransparentMode"),
  },
  { value: "proxy", title: t("setting.options.global") },
  { value: "pac", title: t("setting.options.pac") },
];

/** what the transparent proxy can be implemented with on this instance */
export const transparentTypes = (
  t: T,
  host: { lite: boolean; os: string; isRoot: boolean; tunSupported: boolean },
) => {
  const items: {
    value: string;
    title: string;
    props?: { disabled: boolean };
  }[] = [];
  if (!host.lite && host.os === "linux")
    items.push(
      { value: "redirect", title: "redirect" },
      { value: "tproxy", title: "tproxy" },
    );
  if (!host.lite)
    items.push({
      value: "tun",
      title: host.tunSupported
        ? "tun"
        : `tun — ${t("setting.options.tunUnsupported")}`,
      props: { disabled: !host.tunSupported },
    });
  if (!(host.isRoot && (host.os === "linux" || host.os === "darwin")))
    items.push({
      value: "system_proxy",
      title: t("setting.options.systemProxy"),
    });
  return items;
};
