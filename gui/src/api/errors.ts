// The text a failed operation shows. The backend answers with an
// errorCode and params that the locale's backend.* messages know, or with
// a message of its own; a transport failure has the client's message.
// Plain text: the notice renders it as text, so nothing is escaped here.
import i18n from "@/plugins/i18n";
import { ApiError } from "./client";

export function errorText(err: unknown): string {
  const t = i18n.global.t;
  if (!(err instanceof ApiError)) {
    return err instanceof Error && err.message ? err.message : t("common.fail");
  }
  const body = err.body;
  if (body) {
    const code = body.errorCode;
    if (code && i18n.global.te("backend." + code)) {
      return t("backend." + code, {
        ...(body.params ?? {}),
        message: body.message,
      });
    }
    if (body.message) return body.message;
  }
  return err.message || t("common.fail");
}
