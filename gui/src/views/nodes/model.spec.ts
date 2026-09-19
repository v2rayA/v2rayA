// @vitest-environment happy-dom
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { createPinia, setActivePinia } from "pinia";
import { effectScope, type EffectScope } from "vue";
import "@/plugins/dayjs";

const api = vi.hoisted(() => ({
  getPingLatency: vi.fn(),
  getHttpLatency: vi.fn(),
  putOutboundConnections: vi.fn(),
}));
vi.mock("@/api", () => api);

import { filterRows, useNodes, type Row } from "./model";

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
  test.each([false, true])(
    "testAll uses the supplied rows (HTTP: %s)",
    async (http) => {
      const request = http ? api.getHttpLatency : api.getPingLatency;
      const nodes = createNodes();
      const other = http ? api.getPingLatency : api.getHttpLatency;
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
      const pending = nodes.testAll(rows, http, "Testing");
      expect(rows.map((r) => r.pingLatency)).toEqual(["Testing", "Testing"]);
      expect(untouched.pingLatency).toBe("");
      expect(request).toHaveBeenCalledExactlyOnceWith([
        { _type: "server", id: 2, sub: null },
        { _type: "subscriptionServer", id: 1, sub: 0 },
      ]);
      expect(other).not.toHaveBeenCalled();
      await pending;
      expect(rows.map((r) => r.pingLatency)).toEqual(["25ms", "8ms"]);
    },
  );

  test("failed tests clear testing marks and propagate the error", async () => {
    const nodes = createNodes();
    const error = new Error("latency unavailable");
    api.getPingLatency.mockRejectedValue(error);
    const rows = nodes.touch.value.servers;
    await expect(nodes.testAll(rows, false, "Testing")).rejects.toBe(error);
    expect(rows.map((r) => r.pingLatency)).toEqual(["", ""]);
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
