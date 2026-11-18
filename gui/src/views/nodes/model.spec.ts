// @vitest-environment happy-dom
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { effectScope, type EffectScope } from "vue";
import "@/plugins/dayjs";

const api = vi.hoisted(() => ({
  getNodeLatency: vi.fn(),
  getHttpLatency: vi.fn(),
  putOutboundConnections: vi.fn(),
}));
vi.mock("@/api", () => api);

import {
  compareLatency,
  filterRows,
  latencyFailed,
  useNodes,
  type Row,
} from "./model";
import { useAppStore } from "@/stores/app";

function row(id: number, pingLatency = ""): Row {
  return {
    id,
    _type: "server",
    name: "Tokyo",
    address: "node.example:443",
    net: "WS",
    pingLatency,
  };
}

let scope: EffectScope;
function createNodes() {
  const nodes = scope.run(useNodes)!;
  nodes.touch.value = {
    servers: [row(1), row(2)],
    subscriptions: [
      {
        id: 1,
        _type: "subscription",
        host: "example",
        address: "https://example/sub",
        status: "2026-09-19",
        info: "",
        autoSelect: false,
        servers: [{ ...row(1), _type: "subscriptionServer", sub: 0 }],
      },
    ],
    connectedServer: [],
  };
  return nodes;
}
beforeEach(() => {
  vi.resetAllMocks();
  localStorage.clear();
  setActivePinia(createPinia());
  scope = effectScope();
});
afterEach(() => scope.stop());

describe("latency testing", () => {
  test("testAll uses the supplied rows and the global probe policy", async () => {
    const request = api.getNodeLatency;
    const nodes = createNodes();
    request.mockResolvedValue({
      whiches: [
        { _type: "server", id: 2, pingLatency: "25ms" },
        { _type: "subscriptionServer", sub: 0, id: 1, pingLatency: "8ms" },
      ],
    });
    const untouched = nodes.touch.value.servers[0];
    const rows = [
      nodes.touch.value.servers[1],
      nodes.touch.value.subscriptions[0].servers[0],
    ];
    const pending = nodes.testAll(rows, "Testing");
    expect(rows.map((r) => r.pingLatency)).toEqual(["Testing", "Testing"]);
    expect(untouched.pingLatency).toBe("");
    expect(request).toHaveBeenCalledExactlyOnceWith([
      { _type: "server", id: 2, sub: null },
      { _type: "subscriptionServer", id: 1, sub: 0 },
    ]);
    await pending;
    expect(rows.map((r) => r.pingLatency)).toEqual(["25ms", "8ms"]);
  });

  test("failed tests clear testing marks and propagate the error", async () => {
    const nodes = createNodes();
    const error = new Error("latency unavailable");
    api.getNodeLatency.mockRejectedValue(error);
    const rows = nodes.touch.value.servers;
    await expect(nodes.testAll(rows, "Testing")).rejects.toBe(error);
    expect(rows.map((r) => r.pingLatency)).toEqual(["", ""]);
  });
});

describe("a page that is gone", () => {
  test("its late touch response no longer reaches the store", () => {
    const nodes = createNodes();
    const store = useAppStore();
    const res = {
      code: "SUCCESS" as const,
      running: true,
      networkPaused: false,
      touch: {
        servers: [row(1)],
        subscriptions: [],
        connectedServer: [
          { id: 1, _type: "server" as const, outbound: "proxy" },
        ],
      },
    };
    scope.stop();
    nodes.apply(res);
    expect(store.connectedServer).toEqual([]);
    expect(nodes.touch.value.servers).toHaveLength(2);
    scope = effectScope();
  });
});

describe("filterRows", () => {
  test("matches name, address or transport without changing rows or their order", () => {
    const rows = [
      row(1),
      { ...row(2), name: "Berlin", address: "other.example:80", net: "grpc" },
    ];
    expect(filterRows(rows, "TOK")).toEqual([rows[0]]);
    expect(filterRows(rows, "NODE.EXAMPLE")).toEqual([rows[0]]);
    expect(filterRows(rows, "wS")).toEqual([rows[0]]);
    expect(filterRows(rows, "EXAMPLE")).toEqual(rows);
    expect(filterRows(rows, "missing")).toEqual([]);
    expect(filterRows(rows, "")).toBe(rows);
    expect(rows.map((r) => r.id)).toEqual([1, 2]);
  });
});

describe("latency order", () => {
  const latencies = [
    "120ms",
    "",
    "TIMEOUT",
    "35ms",
    "testing",
    "PROXY ERROR",
    "800ms",
  ];
  const rows = latencies.map((latency, i) => row(i + 1, latency));
  const sorted = (asc: boolean) =>
    [...rows]
      .sort((a, b) => compareLatency(a, b, asc))
      .map((r) => r.pingLatency);
  test("sorts measured rows both ways and keeps untested, testing and failed rows last in their order", () => {
    const rest = ["", "TIMEOUT", "testing", "PROXY ERROR"];
    expect(sorted(true)).toEqual(["35ms", "120ms", "800ms", ...rest]);
    expect(sorted(false)).toEqual(["800ms", "120ms", "35ms", ...rest]);
  });
  test("counts a finished test without a time as failed, not an untested or running one", () => {
    expect(
      rows.filter((r) => latencyFailed(r, "testing")).map((r) => r.pingLatency),
    ).toEqual(["TIMEOUT", "PROXY ERROR"]);
  });
});
