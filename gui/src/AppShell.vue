<script setup lang="ts">
// The page, laid out the Material 3 way: a navigation rail from 600 dp
// (a bottom navigation bar below), a top app bar with the page's title,
// the core's state chip and the account, theme and language menus, and
// the current page's pane under v-main; the hosts for notices, dialogs
// and the loading overlay. The shell also runs the session: it is the
// starter resetSession() calls.
import {
  computed,
  defineAsyncComponent,
  onBeforeUnmount,
  onMounted,
  ref,
  watch,
  watchEffect,
} from "vue";
import { useI18n } from "vue-i18n";
import { useDisplay, useLocale, useTheme } from "vuetify";
import dayjs from "dayjs";
import {
  mdiBookOpenPageVariantOutline,
  mdiDotsVertical,
  mdiPower,
} from "@mdi/js";
import {
  deleteV2ray,
  getAccount,
  getOutbounds,
  getVersion,
  postV2ray,
  getTouch,
} from "@/api";
import { ApiError, backendAddress, currentSession } from "@/api/client";
import { watchConnected } from "@/api/connect";
import { errorText } from "@/api/errors";
import type {
  ObservatoryMessage,
  RunningStateMessage,
  TrafficMessage,
  WsMessage,
} from "@/api/types";
import { installClientHooks } from "@/clientHooks";
import {
  createMessageSocket,
  openDialog,
  openLoading,
  useBanner,
  useNotify,
  useTraffic,
} from "@/composables";
import { useDialog } from "@/composables/useDialog";
import BannerHost from "@/components/hosts/BannerHost.vue";
import DialogHost from "@/components/hosts/DialogHost.vue";
import LoadingHost from "@/components/hosts/LoadingHost.vue";
import NoticeHost from "@/components/hosts/NoticeHost.vue";
import NavBar from "@/components/NavBar.vue";
import NavDrawer from "@/components/NavDrawer.vue";
import NavRail from "@/components/NavRail.vue";
import ShellMenus from "@/components/ShellMenus.vue";
import { destinations } from "@/components/destinations";
import { isSection } from "@/docs";
import { languages } from "@/components/languages";
import LoginDialog from "@/dialogs/Login.vue";
import OnboardingDialog, {
  shouldShowOnboarding,
} from "@/dialogs/Onboarding.vue";
import { onSessionTeardown, setSessionStarter } from "@/session";
import { setRefresher } from "@/session/refresh";
import { useAppStore, type Running } from "@/stores/app";
import { runningOf } from "@/views/nodes/model";
import { vuetifyLocales } from "@/theme";
import { schemeColors } from "@/theme/scheme";
import logo from "@/assets/img/v2raya-icon.svg";
import OutboundMenu from "@/components/OutboundMenu.vue";
import PortsDialog from "@/dialogs/settings/Ports.vue";
import AboutView from "@/views/AboutView.vue";
import DashboardView from "@/views/DashboardView.vue";
import LogsView from "@/views/LogsView.vue";
import ProxiesView from "@/views/ProxiesView.vue";
import SettingsView from "@/views/SettingsView.vue";
import SubscriptionsView from "@/views/SubscriptionsView.vue";
// the docs and their Markdown load only when the page is opened
const DocsView = defineAsyncComponent(() => import("@/views/DocsView.vue"));

const store = useAppStore();
const { t, locale } = useI18n();
const notify = useNotify();
const banner = useBanner();
const traffic = useTraffic();
const theme = useTheme();
const vuetifyLocale = useLocale();
// Material's window size classes: compact < 600, medium < 840, expanded
const { width } = useDisplay();
const compact = computed(() => width.value < 600);
// Material's window classes: compact < 600 (bottom bar), medium and
// expanded < 1200 (rail with an app bar), large ≥ 1200 (standard drawer)
const expanded = computed(() => width.value >= 1200);
// the drawer folded to the rail, remembered
const folded = ref(localStorage.getItem("drawer") === "rail");
watch(folded, (v) => localStorage.setItem("drawer", v ? "rail" : "open"));
const pageTitle = computed(() =>
  t(destinations.find((d) => d.view === store.view)?.label ?? "common.about"),
);

// ---- the page -------------------------------------------------------------------

// the rendered view; the shell calls its sync when the socket reopens
const pageRef = ref<{ sync?(): Promise<void> } | null>(null);
// a new session gets a new page
const sessionSerial = ref(0);

function labelOf(running: Running): string {
  return t(
    {
      running: "common.isRunning",
      stopped: "common.notRunning",
      paused: "common.waitingNetwork",
      checking: "common.checkRunning",
    }[running],
  );
}

// ---- the session -------------------------------------------------------------

function applyTitle() {
  const address = store.backendAddress;
  const relative =
    !address || (address.startsWith("/") && !address.startsWith("//"));
  let host = location.host;
  if (!relative) {
    try {
      host = new URL(address).host;
    } catch {
      host = address;
    }
  }
  document.title = `v2rayA - ${host}`;
}

// No token: ask whether an account exists, with a few retries in case the
// backend is still coming up, then show login or registration.
async function askForLogin() {
  const session = currentSession();
  for (let attempt = 0; ; attempt++) {
    try {
      const { hasAnyAccounts } = await getAccount();
      if (session !== currentSession()) return;
      openDialog(
        LoginDialog,
        { first: !hasAnyAccounts },
        { persistent: true, width: 420 },
      );
      return;
    } catch (err) {
      if (session !== currentSession()) return;
      const transient =
        err instanceof ApiError &&
        (err.kind === "network" || err.kind === "timeout");
      if (transient && attempt < 3) {
        await new Promise((r) => setTimeout(r, 2000));
        continue;
      }
      // a backend that is still not there leaves nothing on the page
      // otherwise: the banner carries the retry
      banner.show({
        key: "login",
        kind: "warning",
        text: t("axios.messages.noBackendFound", { url: backendAddress() }),
        action: {
          label: t("operations.refresh"),
          onClick: () => {
            banner.withdraw("login");
            void askForLogin();
          },
        },
      });
      return;
    }
  }
}

async function announceVersion() {
  const v = await getVersion();
  store.applyVersion(v);
  // the running notice once per browser session, not on every reload
  const seenKey = "welcomeShown:" + v.version;
  let seen = false;
  try {
    seen = sessionStorage.getItem(seenKey) === "1";
    sessionStorage.setItem(seenKey, "1");
  } catch {
    // storage unavailable: show it
  }
  if (!seen)
    notify.info(
      t(v.docker ? "welcome.docker" : "welcome.default", {
        version: v.version,
      }),
    );
  // what stays true stays on screen: a banner, not a toast
  if (v.foundNew)
    banner.show({
      key: "newVersion",
      kind: "info",
      text: t("welcome.newVersion", { version: v.remoteVersion }),
      action: {
        label: "GitHub",
        onClick: () =>
          window.open("https://github.com/v2rayA/v2rayA/releases", "_blank"),
      },
    });
  if (v.coreVersionValid === false)
    banner.show({
      key: "coreVersion",
      kind: "error",
      text: t("version.coreVersionMismatch", { err: v.coreVersionErr || "" }),
      dismissible: false,
    });
  else if (v.serviceValid === false)
    banner.show({
      key: "serviceInvalid",
      kind: "error",
      text: t("version.v2rayInvalid"),
    });
}

function onMessage(msg: WsMessage) {
  if (msg.type === "observatory") {
    const { body } = msg as ObservatoryMessage;
    if (body?.outboundName)
      store.observatory[body.outboundName] = body.outboundStatus ?? [];
  } else if (msg.type === "traffic") {
    traffic.feed(msg as TrafficMessage);
  } else if (msg.type === "running_state") {
    const { body } = msg as RunningStateMessage;
    if (!body) return;
    const paused = !!body.networkPaused;
    store.setRunning(
      paused ? "paused" : body.running ? "running" : "stopped",
      paused,
    );
  }
}

async function startSession() {
  onSessionTeardown(() => traffic.reset());
  sessionSerial.value++;
  applyTitle();
  if (!store.loggedIn) {
    await askForLogin();
    return;
  }
  void announceVersion().catch(() => {
    // the client hooks announce an unreachable backend
  });
  void getOutbounds()
    .then((r) => {
      store.setOutbounds(r.outbounds);
      if (shouldShowOnboarding())
        openDialog(OnboardingDialog, {}, { width: 560 });
    })
    .catch(() => {
      // an expired token: the 401 hook resets the session
    });
  const socket = createMessageSocket({
    onMessage,
    // messages are not replayed: every open re-syncs the state
    onOpen: () => void pageRef.value?.sync?.(),
  });
  onSessionTeardown(() => socket.stop());
  socket.start();
}

setSessionStarter(startSession);
setRefresher(() => pageRef.value?.sync?.());
installClientHooks({ openAddressDialog: openPorts });

// ---- the core's state ---------------------------------------------------------

const hovering = ref(false);
const statusColor = computed(
  () =>
    ({
      running: "primary",
      stopped: "secondary",
      paused: "tertiary",
      checking: "surface-variant",
    })[store.running],
);
const statusText = computed(() => {
  if (hovering.value && store.running === "running") return t("v2ray.stop");
  if (hovering.value && store.running === "stopped") return t("v2ray.start");
  return labelOf(store.running);
});

/** every text the status button can carry, for its fixed width; the
 * "waiting for network" label is long and rare, so it widens the button
 * while it shows instead of reserving its width all the time */
const statusLabels = computed(() => {
  const labels = (["running", "stopped", "checking"] as Running[]).map(labelOf);
  labels.push(t("v2ray.stop"), t("v2ray.start"));
  if (store.running === "paused") labels.push(labelOf("paused"));
  return [...new Set(labels)];
});

const toggling = ref(false);
async function toggleRunning() {
  if (toggling.value) return;
  toggling.value = true;
  try {
    if (store.running === "stopped" || store.running === "paused") {
      const loading = openLoading();
      const control = new AbortController();
      try {
        const res = await watchConnected(
          postV2ray({ signal: control.signal }),
          () => control.abort(),
          {
            onCheckFailed: (err) =>
              notify.warning(
                t("connection.checkFailed", { message: errorText(err) }),
              ),
          },
        );
        // the watcher may win the race; the confirmed state comes from a touch
        const touch = res ?? (await getTouch());
        store.setRunning(
          runningOf(touch.running, touch.networkPaused),
          touch.networkPaused,
        );
        store.connectedServer = touch.touch.connectedServer ?? [];
        void pageRef.value?.sync?.();
      } catch (err) {
        notify.warning(t("v2ray.startFailed", { message: errorText(err) }));
      } finally {
        loading.close();
      }
    } else if (store.running === "running") {
      try {
        const res = await deleteV2ray();
        store.setRunning("stopped");
        store.connectedServer = res.touch.connectedServer ?? [];
        void pageRef.value?.sync?.();
      } catch (err) {
        notify.warning(t("v2ray.stopFailed", { message: errorText(err) }));
      }
    }
  } finally {
    toggling.value = false;
  }
}

// ---- the address dialog -------------------------------------------------------

function openPorts() {
  useDialog().open(PortsDialog, {}, { width: 520 });
}

// ---- theme and language ---------------------------------------------------------

watchEffect(() => {
  // both palettes follow the seed, so switching appearance later is instant
  theme.themes.value.light.colors = schemeColors(store.themeSeed, false);
  theme.themes.value.dark.colors = schemeColors(store.themeSeed, true);
});
watchEffect(() => {
  theme.global.name.value =
    store.themePreference === "auto" ? "system" : store.themePreference;
});

watch(
  locale,
  (flag) => {
    vuetifyLocale.current.value = vuetifyLocales[flag] ?? "en";
    dayjs.locale(languages.find((l) => l.flag === flag)?.dayjs ?? flag);
    document.documentElement.lang = flag;
    document.documentElement.dir = vuetifyLocale.isRtl.value ? "rtl" : "ltr";
  },
  { immediate: true },
);

// "#docs" or "#docs/<section>" opens the documentation: the help links in
// the dialogs and the About page point there, in this tab or a new one.
function openHash() {
  const [, page, section = ""] =
    location.hash.match(/^#(docs)(?:\/([\w-]*))?$/) ?? [];
  if (page !== "docs") return;
  store.docsSection = isSection(section) ? section : "";
  store.view = "docs";
  history.replaceState(null, "", location.pathname + location.search);
}
onMounted(() => {
  window.addEventListener("hashchange", openHash);
  openHash();
  void startSession();
});
onBeforeUnmount(() => window.removeEventListener("hashchange", openHash));
</script>

<template>
  <v-app>
    <NavDrawer v-if="expanded && !folded" @fold="folded = true" />
    <NavRail
      v-else-if="!compact"
      :foldable="expanded"
      @unfold="folded = false"
    />

    <v-app-bar
      v-if="!expanded"
      :height="64"
      color="surface"
      scroll-behavior="elevate"
    >
      <template v-if="compact" #prepend>
        <img :src="logo" alt="v2rayA" class="bar__logo ms-2" />
      </template>
      <v-app-bar-title class="md3-title-large" :class="{ bar__brand: compact }">
        {{ compact ? "v2rayA" : pageTitle }}
      </v-app-bar-title>
      <v-btn
        :color="statusColor"
        variant="tonal"
        :prepend-icon="mdiPower"
        height="40"
        class="text-none"
        :class="compact ? 'me-1' : 'me-2'"
        :disabled="toggling"
        @mouseenter="hovering = true"
        @mouseleave="hovering = false"
        @click="toggleRunning"
      >
        <!-- every label the button can show is laid out in the same cell, so
             the width is the widest of them and the buttons beside it do not
             move when 就绪 becomes 正在运行 or the hover text takes over -->
        <span class="bar__status">
          <span
            v-for="label in statusLabels"
            :key="label"
            :class="{ 'bar__status-label--hidden': label !== statusText }"
            >{{ label }}</span
          >
        </span>
      </v-btn>
      <OutboundMenu
        :variant="compact ? 'icon' : 'chip'"
        @changed="pageRef?.sync?.()"
      />
      <template #append>
        <ShellMenus v-if="!compact" variant="icons" />
        <v-menu v-else :close-on-content-click="false">
          <template #activator="{ props: menu }">
            <v-btn
              v-bind="menu"
              :icon="mdiDotsVertical"
              variant="text"
              class="me-1"
              :aria-label="t('common.menu')"
            />
          </template>
          <v-list density="compact" min-width="240" class="pa-2">
            <v-list-item
              :prepend-icon="mdiBookOpenPageVariantOutline"
              :title="t('common.docs')"
              :active="store.view === 'docs'"
              rounded="xl"
              @click="store.view = 'docs'"
            />
            <ShellMenus variant="list" />
          </v-list>
        </v-menu>
      </template>
    </v-app-bar>

    <NavBar v-if="compact" />

    <v-main>
      <div
        class="page"
        :class="{
          'page--wide': ['dashboard', 'proxies', 'subscriptions'].includes(
            store.view,
          ),
        }"
      >
        <div v-if="expanded" class="page__header">
          <h1 class="md3-headline-medium page__title">{{ pageTitle }}</h1>
          <div class="d-flex align-center ga-2">
            <v-btn
              :color="statusColor"
              variant="tonal"
              :prepend-icon="mdiPower"
              class="text-none"
              :disabled="toggling"
              @mouseenter="hovering = true"
              @mouseleave="hovering = false"
              @click="toggleRunning"
            >
              {{ statusText }}
            </v-btn>
            <OutboundMenu variant="chip" @changed="pageRef?.sync?.()" />
            <ShellMenus variant="icons" />
          </div>
        </div>
        <BannerHost />
        <DashboardView
          v-if="store.view === 'dashboard'"
          ref="pageRef"
          :key="sessionSerial"
        />
        <ProxiesView
          v-else-if="store.view === 'proxies'"
          ref="pageRef"
          :key="sessionSerial"
        />
        <SubscriptionsView
          v-else-if="store.view === 'subscriptions'"
          ref="pageRef"
          :key="sessionSerial"
        />
        <SettingsView
          v-else-if="store.view === 'settings'"
          ref="pageRef"
          :key="sessionSerial"
        />
        <LogsView
          v-else-if="store.view === 'logs'"
          ref="pageRef"
          :key="sessionSerial"
        />
        <DocsView v-else-if="store.view === 'docs'" :key="sessionSerial" />
        <AboutView
          v-else-if="store.view === 'about'"
          ref="pageRef"
          :key="sessionSerial"
        />
      </div>
    </v-main>

    <NoticeHost />
    <DialogHost />
    <LoadingHost />
  </v-app>
</template>

<style scoped>
.bar__status {
  display: inline-grid;
}
.bar__status > span {
  grid-area: 1 / 1;
  text-align: center;
}
.bar__status-label--hidden {
  visibility: hidden;
}
.bar__logo {
  width: 32px;
  height: 32px;
}
/* the wordmark next to the logo: medium weight, like the drawer's brand */
.bar__brand {
  font-weight: 500;
  letter-spacing: 0;
}
/* Material's margins: 16 dp on compact, 24 dp from medium; readable width */
.page {
  padding: 16px;
  margin: 0 auto;
  max-width: 1040px;
}
.page--wide {
  max-width: 1400px;
}
.page__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin: 0 0 24px;
}
.page__title {
  margin: 0;
}
@media (min-width: 600px) {
  .page {
    padding: 24px;
  }
}
</style>
