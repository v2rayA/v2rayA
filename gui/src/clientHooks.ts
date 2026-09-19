// What the app does with a failed request beyond the caller's own notice:
// a 401 ends the session, and a backend that cannot be reached is
// announced once per address with the way out (the address dialog, the
// manual). This is the old axios interceptor's UI, as a banner.
import { setClientHooks, type ApiError } from "@/api/client";
import { showBanner, withdrawBanner } from "@/composables/useBanner";
import i18n from "@/plugins/i18n";
import { resetSession } from "@/session";
import { useAppStore } from "@/stores/app";

let informed = "";

export function installClientHooks(ui: { openAddressDialog(): void }): void {
  const t = i18n.global.t;
  const informNotRunning = (url: string) => {
    if (informed === url) return;
    informed = url;
    showBanner({
      key: "backend",
      kind: "warning",
      text: t("axios.messages.noBackendFound", { url }),
      action: {
        label: t("axios.messages.optimizeBackend"),
        onClick: ui.openAddressDialog,
      },
    });
  };

  setClientHooks({
    onReached() {
      // the backend is back: the banner about it goes, and a later outage
      // is announced afresh
      if (informed) withdrawBanner("backend");
      informed = "";
    },
    onUnauthorized() {
      const store = useAppStore();
      if (store.loggedIn) void resetSession({ token: "" });
    },
    onError(err: ApiError) {
      const origin = originOf(err.url);
      if (err.kind === "mixed-content") {
        const host = new URL(err.url).hostname;
        const local = ["localhost", "local", "127.0.0.1"].includes(host);
        const ua = navigator.userAgent;
        if (local && ua.includes("AppleWebKit")) {
          // Chrome allows an https page to reach http://localhost; the
          // failure means the service is not running there
          informNotRunning(origin);
          return;
        }
        const gecko = ua.includes("Gecko") && !ua.includes("KHTML");
        showBanner({
          key: "backend",
          kind: "warning",
          text: t(`axios.messages.cannotCommunicate.${local && gecko ? 1 : 0}`),
          action: {
            label: t("operations.switchSite"),
            onClick: () => window.open("http://v.v2raya.org", "_self"),
          },
        });
      } else if (err.kind === "network") {
        informNotRunning(origin);
      }
    },
  });
}

/** originOf gives the backend's origin for a request URL: the absolute address, or this page's for a relative one. */
function originOf(url: string): string {
  try {
    return new URL(url, location.href).origin;
  } catch {
    return url;
  }
}
