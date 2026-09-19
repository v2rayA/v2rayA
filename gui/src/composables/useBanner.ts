import { reactive, readonly } from "vue";

// Banners are the persistent notices — a new version, a core that does
// not match, a backend that cannot be reached — shown at the top of the
// page until dismissed or withdrawn, with an action when there is one.
// A notice with the same key replaces the earlier one.
export interface Banner {
  key: string;
  kind: "info" | "warning" | "error";
  text: string;
  action?: { label: string; onClick: () => void };
  /** the user may close it; false keeps it until withdrawn */
  dismissible: boolean;
}

const state = reactive({ list: [] as Banner[] });

/** showBanner adds or replaces the banner with that key. */
export function showBanner(
  banner: Omit<Banner, "dismissible"> & { dismissible?: boolean },
): void {
  const entry: Banner = { dismissible: true, ...banner };
  const i = state.list.findIndex((b) => b.key === banner.key);
  if (i >= 0) state.list.splice(i, 1, entry);
  else state.list.push(entry);
}

export function withdrawBanner(key: string): void {
  const i = state.list.findIndex((b) => b.key === key);
  if (i >= 0) state.list.splice(i, 1);
}

/** closeAllBanners drops every banner; the session reset calls it. */
export function closeAllBanners(): void {
  state.list.length = 0;
}

export const bannerState = readonly(state);

export function useBanner() {
  return { show: showBanner, withdraw: withdrawBanner };
}
