import { computed, ref, watch } from "vue";
import { useI18n } from "vue-i18n";
import {
  deleteTouch,
  getSharingAddress,
  putOutboundConnections,
  putOutboundSelection,
} from "@/api";
import { errorText } from "@/api/errors";
import { copyText } from "@/lib/clipboard";
import type { Touch, TouchSubscription, Which } from "@/api/types";
import {
  useConfirm,
  useDialog,
  useNotify,
  useOutboundGroups,
} from "@/composables";
import ImportDialog from "@/dialogs/Import.vue";
import OutboundGroupDialog from "@/dialogs/OutboundGroup.vue";
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
  "test" | "select" | "membership" | "edit" | "share" | "delete";
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
  const groups = useOutboundGroups();
  const query = ref("");
  const source = ref("all");
  const membersOnly = ref(false);
  const view = ref(
    localStorage.getItem("proxiesView") === "list" ? "list" : "cards",
  );
  const loading = ref(true);
  const loadError = ref("");
  const busy = ref(false);
  const testing = ref(false);
  const manualIntent = ref<string | null>(null);
  const selectedKeys = ref<string[]>([]);
  const subscriptions = computed(() => nodes.touch.value.subscriptions);
  const rows = computed(() => [
    ...nodes.touch.value.servers,
    ...subscriptions.value.flatMap((s) => s.servers),
  ]);
  const members = computed(() =>
    groupMembers(nodes.touch.value, store.connectedServer, store.outboundName),
  );
  /** every group with its member count; the page lists them as chips */
  const groupList = computed(() =>
    store.outbounds.map((name) => ({
      name,
      count: store.connectedServer.filter(
        (w) => (w.outbound ?? "proxy") === name,
      ).length,
    })),
  );
  const selectedMember = computed(() =>
    store.connectedServer.find(
      (w) => (w.outbound ?? "proxy") === store.outboundName && w.selected,
    ),
  );
  const mode = computed(() =>
    selectedMember.value || manualIntent.value === store.outboundName
      ? "manual"
      : "auto",
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
      rows.value.filter((row) => {
        const matchesSource =
          source.value === "all" ||
          (source.value === "local"
            ? row._type === "server"
            : subscriptions.value[row.sub ?? -1]?.address === source.value);
        return matchesSource && (!membersOnly.value || isMember(row));
      }),
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
      manualIntent.value = null;
      selectedKeys.value = [];
    },
  );
  watch(sources, (value) => {
    if (!value.some((s) => s.value === source.value)) source.value = "all";
  });

  function isMember(row: Row) {
    return nodes.inGroup(row, store.outboundName);
  }
  function isSelected(row: Row) {
    return (
      !!selectedMember.value && sameWhich(selectedMember.value, whichOf(row))
    );
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
  async function toggleGroup(row: Row, group = store.outboundName) {
    await run(() => nodes.toggleGroup(row, group));
  }
  async function selectMember(row: Row | null) {
    if (row && !isMember(row)) return;
    await run(async () => {
      nodes.apply(
        await putOutboundSelection({
          outbound: store.outboundName,
          which: row ? whichOf(row) : null,
        }),
      );
      manualIntent.value = null;
    });
  }
  async function setMode(value: string) {
    if (value === "manual") manualIntent.value = store.outboundName;
    else if (selectedMember.value) await selectMember(null);
    else manualIntent.value = null;
  }
  async function batchMembership(add: boolean) {
    const selectedRows = selected.value;
    if (!selectedRows.length) return;
    await run(async () => {
      const outbound = store.outboundName;
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
  async function exportSelected() {
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
      await copyText(links.join("\n"));
      notify.success(t("operations.copySelectedDone"));
    });
  }
  /** the splitting rules, edited in place */
  /** the group's balancing: probe URL, interval and strategy */
  function groupSettings() {
    open(OutboundGroupDialog, { outbound: store.outboundName }, { width: 440 });
  }
  async function removeGroup() {
    await run(async () => {
      if (await groups.remove(store.outboundName)) await sync();
    });
  }
  async function newGroup() {
    await run(async () => {
      if (await groups.add()) await sync();
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
  async function nodeAction(row: Row, action: NodeAction) {
    if (action === "test") return testRows([row]);
    if (action === "membership") return toggleGroup(row);
    if (action === "select") return selectMember(isSelected(row) ? null : row);
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
    membersOnly,
    view,
    loading,
    loadError,
    busy,
    testing,
    rows,
    subscriptions,
    members,
    selectedMember,
    mode,
    listed,
    selected,
    selectedKeys,
    allSelected,
    canDelete,
    preferred,
    isMember,
    isSelected,
    inUse,
    sourceName,
    selectAll,
    selectRow,
    sync,
    toggleGroup,
    selectMember,
    setMode,
    batchMembership,
    testRows,
    testListed,
    removeRows,
    exportSelected,
    newGroup,
    removeGroup,
    groupSettings,
    groupList,
    newNode,
    importNodes,
    subscriptionSettings,
    nodeAction,
    subscriptionAction,
  };
}
