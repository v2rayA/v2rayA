// One HTTP client for the redo. It carries what plugins/axios.js gave the
// old components — the JSON query serializer the backend reads, the token
// and request id on the backend's own requests only, the 401 handling —
// and nothing of the old file's UI: an error comes back as an ApiError the
// caller (or the app's error handler) turns into a notice.
import axios, {
  AxiosError,
  type AxiosRequestConfig,
  type AxiosResponse,
} from "axios";
import { nanoid } from "nanoid";
import { reactive, readonly } from "vue";

export type ApiErrorKind =
  "http" | "network" | "timeout" | "mixed-content" | "cancelled" | "stale";

export class ApiError extends Error {
  kind: ApiErrorKind;
  status?: number;
  /** The backend's envelope when it answered: code, message, errorCode, params. */
  body?: ApiEnvelope<unknown>;
  url: string;
  constructor(
    kind: ApiErrorKind,
    message: string,
    url: string,
    extra: { status?: number; body?: ApiEnvelope<unknown> } = {},
  ) {
    super(message);
    this.name = "ApiError";
    this.kind = kind;
    this.url = url;
    this.status = extra.status;
    this.body = extra.body;
  }
}

/** The backend's response envelope. */
export interface ApiEnvelope<T> {
  code: "SUCCESS" | "FAIL" | string;
  message?: string;
  errorCode?: string;
  params?: Record<string, unknown>;
  data: T;
}

/** The backend address as stored; empty means same origin. */
export function backendAddress(): string {
  return localStorage.getItem("backendAddress") ?? "";
}

export function apiRoot(): string {
  return `${backendAddress()}/api`;
}

// Per-operation timeouts the old components set on their calls; the
// default is the client's.
export const timeouts = {
  default: 60_000,
  import: 120_000,
  /** latency tests, saving a node, updating GFWList: no limit */
  none: 0,
} as const;

// A session is one token on one backend. Requests carry the session's
// number; when the session is reset (logout, another backend) the ones in
// flight are aborted and a response that still arrives is refused, so it
// cannot write into the new session's state.
let sessionNumber = 0;
const inFlight = new Set<AbortController>();

export function currentSession(): number {
  return sessionNumber;
}

/** abortSession aborts every request in flight and starts a new session number. */
export function abortSession(): void {
  sessionNumber++;
  for (const c of inFlight) c.abort();
  inFlight.clear();
}

type Hooks = {
  /** A 401 on a request that is not the login or registration call. */
  onUnauthorized?: () => void;
  /** Any failure other than a cancellation; the app decides how to show it. */
  onError?: (err: ApiError) => void;
  /** The backend answered: whatever said it was unreachable can go. */
  onReached?: () => void;
};
const hooks: Hooks = {};

export function setClientHooks(h: Hooks): void {
  Object.assign(hooks, h);
}

export const client = axios.create({
  timeout: timeouts.default,
  // The backend reads object query parameters (touch, whiches) as JSON
  // text, which is how axios 0.21 serialized them; later versions expand
  // them into touch[id]=… instead.
  paramsSerializer: (params) =>
    Object.entries(params as Record<string, unknown>)
      .filter(([, v]) => v !== undefined && v !== null)
      .map(
        ([k, v]) =>
          `${encodeURIComponent(k)}=${encodeURIComponent(typeof v === "object" ? JSON.stringify(v) : String(v))}`,
      )
      .join("&"),
});

client.interceptors.request.use((config) => {
  const url = config.url ?? "";
  // Only the backend that issued the token gets it: the address dialog
  // probes a user-typed URL with /api/version and that must go bare.
  const token = localStorage.getItem("token");
  if (token !== null && url.startsWith(apiRoot())) {
    config.headers = {
      ...config.headers,
      Authorization: token,
      "X-V2raya-Request-Id": nanoid(),
    };
  }
  // Every request gets the session's controller, so a reset aborts it; a
  // caller's own signal (the connect watcher) is joined to it.
  const controller = new AbortController();
  inFlight.add(controller);
  const own = config.signal as AbortSignal | undefined;
  config.signal = own
    ? AbortSignal.any([own, controller.signal])
    : controller.signal;
  (config as SessionConfig).sessionController = controller;
  (config as SessionConfig).session = sessionNumber;
  return config;
});

interface SessionConfig extends AxiosRequestConfig {
  session?: number;
  sessionController?: AbortController;
}

function settle(config: SessionConfig | undefined) {
  if (config?.sessionController) inFlight.delete(config.sessionController);
}

client.interceptors.response.use(
  (res: AxiosResponse) => {
    settle(res.config as SessionConfig);
    if ((res.config as SessionConfig).session !== sessionNumber) {
      throw new ApiError(
        "stale",
        "response from a previous session",
        res.config.url ?? "",
      );
    }
    hooks.onReached?.();
    return res;
  },
  (raw: unknown) => {
    // axios's isCancel is a type guard to Cancel, which would narrow the
    // error to never in the branches below; keep the error untyped here.
    const err = raw as AxiosError<ApiEnvelope<unknown>>;
    const config = err.config as SessionConfig | undefined;
    settle(config);
    const url = config?.url ?? "";
    const cancelled = axios.isCancel(raw) || err.code === "ERR_CANCELED";
    let apiErr: ApiError;
    if (cancelled) {
      apiErr = new ApiError("cancelled", "cancelled", url);
    } else if (config && config.session !== sessionNumber) {
      apiErr = new ApiError("stale", "response from a previous session", url);
    } else if (err.code === "ECONNABORTED") {
      apiErr = new ApiError("timeout", err.message, url);
    } else if (err.response) {
      apiErr = new ApiError("http", err.message, url, {
        status: err.response.status,
        body: err.response.data,
      });
      const isAuthAction =
        url.includes("/api/login") || url.includes("/api/account");
      if (err.response.status === 401 && !isAuthAction) {
        hooks.onUnauthorized?.();
      }
    } else if (location.protocol === "https:" && /^http:\/\//i.test(url)) {
      // an https page cannot reach an http backend; the browser blocks it
      apiErr = new ApiError("mixed-content", err.message, url);
    } else {
      apiErr = new ApiError("network", err.message, url);
    }
    if (apiErr.kind !== "cancelled" && apiErr.kind !== "stale")
      hooks.onError?.(apiErr);
    return Promise.reject(apiErr);
  },
);

// how many calls are in flight; the shell's progress bar reads it
const activity = reactive({ inFlight: 0 });
export const requestActivity = readonly(activity);

/** call performs one backend operation and unwraps the envelope; a FAIL envelope is an ApiError of kind http with status 200. */
// The backend serves one state-changing request at a time and answers
// REQUEST_IN_PROGRESS to a read that arrives meanwhile (a touch during a
// latency test, say). A read waits and asks again a few times before the
// page hears of it.
const busyRetries = [400, 800, 1600];

export async function call<T>(
  config: AxiosRequestConfig & { url: string },
): Promise<T> {
  const url =
    config.url.startsWith("http") || config.url.startsWith("/")
      ? config.url
      : `${apiRoot()}/${config.url}`;
  const isRead = (config.method ?? "get").toLowerCase() === "get";
  for (let attempt = 0; ; attempt++) {
    activity.inFlight++;
    let res;
    try {
      res = await client.request<ApiEnvelope<T>>({ ...config, url });
    } finally {
      activity.inFlight--;
    }
    const body = res.data;
    if (body?.code === "SUCCESS") return body.data;
    if (
      isRead &&
      body?.errorCode === "REQUEST_IN_PROGRESS" &&
      attempt < busyRetries.length
    ) {
      await new Promise((r) => setTimeout(r, busyRetries[attempt]));
      continue;
    }
    throw new ApiError("http", body?.message ?? "request failed", config.url, {
      status: res.status,
      body,
    });
  }
}

/** probe asks another backend for its version without credentials; the address dialog uses it. */
export async function probe(
  address: string,
  timeout = 5_000,
): Promise<ApiEnvelope<{ version: string }>> {
  const res = await axios.get<ApiEnvelope<{ version: string }>>(
    `${address}/api/version`,
    { timeout },
  );
  return res.data;
}
