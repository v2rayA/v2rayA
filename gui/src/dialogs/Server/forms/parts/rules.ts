// Validation rules the protocol forms share; Vuetify runs them on the
// fields that are mounted, so a field hidden by its condition is not
// checked.
import i18n from "@/plugins/i18n";

export const required = (v: unknown) =>
  (v !== undefined && v !== null && String(v).trim() !== "") ||
  i18n.global.t("configureServer.required");

// Empty is allowed (the core default applies), but a set value must not go
// below the bound the core accepts.
export const minValue = (min: number) => (v: unknown) =>
  v === undefined ||
  v === null ||
  String(v).trim() === "" ||
  Number(v) >= min ||
  i18n.global.t("configureServer.minValue", { value: min });
