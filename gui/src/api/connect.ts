// The old networkInspect.waitingConnected: while a request that changes
// the proxy runs (start, connect, a new outbound), the transparent proxy
// it sets up can cut the very connection carrying the answer. So GET
// /touch is polled meanwhile; once the core reports running with a
// connected server the work is done, the request is abandoned and the
// poll stops. The poll also stops when the request settles, on a 401,
// and after `timeout` at the latest.
import { getTouch } from "./index";
import { ApiError } from "./client";

export interface WatchOptions {
  /** poll period and the poll's own timeout, ms */
  interval?: number;
  /** hard stop for the poll, ms */
  timeout?: number;
  /** a FAIL from /touch other than "request in progress" */
  onCheckFailed?: (err: ApiError) => void;
}

/** watchConnected runs `request` and resolves with its result, or with `undefined` when the poll saw the core connected first. */
export function watchConnected<T>(
  request: Promise<T>,
  abort: () => void,
  { interval = 3000, timeout = 30_000, onCheckFailed }: WatchOptions = {},
): Promise<T | undefined> {
  let settled = false;
  let timer: ReturnType<typeof setInterval> | null = null;
  const stop = () => {
    if (timer !== null) clearInterval(timer);
    timer = null;
  };
  return new Promise<T | undefined>((resolve, reject) => {
    const finish = (fn: () => void) => {
      if (settled) return;
      settled = true;
      stop();
      fn();
    };
    timer = setInterval(async () => {
      try {
        const touch = await getTouch({ timeout: interval });
        if (touch.running && touch.touch.connectedServer) {
          abort();
          finish(() => resolve(undefined));
        }
      } catch (err) {
        if (!(err instanceof ApiError)) return;
        if (err.status === 401) {
          abort();
          finish(() => reject(err));
        } else if (
          err.kind === "http" &&
          err.body?.errorCode !== "REQUEST_IN_PROGRESS" &&
          err.body?.message !== "the last request is being processed"
        ) {
          onCheckFailed?.(err);
        }
      }
    }, interval);
    setTimeout(stop, timeout);
    request.then(
      (v) => finish(() => resolve(v)),
      (e) => {
        // the abort above rejects the request with a cancellation, which
        // the poll already answered
        if (settled) return;
        finish(() => reject(e));
      },
    );
  });
}
