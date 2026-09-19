// A session is one token on one backend. Logging in or out and changing
// the backend address reset it; the old GUI rebuilt its whole root for
// that. Here the reset is an explicit order: stop what the session started,
// refuse late responses, close what is on screen, reset the stores, start
// again.
import { getActivePinia, type Pinia } from "pinia";
import { abortSession } from "@/api/client";
import { closeAllDialogs } from "@/composables/useDialog";
import { closeAllNotices } from "@/composables/useNotify";
import { closeAllBanners } from "@/composables/useBanner";
import { closeAllLoadings } from "@/composables/useLoading";
import { useAppStore } from "@/stores/app";

type Teardown = () => void;
const teardowns = new Set<Teardown>();
let starter: (() => void | Promise<void>) | null = null;

/** onSessionTeardown registers what a session started (a socket, a listener, a timer); returns the unregister. */
export function onSessionTeardown(fn: Teardown): () => void {
  teardowns.add(fn);
  return () => teardowns.delete(fn);
}

/** setSessionStarter names the routine that brings a session up: account check, version, socket. */
export function setSessionStarter(fn: () => void | Promise<void>): void {
  starter = fn;
}

export interface SessionChange {
  token?: string;
  backendAddress?: string;
}

/** resetSession tears the current session down in order and starts the next one. */
export async function resetSession(
  change: SessionChange = {},
  pinia: Pinia | undefined = getActivePinia(),
): Promise<void> {
  // 1. what the session started
  for (const fn of [...teardowns]) fn();
  teardowns.clear();
  // 2. requests in flight are aborted and their late answers refused
  abortSession();
  // 3. nothing of the old session stays on screen
  closeAllDialogs();
  closeAllNotices();
  closeAllBanners();
  closeAllLoadings();
  // the dialogs' result handlers run before the next session starts
  await new Promise((r) => setTimeout(r, 0));
  // 4. the stores, then the facts that identify the next session
  const app = useAppStore(pinia);
  app.$reset();
  if (change.backendAddress !== undefined)
    app.setBackendAddress(change.backendAddress);
  if (change.token !== undefined) app.setToken(change.token);
  // 5. up again
  await starter?.();
}
