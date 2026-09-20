// The editor's state and its two requests, without the view: which
// protocol tab is up, the model behind each tab, loading a node's share
// link into the model and saving the model as a share link. The dialog
// renders it; the spec drives it.
import { reactive, ref, type Ref } from "vue";
import { getSharingAddress, postImport } from "@/api";
import { timeouts } from "@/api/client";
import type { Which } from "@/api/types";
import { generateShareLink, parseShareLink } from "@/lib/serverCodec";
import { defaultModels, type Models } from "./models";

export const protocols = [
  "vmess",
  "vless",
  "wireguard",
  "ss",
  "trojan",
  "juicity",
  "tuic",
  "hysteria2",
  "http",
  "socks5",
  "anytls",
] as const;
export type Protocol = (typeof protocols)[number];

export const protocolLabels: Record<Protocol, string> = {
  vmess: "VMESS",
  vless: "VLESS",
  wireguard: "WireGuard",
  ss: "SS",
  trojan: "Trojan",
  juicity: "Juicity",
  tuic: "Tuic",
  hysteria2: "Hysteria2",
  http: "HTTP",
  socks5: "SOCKS5",
  anytls: "AnyTLS",
};

// A share link's scheme names its protocol; vmess and vless share one
// model, so the model key is separate from the tab.
const schemes: [prefix: string, protocol: Protocol][] = [
  ["vmess://", "vmess"],
  ["vless://", "vless"],
  ["wireguard://", "wireguard"],
  ["ss://", "ss"],
  ["trojan://", "trojan"],
  ["trojan-go://", "trojan"],
  ["juicity://", "juicity"],
  ["tuic://", "tuic"],
  ["hysteria2://", "hysteria2"],
  ["hy2://", "hysteria2"],
  ["http://", "http"],
  ["https://", "http"],
  ["socks5://", "socks5"],
  ["anytls://", "anytls"],
];

export function protocolOf(link: string): Protocol | null {
  const lower = link.toLowerCase();
  const hit = schemes.find(([prefix]) => lower.startsWith(prefix));
  return hit ? hit[1] : null;
}

export function modelKey(protocol: Protocol): keyof Models {
  return protocol === "vmess" || protocol === "vless" ? "v2ray" : protocol;
}

export interface Editor {
  models: Models;
  protocol: Ref<Protocol>;
  /** load fills the model from the node's share link; resolves false when the link is not one the editor knows */
  load(): Promise<boolean>;
  /** link is the share link of the current tab's model */
  link(): string | null;
  /** save imports the link as this node (or as a new one); the backend's TouchResponse follows */
  save(): Promise<void>;
}

export function createEditor(which: Which | null): Editor {
  const models = reactive(defaultModels()) as Models;
  const protocol = ref<Protocol>("vmess");

  async function load(): Promise<boolean> {
    if (which === null) return true;
    const { sharingAddress } = await getSharingAddress(which);
    const p = protocolOf(sharingAddress);
    const form = p && parseShareLink(sharingAddress);
    if (!p || !form) return false;
    // the parsed form replaces the model, as the old editor did: a key the
    // link does not carry is absent, not the default
    (models as Record<string, unknown>)[modelKey(p)] = form;
    protocol.value = p;
    return true;
  }

  function link(): string | null {
    const model = models[modelKey(protocol.value)];
    if (modelKey(protocol.value) === "v2ray") model.protocol = protocol.value;
    return generateShareLink(model);
  }

  async function save(): Promise<void> {
    const url = link();
    if (url === null) throw new Error("no link for " + protocol.value);
    await postImport({ url, kind: "server", which }, timeouts.none);
  }

  return { models, protocol, load, link, save };
}
