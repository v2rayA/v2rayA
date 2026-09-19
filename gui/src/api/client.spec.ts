// @vitest-environment happy-dom
import { beforeEach, describe, expect, test, vi } from "vitest";
import type { AxiosRequestConfig } from "axios";
import {
  ApiError,
  abortSession,
  apiRoot,
  call,
  client,
  setClientHooks,
} from "./client";

// A fake transport: records what the client sends and answers what the
// test scripted, so the assertions are on the wire shape, not on axios.
type Sent = {
  url: string;
  method: string;
  headers: Record<string, unknown>;
  query: string;
  body: unknown;
};
const sent: Sent[] = [];
let reply: (
  config: AxiosRequestConfig,
) => Promise<{ status: number; data: unknown }> = async () => ({
  status: 200,
  data: { code: "SUCCESS", data: {} },
});

client.defaults.adapter = async (config) => {
  const serializer = config.paramsSerializer as
    | ((p: unknown) => string)
    | { serialize?: (p: unknown) => string }
    | undefined;
  const fn =
    typeof serializer === "function" ? serializer : serializer?.serialize;
  const query = fn && config.params ? fn(config.params) : "";
  sent.push({
    url: config.url ?? "",
    method: config.method ?? "get",
    headers: { ...(config.headers as Record<string, unknown>) },
    query,
    body: config.data,
  });
  const r = await reply(config);
  if (r.status >= 400) {
    const err = Object.assign(
      new Error(`Request failed with status code ${r.status}`),
      {
        config,
        response: { status: r.status, data: r.data, headers: {}, config },
        isAxiosError: true,
      },
    );
    throw err;
  }
  return {
    status: r.status,
    statusText: "OK",
    headers: {},
    config,
    data: r.data,
  };
};

beforeEach(() => {
  sent.length = 0;
  localStorage.setItem("backendAddress", "http://backend.test:2017");
  localStorage.setItem("token", "tok-1");
  reply = async () => ({
    status: 200,
    data: { code: "SUCCESS", data: { ok: 1 } },
  });
});

describe("the client sends what the old components sent", () => {
  test("token and request id only on the backend's own requests", async () => {
    await call({ url: "touch", method: "get" });
    expect(sent[0].url).toBe(`${apiRoot()}/touch`);
    expect(sent[0].headers.Authorization).toBe("tok-1");
    expect(String(sent[0].headers["X-V2raya-Request-Id"])).toMatch(
      /^[A-Za-z0-9_-]{21}$/,
    );
    await client
      .get("http://elsewhere.test/api/version")
      .catch(() => undefined);
    expect(sent[1].headers.Authorization).toBeUndefined();
    expect(sent[1].headers["X-V2raya-Request-Id"]).toBeUndefined();
  });

  test("object query parameters go as JSON text", async () => {
    await call({
      url: "sharingAddress",
      method: "get",
      params: { touch: { id: 1, _type: "server" }, n: 2, skip: undefined },
    });
    expect(sent[0].query).toBe(
      `touch=${encodeURIComponent('{"id":1,"_type":"server"}')}&n=2`,
    );
  });

  test("a FAIL envelope is an error with the backend's message", async () => {
    reply = async () => ({
      status: 200,
      data: {
        code: "FAIL",
        message: "no such node",
        errorCode: "NODE_NOT_FOUND",
        data: null,
      },
    });
    const err = (await call({ url: "touch", method: "get" }).catch(
      (e) => e,
    )) as ApiError;
    expect(err).toBeInstanceOf(ApiError);
    expect(err.kind).toBe("http");
    expect(err.body?.errorCode).toBe("NODE_NOT_FOUND");
  });

  test("a busy backend is asked again for a read, not for a write", async () => {
    vi.useFakeTimers();
    let calls = 0;
    reply = async () => {
      calls++;
      return calls < 3
        ? {
            status: 200,
            data: {
              code: "FAIL",
              message: "busy",
              errorCode: "REQUEST_IN_PROGRESS",
              data: null,
            },
          }
        : { status: 200, data: { code: "SUCCESS", data: { ok: true } } };
    };
    const read = call<{ ok: boolean }>({ url: "touch", method: "get" });
    await vi.advanceTimersByTimeAsync(400);
    await vi.advanceTimersByTimeAsync(800);
    expect(await read).toEqual({ ok: true });
    expect(calls).toBe(3);
    calls = 0;
    const err = (await call({ url: "v2ray", method: "post" }).catch(
      (e) => e,
    )) as ApiError;
    expect(err.body?.errorCode).toBe("REQUEST_IN_PROGRESS");
    expect(calls).toBe(1);
    vi.useRealTimers();
  });

  test("401 outside login and account calls the hook", async () => {
    const onUnauthorized = vi.fn();
    setClientHooks({ onUnauthorized });
    reply = async () => ({
      status: 401,
      data: { code: "FAIL", message: "unauthorized", data: null },
    });
    await call({ url: "touch", method: "get" }).catch(() => undefined);
    expect(onUnauthorized).toHaveBeenCalledTimes(1);
    await call({
      url: "login",
      method: "post",
      data: { username: "u", password: "p" },
    }).catch(() => undefined);
    expect(onUnauthorized).toHaveBeenCalledTimes(1);
    setClientHooks({ onUnauthorized: undefined });
  });

  test("a response from a previous session is refused", async () => {
    let release!: () => void;
    reply = () =>
      new Promise(
        (r) =>
          (release = () =>
            r({
              status: 200,
              data: { code: "SUCCESS", data: { late: true } },
            })),
      );
    const p = call({ url: "touch", method: "get" }).catch((e) => e as ApiError);
    await new Promise((r) => setTimeout(r, 0)); // the transport is now waiting
    abortSession();
    release();
    const err = (await p) as ApiError;
    expect(err).toBeInstanceOf(ApiError);
    expect(["stale", "cancelled"]).toContain(err.kind);
  });
});
