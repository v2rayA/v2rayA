// Creating and deleting outbound groups, shared by the group menu and
// the proxies page: a prompt for the name, a confirmation before the
// delete, the store's list refreshed from the backend's answer.
import { useI18n } from "vue-i18n";
import { deleteOutbound, getOutbounds, postOutbound } from "@/api";
import { watchConnected } from "@/api/connect";
import { errorText } from "@/api/errors";
import { useAppStore } from "@/stores/app";
import { useConfirm } from "./useConfirm";
import { usePrompt } from "./useConfirm";
import { useNotify } from "./useNotify";

export function useOutboundGroups() {
  const { t } = useI18n();
  const store = useAppStore();
  const notify = useNotify();
  const prompt = usePrompt();
  const confirm = useConfirm();

  /** add asks for a name and creates the group; resolves true when one was created. */
  async function add(): Promise<boolean> {
    const outbound = await prompt({
      message: t("outbound.addMessage"),
      input: { maxlength: 10 },
    });
    if (outbound === null) return false;
    const control = new AbortController();
    try {
      const res = await watchConnected(
        postOutbound({ outbound }, { signal: control.signal }),
        () => control.abort(),
      );
      if (res) {
        notify.success(t("outbound.added"));
        store.setOutbounds((res as { outbounds: unknown }).outbounds);
      } else {
        store.setOutbounds((await getOutbounds()).outbounds);
      }
      return true;
    } catch (err) {
      notify.warning(t("outbound.addFailed", { message: errorText(err) }));
      return false;
    }
  }

  /** remove confirms and deletes the group; resolves true when it was deleted. */
  async function remove(outbound: string): Promise<boolean> {
    const ok = await confirm({
      message: t("outbound.deleteMessage", { outboundName: outbound }),
      confirmText: t("operations.delete"),
      destructive: true,
    });
    if (!ok) return false;
    try {
      const res = (await deleteOutbound({ outbound })) as {
        outbounds: unknown;
      };
      notify.success(t("outbound.deleted"));
      store.setOutbounds(res.outbounds);
      return true;
    } catch (err) {
      notify.warning(
        t("outbound.deleteFailed", {
          group: outbound,
          message: errorText(err),
        }),
      );
      return false;
    }
  }

  return { add, remove };
}
