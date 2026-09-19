import { computed, onMounted, ref, shallowRef } from "vue";
import { useI18n } from "vue-i18n";
import dayjs from "dayjs";
import "@/plugins/dayjs";
import {
  deleteV2ray,
  getPingLatency,
  getTouch,
  postV2ray,
  putOutboundConnections,
  putOutboundSelection,
  putSubscription,
} from "@/api";
import { watchConnected } from "@/api/connect";
import { errorText } from "@/api/errors";
import type {
  TouchSubscription,
  Touch,
  TouchResponse,
  TouchServer,
  Which,
} from "@/api/types";
import { openLoading, useConfirm, useDialog, useNotify } from "@/composables";
import ImportDialog from "@/dialogs/Import.vue";
import SharingDialog from "@/dialogs/Sharing.vue";
import SubscriptionDialog from "@/dialogs/Subscription.vue";
import PortsDialog from "@/dialogs/settings/Ports.vue";
import RoutingADialog from "@/dialogs/settings/RoutingA.vue";
import GroupMembersDialog from "./GroupMembers.vue";
import { useAppStore } from "@/stores/app";
import { locate, runningOf, sameWhich } from "@/views/nodes/model";
import { parseQuota } from "@/lib/quota";
import { useSettings, type SettingForm } from "@/views/settings/model";
import type { SubscriptionAction } from "@/views/proxies/model";
import { useSubscriptions } from "@/views/subscriptions/model";

export interface DashboardMember {
  key: string;
  which: Which;
  row: TouchServer;
  alive: boolean;
  latency: string;
  /** the observatory's delay */
  delay?: number;
  /** the number behind `latency`: the observatory's delay, else a measured ping; undefined when untested or timed out */
  ms?: number;
}

/** parseMs reads "157ms" / "157 ms"; anything else (TIMEOUT, empty) is undefined. */
function parseMs(text: string): number | undefined {
  const match = /^(\d+)\s*ms$/i.exec(text.trim());
  return match ? Number(match[1]) : undefined;
}

// The backend formats both quota values in GiB.
export function useDashboard() {
  const store = useAppStore();
  const { t, locale } = useI18n();
  const notify = useNotify();
  const settings = useSettings();
  const touch = shallowRef<Touch>();
  const loading = ref(true);
  const busy = ref(false);
  const error = ref("");
  const quickLoading = ref(true);
  const quickSaving = ref(false);
  const selecting = ref(false);
  const testing = ref<string>();
  const measured = ref(new Map<string, string>());
  const updating = ref<number>();
  const updatingAll = ref(false);
  const members = computed<DashboardMember[]>(() =>
    store.connectedServer
      .filter((which) => (which.outbound ?? "proxy") === store.outboundName)
      .flatMap((which) => {
        const row = touch.value && locate(touch.value, which);
        if (!row) return [];
        const status = store.observatory[store.outboundName]?.find((entry) =>
          sameWhich(entry.which, which),
        );
        const delay =
          status && Number.isFinite(status.delay) && status.delay >= 0
            ? status.delay
            : undefined;
        const key = JSON.stringify([
          store.outboundName,
          which._type,
          which.sub,
          which.id,
          row.address,
          row.name,
        ]);
        return [
          {
            key,
            which,
            row,
            delay,
            alive: status?.alive ?? false,
            latency:
              measured.value.get(key) ??
              (delay !== undefined ? `${delay} ms` : row.pingLatency || "—"),
            ms:
              delay ??
              parseMs(measured.value.get(key) ?? row.pingLatency ?? ""),
          },
        ];
      }),
  );
  const nodeInUse = computed(() => {
    const pinned = members.value.find((member) => member.which.selected);
    if (pinned) return pinned;
    let best: DashboardMember | undefined;
    for (const member of members.value) {
      if (
        member.alive &&
        member.delay !== undefined &&
        (best?.delay === undefined || member.delay < best.delay)
      )
        best = member;
    }
    return best ?? (members.value.length === 1 ? members.value[0] : undefined);
  });
  const subscriptions = computed(() =>
    (touch.value?.subscriptions ?? []).map((subscription) => {
      const quota = parseQuota(subscription.info);
      const date = dayjs(subscription.status);
      const dateLocale =
        locale.value === "zh"
          ? "zh-cn"
          : locale.value === "fa-ir"
            ? "fa"
            : locale.value;
      return {
        ...subscription,
        usage: quota,
        summary: quota
          ? [
              quota.used && quota.total
                ? t("dashboard.usage", { used: quota.used, total: quota.total })
                : quota.used
                  ? t("dashboard.usedOnly", { used: quota.used })
                  : quota.total
                    ? t("dashboard.totalOnly", { total: quota.total })
                    : "",
              quota.expires
                ? t("dashboard.expires", { date: quota.expires })
                : "",
            ]
              .filter((part) => part)
              .join("  ")
          : subscription.info,
        updatedAt:
          subscription.status && date.isValid()
            ? date.locale(dateLocale).fromNow()
            : subscription.status || "—",
      };
    }),
  );
  const stateLabel = computed(() =>
    t(
      {
        running: "common.isRunning",
        stopped: "common.notRunning",
        paused: "common.waitingNetwork",
        checking: "common.checkRunning",
      }[store.running],
    ),
  );
  const canToggle = computed(
    () => !loading.value && !busy.value && store.running !== "checking",
  );
  const quickDisabled = computed(
    () => quickLoading.value || !settings.ready.value || quickSaving.value,
  );
  const subscriptionsBusy = computed(
    () => updatingAll.value || updating.value !== undefined,
  );

  function apply(response: TouchResponse) {
    touch.value = response.touch;
    store.connectedServer = response.touch.connectedServer ?? [];
    store.setRunning(
      runningOf(response.running, response.networkPaused),
      response.networkPaused,
    );
  }

  // a load failure shows as the page's alert; a toast on top would say it twice
  function report(err: unknown) {
    error.value = errorText(err);
  }

  async function loadQuick() {
    quickLoading.value = true;
    settings.ready.value = false;
    try {
      await settings.load();
    } catch (err) {
      report(err);
    } finally {
      quickLoading.value = false;
    }
  }

  const { open: openDialog } = useDialog();
  /** importNodes opens the import dialog; a subscription or link added there shows at once. */
  async function importNodes() {
    const imported = await openDialog<boolean>(ImportDialog, {}, { width: 480 })
      .result;
    if (imported) await getTouch().then(apply).catch(report);
  }
  /** editGroup lets the user pick the group's members from every node; Save replaces the list. */
  async function editGroup() {
    if (!touch.value) return;
    const outbound = store.outboundName;
    const touches = await openDialog<Which[]>(
      GroupMembersDialog,
      {
        outbound,
        touch: touch.value,
        members: store.connectedServer.filter(
          (w) => (w.outbound ?? "proxy") === outbound,
        ),
      },
      { width: 560 },
    ).result;
    if (!touches) return;
    const overlay = openLoading();
    try {
      apply(await putOutboundConnections({ outbound, touches }));
      measured.value.clear();
      void testMembers();
    } catch (err) {
      notify.warning(errorText(err));
    } finally {
      overlay.close();
    }
  }
  /** editRoutingA opens the RoutingA editor. */
  function editRoutingA() {
    openDialog(RoutingADialog, {}, { width: 960 });
  }
  /** editPorts opens the address dialog. */
  function editPorts() {
    openDialog(PortsDialog, {}, { width: 520 });
  }

  /** sync reloads the touch; the shell calls it when the socket reopens. */
  async function sync() {
    await getTouch().then(apply).catch(report);
  }

  onMounted(async () => {
    await Promise.all([sync(), loadQuick()]);
    loading.value = false;
    // the members' latency once on arrival, so the tile reads at a glance
    void testMembers();
  });

  async function setQuick<K extends keyof SettingForm>(
    key: K,
    value: SettingForm[K],
  ) {
    if (quickDisabled.value) return;
    settings.form[key] = value;
    quickSaving.value = true;
    try {
      await settings.save();
      notify.success(t("setting.saved"));
    } catch (err) {
      notify.warning(t("setting.saveFailed", { message: errorText(err) }));
      await loadQuick();
    } finally {
      quickSaving.value = false;
    }
  }

  async function selectNode(
    which: Which | null,
    outbound = store.outboundName,
  ) {
    if (selecting.value) return false;
    selecting.value = true;
    try {
      apply(await putOutboundSelection({ outbound, which }));
      return true;
    } catch (err) {
      notify.warning(errorText(err));
      return false;
    } finally {
      selecting.value = false;
    }
  }

  async function testNode() {
    const member = nodeInUse.value;
    if (!member || testing.value) return;
    testing.value = member.key;
    try {
      const result = await getPingLatency([member.which]);
      const latency = result.whiches[0]?.pingLatency;
      if (latency) measured.value.set(member.key, latency);
    } catch (err) {
      notify.warning(errorText(err));
    } finally {
      testing.value = undefined;
    }
  }

  /** testMembers pings every member of the group; the latency tile ranks the results. */
  async function testMembers() {
    if (!members.value.length || testing.value) return;
    testing.value = "all";
    try {
      const result = await getPingLatency(members.value.map((m) => m.which));
      for (const which of result.whiches) {
        const member = members.value.find((m) => sameWhich(m.which, which));
        if (member && which.pingLatency)
          measured.value.set(member.key, which.pingLatency);
      }
    } catch (err) {
      notify.warning(errorText(err));
    } finally {
      testing.value = undefined;
    }
  }

  async function refreshSubscription(id: number) {
    updating.value = id;
    try {
      apply(await putSubscription({ _type: "subscription", id }));
      measured.value.clear();
    } catch (err) {
      notify.warning(errorText(err));
    } finally {
      updating.value = undefined;
    }
  }

  const confirm = useConfirm();
  const subscriptionsModel = useSubscriptions();
  /** subscriptionAction handles the card's menu: update here, the rest through the shared model. */
  async function subscriptionAction(
    subscription: TouchSubscription,
    action: SubscriptionAction,
  ) {
    if (action === "update") return updateSubscription(subscription.id);
    try {
      if (action === "share") {
        const link = await subscriptionsModel.sharingLink(subscription);
        openDialog(
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
      if (action === "edit") {
        const saved = await openDialog<boolean>(
          SubscriptionDialog,
          { subscription },
          { width: 480 },
        ).result;
        if (!saved) return;
      }
      if (action === "delete") {
        const ok = await confirm({
          title: t("delete.title"),
          message: t("delete.message", 1),
          confirmText: t("operations.delete"),
          destructive: true,
        });
        if (!ok) return;
        await subscriptionsModel.remove(subscription);
      }
      apply(await getTouch());
    } catch (err) {
      notify.warning(errorText(err));
    }
  }

  async function updateSubscription(id: number) {
    if (subscriptionsBusy.value) return;
    await refreshSubscription(id);
  }

  async function updateAll() {
    if (subscriptionsBusy.value) return;
    updatingAll.value = true;
    try {
      for (const { id } of subscriptions.value) await refreshSubscription(id);
    } finally {
      updatingAll.value = false;
    }
  }

  async function toggleRunning() {
    if (!canToggle.value) return;
    const starting = store.running !== "running";
    busy.value = true;
    const overlay = openLoading();
    try {
      if (starting) {
        const control = new AbortController();
        const response = await watchConnected(
          postV2ray({ signal: control.signal }),
          () => control.abort(),
          {
            onCheckFailed: (err) =>
              notify.warning(
                t("connection.checkFailed", { message: errorText(err) }),
              ),
          },
        );
        // The watcher can finish before POST answers; fetch the confirmed state.
        apply(response ?? (await getTouch()));
      } else {
        apply(await deleteV2ray());
      }
    } catch (err) {
      notify.warning(
        t(starting ? "v2ray.startFailed" : "v2ray.stopFailed", {
          message: errorText(err),
        }),
      );
    } finally {
      overlay.close();
      busy.value = false;
    }
  }

  return {
    store,
    loading,
    busy,
    error,
    members,
    nodeInUse,
    editPorts,
    editRoutingA,
    editGroup,
    importNodes,
    subscriptions,
    quick: settings.form,
    quickLoading,
    quickSaving,
    quickDisabled,
    selecting,
    testing,
    updating,
    updatingAll,
    subscriptionsBusy,
    stateLabel,
    canToggle,
    toggleRunning,
    sync,
    setQuick,
    selectNode,
    testNode,
    testMembers,
    updateAll,
    updateSubscription,
    subscriptionAction,
  };
}
