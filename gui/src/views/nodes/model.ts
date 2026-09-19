// Shared node state and backend operations for the dashboard and proxies page.
import { computed, ref, watch } from "vue";
import dayjs from "dayjs";
import {
  getHttpLatency,
  getPingLatency,
  getTouch,
  putOutboundConnections,
} from "@/api";
import type {
  Touch,
  TouchResponse,
  TouchServer,
  TouchSubscription,
  Which,
} from "@/api/types";
import { openLoading } from "@/composables/useLoading";
import { useAppStore, type Running } from "@/stores/app";

export type Row = TouchServer;
/** what a table row may be: a server, a subscription's server, or a subscription */
export type Selectable = TouchServer | TouchSubscription;
export function runningOf(running: boolean, networkPaused: boolean): Running {
  if (networkPaused) return "paused";
  return running ? "running" : "stopped";
}

/** whichOf addresses a row the way the backend does. */
export function whichOf(row: Selectable): Which {
  return row._type === "subscriptionServer"
    ? { _type: row._type, id: row.id, sub: (row as Row).sub }
    : { _type: row._type, id: row.id };
}

export function sameWhich(a: Which, b: Which): boolean {
  return (
    a._type === b._type &&
    a.id === b.id &&
    (a._type !== "subscriptionServer" || a.sub === b.sub)
  );
}

/** rowKey identifies a row across refreshes: a subscription update reorders nodes and reuses ids, so the key is what the row points at. */
export function rowKey(row: Selectable): string {
  if (row._type === "subscription")
    return `subscription|${row.id}|${row.address}`;
  const r = row as Row;
  return `${r._type}|${r.sub ?? -1}|${r.address}|${r.name}|${r.net}`;
}

const latencyOf = (row: Row) => parseInt(row.pingLatency);

/** compareLatency sorts numbers ascending and puts untested rows last, whatever the direction. */
export function compareLatency(a: Row, b: Row, asc = true): number {
  const x = latencyOf(a);
  const y = latencyOf(b);
  if (isNaN(x)) return 1;
  if (isNaN(y)) return -1;
  return asc ? x - y : y - x;
}

/** compareConnection puts connected rows first, then by latency. */
export function compareConnection(a: Row, b: Row, asc = true): number {
  if (a.connected && !b.connected) return -1;
  if (!a.connected && b.connected) return 1;
  return compareLatency(a, b, asc);
}

export function filterRows(rows: Row[], query: string): Row[] {
  const search = query.trim().toLowerCase();
  if (!search) return rows;
  return rows.filter(
    (row) =>
      row.name.toLowerCase().includes(search) ||
      row.address.toLowerCase().includes(search) ||
      row.net.toLowerCase().includes(search),
  );
}

export function useNodes() {
  const store = useAppStore();
  const touch = ref<Touch>({
    servers: [],
    subscriptions: [],
    connectedServer: [],
  });
  const ready = ref(false);

  const connected = computed<Which[]>(() => touch.value.connectedServer ?? []);
  /** the rows connected in the current outbound */
  const connectedRows = computed(() =>
    connected.value
      .filter((w) => (w.outbound ?? "proxy") === store.outboundName)
      .map((w) => ({ which: w, row: locate(touch.value, w) }))
      .filter((x): x is { which: Which; row: Row } => x.row !== null),
  );
  function apply(res: TouchResponse) {
    const next = res.touch;
    next.subscriptions.forEach((s, i) => {
      s.status = dayjs(s.status)
        .tz(dayjs.tz.guess())
        .format("YYYY-MM-DD HH:mm:ss");
      s.servers.forEach((v) => {
        v.sub = i;
        v.connected = false;
      });
    });
    next.servers.forEach((v) => (v.connected = false));
    for (const w of next.connectedServer ?? []) {
      if ((w.outbound ?? "proxy") !== store.outboundName) continue;
      const row = locate(next, w);
      if (row) row.connected = true;
    }
    touch.value = next;
    store.setRunning(
      runningOf(res.running, !!res.networkPaused),
      !!res.networkPaused,
    );
    store.connectedServer = next.connectedServer ?? [];
  }

  // the connected marks depend on the outbound in view
  watch(
    () => store.outboundName,
    () => {
      for (const row of allRows()) row.connected = false;
      for (const { row } of connectedRows.value) row.connected = true;
    },
  );

  function allRows(): Row[] {
    return [
      ...touch.value.servers,
      ...touch.value.subscriptions.flatMap((s) => s.servers),
    ];
  }

  /** sync reloads the touch; the socket's open and every mutation call it. */
  async function sync(): Promise<void> {
    apply(await getTouch());
    ready.value = true;
  }

  function inGroup(row: Row, group: string): boolean {
    const w = whichOf(row);
    return connected.value.some(
      (c) => (c.outbound ?? "proxy") === group && sameWhich(c, w),
    );
  }

  /** toggleGroup adds the row to the group's members, or removes it. */
  async function toggleGroup(row: Row, group: string): Promise<void> {
    const w = whichOf(row);
    const members = connected.value
      .filter((c) => (c.outbound ?? "proxy") === group)
      .map((c) => ({
        id: c.id,
        _type: c._type,
        sub: c._type === "subscriptionServer" ? c.sub : 0,
        outbound: group,
      }));
    const next = inGroup(row, group)
      ? members.filter((m) => !sameWhich(m, w))
      : members.concat([
          { id: w.id, _type: w._type, sub: w.sub ?? 0, outbound: group },
        ]);
    const loading = openLoading();
    try {
      apply(await putOutboundConnections({ outbound: group, touches: next }));
    } finally {
      loading.close();
    }
  }

  async function testAll(
    rows: Row[],
    http: boolean,
    testingText: string,
  ): Promise<void> {
    const whiches = rows.map(
      (r) => ({ ...whichOf(r), sub: r.sub ?? null }) as Which,
    );
    rows.forEach((r) => (r.pingLatency = testingText));
    try {
      const res = await (http
        ? getHttpLatency(whiches)
        : getPingLatency(whiches));
      for (const w of res.whiches) {
        const row = locate(touch.value, w);
        if (row) row.pingLatency = w.pingLatency ?? "";
      }
    } catch (err) {
      rows.forEach((r) => (r.pingLatency = ""));
      throw err;
    }
  }

  return {
    touch,
    ready,
    connectedRows,
    sync,
    inGroup,
    toggleGroup,
    testAll,
    apply,
  };
}

/** locate finds the row a Which points at. */
export function locate(touch: Touch, which: Which): Row | null {
  if (which._type === "server") return touch.servers[which.id - 1] ?? null;
  if (which._type === "subscriptionServer")
    return touch.subscriptions[which.sub ?? -1]?.servers[which.id - 1] ?? null;
  return null;
}
