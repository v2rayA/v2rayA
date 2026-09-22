import { computed, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import { deleteTouch, getSharingAddress, putOutboundConnections } from "@/api";
import { errorText } from "@/api/errors";
import { copyText } from "@/lib/clipboard";
import { exportName, saveText } from "@/lib/download";
import type { Touch, TouchSubscription, Which } from "@/api/types";
import { useConfirm, useDialog, useNotify } from "@/composables";
import ImportDialog from "@/dialogs/Import.vue";
import ServerDialog from "@/dialogs/Server/index.vue";
import SharingDialog from "@/dialogs/Sharing.vue";
import SubscriptionDialog from "@/dialogs/Subscription.vue";
import SubscriptionSettings from "@/dialogs/SubscriptionSettings.vue";
import { useAppStore } from "@/stores/app";
import {
  filterRows,
  locate,
  rowKey,
  sameWhich,
  useNodes,
  whichOf,
  type Row,
} from "../nodes/model";
import { useSubscriptions } from "../subscriptions/model";

export type NodeAction =
  "test" | "addToGroup" | "removeFromGroup" | "edit" | "share" | "delete";
export type SubscriptionAction = "update" | "edit" | "share" | "delete";

export function groupMembers(
  touch: Touch,
  connected: Which[],
  group: string,
): Row[] {
  return connected
    .filter((w) => (w.outbound ?? "proxy") === group)
    .map((w) => locate(touch, w))
    .filter((row): row is Row => row !== null);
}

export function useProxies() {
  const store = useAppStore();
  const nodes = useNodes();
  const subscriptionsModel = useSubscriptions();
  const { t } = useI18n();
  const notify = useNotify();
  const confirm = useConfirm();
  const { open } = useDialog();
  const query = ref("");
  const source = ref("all");
  const view = ref(
    localStorage.getItem("proxiesView") === "list" ? "list" : "cards",
  );
  const loading = ref(true);
  const loadError = ref("");
  const busy = ref(false);
  const testing = ref(false);
  const selectedKeys = ref<string[]>([]);
  const subscriptions = computed(() => nodes.touch.value.subscriptions);
  const rows = computed(() => [
    ...nodes.touch.value.servers,
    ...subscriptions.value.flatMap((s) => s.servers),
  ]);
  const members = computed(() =>
    groupMembers(nodes.touch.value, store.connectedServer, store.outboundName),
  );
  const sources = computed(() => [
    { value: "all", title: t("proxies.sources.all") },
    { value: "local", title: t("proxies.sources.local") },
    ...subscriptions.value.map((s) => ({
      value: s.address,
      title: s.remarks || s.host,
    })),
  ]);
  const listed = computed(() =>
    filterRows(
      rows.value.filter(
        (row) =>
          source.value === "all" ||
          (source.value === "local"
            ? row._type === "server"
            : subscriptions.value[row.sub ?? -1]?.address === source.value),
      ),
      query.value ?? "",
    ),
  );
  const selected = computed(() =>
    listed.value.filter((row) => selectedKeys.value.includes(rowKey(row))),
  );
  const allSelected = computed(
    () =>
      listed.value.length > 0 && selected.value.length === listed.value.length,
  );
  const canDelete = computed(
    () =>
      selected.value.length > 0 &&
      selected.value.every((r) => r._type === "server"),
  );
  const preferred = computed(() => inUse(store.outboundName));

  watch(view, (value) => localStorage.setItem("proxiesView", value));
  watch(listed, (value) => {
    const keys = new Set(value.map(rowKey));
    selectedKeys.value = selectedKeys.value.filter((key) => keys.has(key));
  });
  watch(
    () => store.outboundName,
    () => {
      selectedKeys.value = [];
    },
  );
  watch(sources, (value) => {
    if (!value.some((s) => s.value === source.value)) source.value = "all";
  });

  function isMember(row: Row) {
    return nodes.inGroup(row, store.outboundName);
  }
  /** the groups the row belongs to, in the order the app bar lists them */
  function memberGroups(row: Row) {
    return store.outbounds.filter((group) => nodes.inGroup(row, group));
  }
  function inUse(group: string): Row | null {
    const connected = store.connectedServer.filter(
      (w) => (w.outbound ?? "proxy") === group,
    );
    const chosen = connected.find((w) => w.selected);
    if (chosen) return locate(nodes.touch.value, chosen);
    let best: Which | null = null;
    let delay = Infinity;
    for (const status of store.observatory[group] ?? []) {
      if (
        status.alive &&
        status.delay < delay &&
        connected.some((w) => sameWhich(w, status.which))
      ) {
        best = status.which;
        delay = status.delay;
      }
    }
    return best ? locate(nodes.touch.value, best) : null;
  }
  function sourceName(row: Row) {
    const subscription = subscriptions.value[row.sub ?? -1];
    return row._type === "server"
      ? t("proxies.sources.local")
      : subscription?.remarks || subscription?.host || "";
  }
  function selectAll(value: boolean) {
    selectedKeys.value = value ? listed.value.map(rowKey) : [];
  }
  function selectRow(row: Row, value: boolean) {
    const key = rowKey(row);
    selectedKeys.value = value
      ? [...new Set([...selectedKeys.value, key])]
      : selectedKeys.value.filter((k) => k !== key);
  }
  async function run(action: () => Promise<unknown>) {
    if (busy.value) return;
    busy.value = true;
    try {
      await action();
    } catch (err) {
      notify.warning(errorText(err));
    } finally {
      busy.value = false;
    }
  }
  async function sync() {
    loading.value = true;
    loadError.value = "";
    try {
      await nodes.sync();
    } catch (err) {
      loadError.value = errorText(err);
    } finally {
      loading.value = false;
    }
  }
  /** setMembership adds or removes the row, leaving the other members alone */
  async function setMembership(row: Row, group: string, member: boolean) {
    if (nodes.inGroup(row, group) === member) return;
    await run(() => nodes.toggleGroup(row, group));
  }
  async function batchMembership(add: boolean, outbound = store.outboundName) {
    const selectedRows = selected.value;
    if (!selectedRows.length) return;
    await run(async () => {
      const touches: Which[] = store.connectedServer
        .filter((w) => (w.outbound ?? "proxy") === outbound)
        .map(({ id, _type, sub }) => ({ id, _type, sub }));
      const next = add
        ? [...touches]
        : touches.filter(
            (w) => !selectedRows.some((row) => sameWhich(w, whichOf(row))),
          );
      if (add)
        for (const row of selectedRows) {
          const which = whichOf(row);
          if (!next.some((w) => sameWhich(w, which))) next.push(which);
        }
      nodes.apply(await putOutboundConnections({ outbound, touches: next }));
    });
  }
  async function testRows(targets: Row[], http = false) {
    if (testing.value || !targets.length) return;
    testing.value = true;
    try {
      await nodes.testAll(targets, http, t("latency.testing"));
    } catch (err) {
      notify.warning(t("latency.failed", { message: errorText(err) }));
    } finally {
      testing.value = false;
    }
  }
  const testListed = (http = false) => testRows(listed.value, http);
  async function removeRows(targets: Row[]) {
    if (!targets.length || targets.some((r) => r._type !== "server")) return;
    await run(async () => {
      if (
        !(await confirm({
          title: t("delete.title"),
          message: t("delete.message", targets.length),
          confirmText: t("operations.delete"),
          destructive: true,
        }))
      )
        return;
      await deleteTouch(targets.map(whichOf));
      selectedKeys.value = [];
      await sync();
    });
  }
  /** exportSelected copies the selected nodes' links, or saves them as a file */
  async function exportSelected(where: "clipboard" | "file" = "clipboard") {
    const targets = [...selected.value];
    await run(async () => {
      const links = (
        await Promise.all(targets.map((row) => getSharingAddress(whichOf(row))))
      )
        .map((res) => res.sharingAddress)
        .filter(Boolean);
      if (!links.length) {
        notify.warning(t("operations.exportEmpty"));
        return;
      }
      if (where === "file") {
        const file = exportName();
        saveText(file, links.join("\n"));
        notify.success(t("operations.exportSaved", { file }));
        return;
      }
      await copyText(links.join("\n"));
      notify.success(t("operations.copySelectedDone"));
    });
  }
  async function newNode() {
    await run(async () => {
      if (
        await open<boolean>(ServerDialog, { which: null }, { width: 560 })
          .result
      )
        await sync();
    });
  }
  async function importNodes(kind: "server" | "subscription" = "server") {
    await run(async () => {
      if (await open<boolean>(ImportDialog, { kind }, { width: 480 }).result)
        await sync();
    });
  }
  const subscriptionSettings = () =>
    open(SubscriptionSettings, {}, { width: 560 });
  async function nodeAction(
    row: Row,
    action: NodeAction,
    group = store.outboundName,
  ) {
    if (action === "test") return testRows([row]);
    if (action === "addToGroup") return setMembership(row, group, true);
    if (action === "removeFromGroup") return setMembership(row, group, false);
    if (action === "delete") return removeRows([row]);
    await run(async () => {
      if (action === "edit") {
        if (
          await open<boolean>(
            ServerDialog,
            { which: whichOf(row), readonly: row._type !== "server" },
            { width: 560 },
          ).result
        )
          await sync();
      } else {
        const { sharingAddress: link } = await getSharingAddress(whichOf(row));
        open(
          SharingDialog,
          {
            title: t("sharing.serverTitle"),
            link,
            name: row.name,
            type: row._type,
          },
          { width: 420 },
        );
      }
    });
  }
  async function subscriptionAction(
    subscription: TouchSubscription,
    action: SubscriptionAction,
  ) {
    await run(async () => {
      if (action === "share") {
        const link = await subscriptionsModel.sharingLink(subscription);
        open(
          SharingDialog,
          {
            title: t("sharing.subscriptionTitle"),
            link,
            name: subscription.remarks || subscription.host,
            type: subscription._type,
          },
          { width: 420 },
        );
        return;
      }
      if (action === "update") {
        await subscriptionsModel.update(subscription);
        notify.success(t("subscription.updated"));
      }
      if (
        action === "edit" &&
        !(await open<boolean>(
          SubscriptionDialog,
          { subscription },
          { width: 480 },
        ).result)
      )
        return;
      if (action === "delete") {
        if (
          !(await confirm({
            title: t("delete.title"),
            message: t("delete.message", 1),
            confirmText: t("operations.delete"),
            destructive: true,
          }))
        )
          return;
        await subscriptionsModel.remove(subscription);
      }
      await sync();
    });
  }
  return {
    store,
    nodes,
    query,
    source,
    sources,
    view,
    loading,
    loadError,
    busy,
    testing,
    rows,
    subscriptions,
    members,
    listed,
    selected,
    selectedKeys,
    allSelected,
    canDelete,
    preferred,
    isMember,
    memberGroups,
    inUse,
    sourceName,
    selectAll,
    selectRow,
    sync,
    setMembership,
    batchMembership,
    testRows,
    testListed,
    removeRows,
    exportSelected,
    newNode,
    importNodes,
    subscriptionSettings,
    nodeAction,
    subscriptionAction,
  };
}
