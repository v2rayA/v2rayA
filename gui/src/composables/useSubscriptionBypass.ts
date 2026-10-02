import { getSetting } from "@/api";
import { useAppStore } from "@/stores/app";
import { useI18n } from "vue-i18n";
import { useConfirm } from "./useConfirm";

/** Returns true for a confirmed one-request direct fetch, false for the configured route, and null on cancel. */
export function useSubscriptionBypass() {
  const store = useAppStore();
  const confirm = useConfirm();
  const { t } = useI18n();

  return async (all = false): Promise<boolean | null> => {
    if (store.running !== "stopped") return false;
    const mode = (await getSetting()).setting.proxyModeWhenSubscribe;
    if (mode !== "proxy" && mode !== "pac") return false;
    return (await confirm({
      message: all
        ? t("subscription.coreStoppedDirectAll")
        : t("subscription.coreStoppedDirect"),
      confirmText: t("operations.yes"),
      cancelText: t("operations.no"),
    }))
      ? true
      : null;
  };
}
