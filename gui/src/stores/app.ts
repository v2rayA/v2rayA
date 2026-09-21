import { defineStore } from "pinia";
import { Base64 } from "js-base64";
import type { OutboundStatus, Which, VersionResponse } from "@/api/types";
import { brandSeed, isSeed } from "@/theme/scheme";

/** normalizeOutbounds keeps the backend's list as names: trimmed, unique, "proxy" first. */
export function normalizeOutbounds(outbounds: unknown): string[] {
  const seen = new Set<string>();
  const names: string[] = [];
  if (Array.isArray(outbounds)) {
    for (const outbound of outbounds) {
      if (typeof outbound !== "string") continue;
      const name = outbound.trim();
      if (!name || seen.has(name)) continue;
      seen.add(name);
      names.push(name);
    }
  }
  if (!seen.has("proxy")) names.unshift("proxy");
  return names;
}

/** The core's state as the backend reports it; text for it comes from the locale. */
export type Running = "checking" | "running" | "stopped" | "paused";
export type ThemePreference = "auto" | "light" | "dark";
/** The page's destinations, in the order the rail and the bar show them. */
export type View =
  | "dashboard"
  | "proxies"
  | "subscriptions"
  | "settings"
  | "logs"
  | "docs"
  | "about";

// One store for the session-wide state the old App.vue kept in data and
// localStorage: what was a translated text ("正在运行") is an enum here, so
// switching the language does not have to rebuild the page.
export const useAppStore = defineStore("app", {
  state: () => ({
    token: localStorage.getItem("token") ?? "",
    backendAddress: localStorage.getItem("backendAddress") ?? "",
    running: "checking" as Running,
    networkPaused: false,
    connectedServer: [] as Which[],
    outboundName: "proxy",
    outbounds: ["proxy"] as string[],
    /** the last observatory frame of each outbound group: what the core sees of its members */
    observatory: {} as Record<string, OutboundStatus[]>,
    version: null as VersionResponse | null,
    /** the last /version banner facts, seeded from localStorage before the request answers */
    // "1"/"0" as the backend sends it; the old settings dialog parses it
    lite: parseInt(localStorage.getItem("lite") ?? "0") > 0,
    docker: localStorage.getItem("docker") === "true",
    variant: localStorage.getItem("variant") ?? "",
    loadBalanceValid: localStorage.getItem("loadBalanceValid") !== "false",
    coreVersionValid: localStorage.getItem("coreVersionValid") !== "false",
    coreVersionErr: localStorage.getItem("coreVersionErr") ?? "",
    themePreference: (["light", "dark"].includes(
      localStorage.getItem("theme") ?? "",
    )
      ? localStorage.getItem("theme")
      : "auto") as ThemePreference,
    /** the seed colour the theme's palettes derive from */
    themeSeed: (() => {
      const seed = localStorage.getItem("themeSeed") ?? "";
      return isSeed(seed) ? seed.toLowerCase() : brandSeed;
    })(),
    language: localStorage.getItem("_lang") ?? "",
    view: "dashboard" as View,
    /** the documentation section to show; "" is the first */
    docsSection: "",
  }),
  getters: {
    loggedIn: (s) => s.token !== "",
    /** the name inside the token; empty when there is no token or it does not parse */
    username: (s) => {
      if (!s.token) return "";
      try {
        const payload = JSON.parse(Base64.decode(s.token.split(".")[1]));
        return typeof payload.uname === "string" ? payload.uname : "";
      } catch {
        return "";
      }
    },
  },
  actions: {
    setToken(token: string) {
      this.token = token;
      if (token) localStorage.setItem("token", token);
      else localStorage.removeItem("token");
    },
    setBackendAddress(address: string) {
      this.backendAddress = address;
      localStorage.setItem("backendAddress", address);
    },
    setRunning(running: Running, networkPaused = false) {
      this.running = running;
      this.networkPaused = networkPaused;
    },
    setOutbounds(outbounds: unknown) {
      this.outbounds = normalizeOutbounds(outbounds);
      if (!this.outbounds.includes(this.outboundName))
        this.outboundName = "proxy";
      // a deleted group's last frame would otherwise outlive it
      for (const name of Object.keys(this.observatory))
        if (!this.outbounds.includes(name)) delete this.observatory[name];
    },
    setTheme(preference: ThemePreference) {
      this.themePreference = preference;
      localStorage.setItem("theme", preference);
    },
    setThemeSeed(seed: string) {
      if (!isSeed(seed)) return;
      this.themeSeed = seed.toLowerCase();
      localStorage.setItem("themeSeed", this.themeSeed);
    },
    setLanguage(code: string) {
      this.language = code;
      localStorage.setItem("_lang", code);
    },
    applyVersion(v: VersionResponse) {
      this.version = v;
      this.lite = v.lite > 0;
      this.docker = !!v.docker;
      this.variant = v.variant;
      this.loadBalanceValid = v.loadBalanceValid;
      this.coreVersionValid = v.coreVersionValid;
      this.coreVersionErr = v.coreVersionErr ?? "";
      localStorage.setItem("lite", String(v.lite));
      localStorage.setItem("docker", String(!!v.docker));
      localStorage.setItem("variant", v.variant ?? "");
      localStorage.setItem("loadBalanceValid", String(v.loadBalanceValid));
      localStorage.setItem("coreVersionValid", String(v.coreVersionValid));
      localStorage.setItem("coreVersionErr", v.coreVersionErr ?? "");
      localStorage.setItem("version", v.version ?? "");
    },
  },
});
