// Material 3 colour roles from one seed colour, the way the Material theme
// builder derives them: the seed's hue and chroma in HCT space give the
// primary, secondary, tertiary and neutral tonal palettes, and each role
// is a fixed tone of one palette, in a light and a dark version. The
// scheme is "tonal spot", Material's default. Only the seed is the user's
// choice; success, warning and info are the primary, error and tertiary roles.
import {
  argbFromHex,
  Hct,
  hexFromArgb,
  MaterialDynamicColors,
  SchemeMonochrome,
  SchemeTonalSpot,
} from "@material/material-color-utilities";

/** The default seed: Google yellow. */
export const brandSeed = "#fbbc04";

export const presetSeeds: { name: string; seed: string }[] = [
  { name: "brand", seed: brandSeed },
  { name: "teal", seed: "#13aab7" },
  { name: "blue", seed: "#4a6fd5" },
  { name: "green", seed: "#3d8f5a" },
  { name: "purple", seed: "#7b57c7" },
  { name: "red", seed: "#c8493f" },
];

// role → Vuetify colour key
const roles = {
  primary: "primary",
  onPrimary: "on-primary",
  primaryContainer: "primary-container",
  onPrimaryContainer: "on-primary-container",
  inversePrimary: "inverse-primary",
  secondary: "secondary",
  onSecondary: "on-secondary",
  secondaryContainer: "secondary-container",
  onSecondaryContainer: "on-secondary-container",
  tertiary: "tertiary",
  onTertiary: "on-tertiary",
  tertiaryContainer: "tertiary-container",
  onTertiaryContainer: "on-tertiary-container",
  error: "error",
  onError: "on-error",
  errorContainer: "error-container",
  onErrorContainer: "on-error-container",
  background: "background",
  onBackground: "on-background",
  surface: "surface",
  onSurface: "on-surface",
  surfaceDim: "surface-dim",
  surfaceBright: "surface-bright",
  surfaceContainerLowest: "surface-container-lowest",
  surfaceContainerLow: "surface-container-low",
  surfaceContainer: "surface-container",
  surfaceContainerHigh: "surface-container-high",
  surfaceContainerHighest: "surface-container-highest",
  surfaceVariant: "surface-variant",
  onSurfaceVariant: "on-surface-variant",
  inverseSurface: "inverse-surface",
  inverseOnSurface: "inverse-on-surface",
  outline: "outline",
  outlineVariant: "outline-variant",
  surfaceTint: "surface-tint",
} as const;

// Vuetify's success / warning / info keys, which v-alert's `type` uses,
// follow the scheme: success is the primary, info the tertiary, warning
// the error role, each with its "on" colour
function statusColors(colors: Record<string, string>) {
  return {
    success: colors.primary,
    "on-success": colors["on-primary"],
    warning: colors.error,
    "on-warning": colors["on-error"],
    info: colors.tertiary,
    "on-info": colors["on-tertiary"],
  };
}

/** isSeed accepts a #rrggbb colour. */
/** seedFromHue gives a vivid seed at a hue (0–360): HCT chroma 48, tone 60, as the theme builder's wheel does. */
export function seedFromHue(hue: number): string {
  return hexFromArgb(Hct.from(((hue % 360) + 360) % 360, 48, 60).toInt());
}

/** hueOf reads a seed's HCT hue, for the slider's position. */
export function hueOf(seed: string): number {
  return isSeed(seed) ? Hct.fromInt(argbFromHex(seed)).hue : 0;
}

export function isSeed(value: string): boolean {
  return /^#[0-9a-f]{6}$/i.test(value);
}

/** The colour set of one theme, complete: Vuetify's required keys and every Material role. */
export type SchemeColors = Record<string, string> &
  Record<
    | "background"
    | "surface"
    | "primary"
    | "secondary"
    | "success"
    | "warning"
    | "error"
    | "info"
    | "on-background"
    | "on-surface"
    | "on-primary"
    | "on-secondary"
    | "on-success"
    | "on-warning"
    | "on-error"
    | "on-info",
    string
  >;

/** schemeColors gives the Vuetify colour set for one seed in one brightness. */
export function schemeColors(seed: string, dark: boolean): SchemeColors {
  const hct = Hct.fromInt(argbFromHex(isSeed(seed) ? seed : brandSeed));
  // a grey, black or white seed has no hue to speak of: tonal spot would
  // paint it pink from hue 0, so those get the monochrome scheme instead
  const scheme =
    hct.chroma < 5
      ? new SchemeMonochrome(hct, dark, 0)
      : new SchemeTonalSpot(hct, dark, 0);
  const colors: Record<string, string> = {};
  for (const [role, key] of Object.entries(roles)) {
    const color = MaterialDynamicColors[role as keyof typeof roles];
    colors[key] = hexFromArgb(color.getArgb(scheme));
  }
  return { ...colors, ...statusColors(colors) } as SchemeColors;
}
