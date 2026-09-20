import { template } from "./template";

// Ready-made rule sets for the reference panel. Every outbound is one of the
// built-in three, so a preset never fails the save-time check that a named
// group has a connected node; the description tells the user to rename
// `proxy` to their own group when they want one. Category names must exist in
// the v2fly geosite.dat and geoip.dat that v2rayA downloads: no Loyalsoldier
// extras such as gfw, apple-cn or geoip:telegram. Telegram's ranges are the
// ones template_routing.go uses for the same reason.
const telegramRanges =
  'ip("91.105.192.0/23", "91.108.4.0/22", "91.108.8.0/21", "91.108.16.0/21", "91.108.56.0/22", "95.161.64.0/20", "149.154.160.0/20", "185.76.151.0/24", "2001:67c:4e8::/48", "2001:b28:f23c::/47", "2001:b28:f23f::/48", "2a0a:f280:203::/48")->proxy';

export const presets: { key: string; code: string }[] = [
  { key: "whitelist", code: template },
  {
    key: "blacklist",
    code: `default: direct
# only sites outside China and Telegram use the proxy
domain(geosite:geolocation-!cn)->proxy
domain(geosite:telegram)->proxy
${telegramRanges}`,
  },
  { key: "ads", code: "domain(geosite:category-ads-all)->block" },
  {
    key: "streaming",
    code: "domain(geosite:netflix, geosite:disney, geosite:hbo, geosite:primevideo, geosite:youtube, geosite:spotify)->proxy",
  },
  {
    key: "telegram",
    code: `domain(geosite:telegram)->proxy
${telegramRanges}`,
  },
  {
    key: "ai",
    code: "domain(geosite:openai, geosite:anthropic, geosite:google-gemini)->proxy",
  },
  {
    key: "cnServices",
    code: "domain(geosite:apple@cn, geosite:google@cn, geosite:microsoft@cn, geosite:steam@cn, geosite:category-games@cn, geosite:bilibili)->direct",
  },
  {
    key: "lan",
    code: `domain(geosite:private)->direct
ip(geoip:private)->direct`,
  },
];
