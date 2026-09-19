import { reactive, readonly } from "vue";
import i18n from "@/plugins/i18n";
import { refresh } from "@/session/refresh";

// Notices show one at a time, in order — the old GUI's defaultNoticeQueue:
// a failing action clicked five times shows one toast, not a stack. Text
// only; a backend message is never rendered as HTML.
export type NoticeKind = "info" | "success" | "warning" | "error";

export interface Notice {
  id: number;
  kind: NoticeKind;
  text: string;
  /** milliseconds; 0 keeps it until dismissed */
  timeout: number;
  action?: { label: string; onClick: () => void };
}

const state = reactive({
  queue: [] as Notice[],
  current: null as Notice | null,
});
let nextId = 1;

function advance() {
  state.current = state.queue.shift() ?? null;
}

/** a notice saying the backend is still busy offers a refresh, so nobody has to guess */
function refreshActionFor(text: string): Notice["action"] | undefined {
  const busy = i18n.global.t("backend.REQUEST_IN_PROGRESS");
  if (!busy || !text.includes(busy)) return undefined;
  return { label: i18n.global.t("operations.refresh"), onClick: refresh };
}

function push(
  kind: NoticeKind,
  text: string,
  options: { timeout?: number; action?: Notice["action"] } = {},
): number {
  const notice: Notice = {
    id: nextId++,
    kind,
    text,
    timeout: options.timeout ?? (kind === "error" ? 8000 : 4000),
    action: options.action ?? refreshActionFor(text),
  };
  state.queue.push(notice);
  if (!state.current) advance();
  return notice.id;
}

/** dismiss ends the notice being shown (or removes a queued one) and shows the next. */
export function dismissNotice(id?: number): void {
  if (id !== undefined && state.current?.id !== id) {
    const i = state.queue.findIndex((n) => n.id === id);
    if (i >= 0) state.queue.splice(i, 1);
    return;
  }
  advance();
}

/** closeAllNotices drops the queue and the notice on screen; the session reset calls it. */
export function closeAllNotices(): void {
  state.queue.length = 0;
  state.current = null;
}

export const noticeState = readonly(state);

export function useNotify() {
  return {
    info: (text: string, o?: { timeout?: number; action?: Notice["action"] }) =>
      push("info", text, o),
    success: (
      text: string,
      o?: { timeout?: number; action?: Notice["action"] },
    ) => push("success", text, o),
    warning: (
      text: string,
      o?: { timeout?: number; action?: Notice["action"] },
    ) => push("warning", text, o),
    error: (
      text: string,
      o?: { timeout?: number; action?: Notice["action"] },
    ) => push("error", text, o),
    dismiss: dismissNotice,
  };
}
