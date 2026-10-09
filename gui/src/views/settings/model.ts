// The settings page's state and its two requests, without the view: the
// form as GET /setting returns it, saved with PUT /setting in the shape
// the old dialog sent (integers where it parsed them). The DNS mode is not
// one of its fields: the DNS settings dialog owns that decision, and a copy
// held here would put a stale mode back the next time anything is saved.
import { computed, reactive, ref } from "vue";
import {
  getRemoteGFWListVersion,
  getSetting,
  getNodeDnsOptions,
  getTouch,
  putSetting,
} from "@/api";
import { watchConnected } from "@/api/connect";
import type { Setting, NodeDnsOption } from "@/api/types";
import { openLoading } from "@/composables/useLoading";
import { useAppStore } from "@/stores/app";
import { runningOf } from "@/views/nodes/model";
import { resolveDnsMode } from "@/dialogs/settings/dnsModel";

export const defaultForm = () => ({
  transparent: "close",
  transparentType: "tproxy",
  ipforward: false,
  portSharing: false,
  tproxyExcludedInterfaces: "",
  tunAutoRoute: true,
  tunRouteShellType: "",
  tunRouteShellPath: "",
  tunSetupScript: "",
  tunTeardownScript: "",
  tunExcludeProcesses: "",
  pacMode: "whitelist",
  pacAutoUpdateMode: "none",
  pacAutoUpdateIntervalHour: 0,
  subscriptionAutoUpdateMode: "none",
  subscriptionAutoUpdateIntervalHour: 0,
  proxyModeWhenSubscribe: "direct",
  tcpFastOpen: "default",
  logLevel: "info",
  inboundSniffing: "disable",
  routeOnly: false,
  muxOn: "no",
  mux: 8,
  nodeDns: "auto",
});
export type SettingForm = ReturnType<typeof defaultForm>;

/** what PUT /setting takes; the numbers are parsed, as the old dialog did */
export function toRequest(form: SettingForm): Setting {
  return {
    ...form,
    pacAutoUpdateIntervalHour: parseInt(String(form.pacAutoUpdateIntervalHour)),
    subscriptionAutoUpdateIntervalHour: parseInt(
      String(form.subscriptionAutoUpdateIntervalHour),
    ),
    mux: parseInt(String(form.mux)),
  } as Setting;
}

export function useSettings({ nodeDns = false }: { nodeDns?: boolean } = {}) {
  const store = useAppStore();
  const form = reactive(defaultForm());
  const ready = ref(false);
  const saved = ref("");
  const dirty = computed(
    () => ready.value && JSON.stringify(form) !== saved.value,
  );
  const localGFWListVersion = ref("");
  const localGeositeVersion = ref("");
  const remoteGFWListVersion = ref("");
  const nodeDnsOptions = ref<NodeDnsOption[]>([]);
  const nodeDnsWarnings = ref<string[]>([]);
  const nodeDnsError = ref<unknown>(null);
  const nodeDnsLoading = ref(false);
  const nodeDnsLoaded = ref(false);
  const nodeDnsEnabled = ref(false);
  let nodeDnsRequest = 0;

  async function loadNodeDnsOptions(refreshMode = false): Promise<void> {
    const request = ++nodeDnsRequest;
    nodeDnsLoading.value = true;
    nodeDnsError.value = null;
    try {
      const [result, setting] = await Promise.all([
        getNodeDnsOptions(),
        refreshMode ? getSetting() : null,
      ]);
      if (request !== nodeDnsRequest) return;
      nodeDnsOptions.value = result.options;
      nodeDnsWarnings.value = result.warnings ?? [];
      nodeDnsLoaded.value = true;
      if (setting)
        nodeDnsEnabled.value = resolveDnsMode(setting.setting) !== "off";
    } catch (err) {
      if (request === nodeDnsRequest) nodeDnsError.value = err;
    } finally {
      if (request === nodeDnsRequest) nodeDnsLoading.value = false;
    }
  }

  async function load(): Promise<void> {
    const res = await getSetting();
    nodeDnsEnabled.value = resolveDnsMode(res.setting) !== "off";
    for (const key of Object.keys(form) as (keyof SettingForm)[]) {
      if (key in res.setting)
        (form as Record<string, unknown>)[key] = res.setting[key];
    }
    localGFWListVersion.value = res.localGFWListVersion ?? "";
    localGeositeVersion.value = res.localGeositeVersion ?? "";
    if (store.lite) form.transparentType = "system_proxy";
    saved.value = JSON.stringify(form);
    ready.value = true;
    if (nodeDns) await loadNodeDnsOptions();
  }

  async function loadRemoteVersion(): Promise<void> {
    remoteGFWListVersion.value = (
      await getRemoteGFWListVersion()
    ).remoteGFWListVersion;
  }

  /** save applies the settings; an invalid config stops the core, which the store then shows. */
  async function save(): Promise<void> {
    const loading = openLoading();
    const control = new AbortController();
    const submitted = toRequest(form);
    const submittedForm = JSON.stringify(form);
    try {
      await watchConnected(
        putSetting(submitted, { signal: control.signal }),
        () => control.abort(),
      );
      saved.value = submittedForm;
      if (nodeDns) await loadNodeDnsOptions(true);
    } catch (err) {
      // the backend restores the previous setting and keeps the core as it
      // was; the touch says which state that is, and the form goes back to
      // what is stored so the rejected values are not resent by the next save
      await Promise.all([
        getTouch().then((res) =>
          store.setRunning(
            runningOf(res.running, !!res.networkPaused),
            !!res.networkPaused,
          ),
        ),
        load(),
      ]).catch(() => {});
      throw err;
    } finally {
      loading.close();
    }
  }

  return {
    form,
    nodeDnsOptions,
    nodeDnsWarnings,
    nodeDnsError,
    nodeDnsLoading,
    nodeDnsLoaded,
    nodeDnsEnabled,
    loadNodeDnsOptions,
    ready,
    dirty,
    localGFWListVersion,
    localGeositeVersion,
    remoteGFWListVersion,
    load,
    loadRemoteVersion,
    save,
  };
}
