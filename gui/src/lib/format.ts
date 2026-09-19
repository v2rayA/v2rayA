const units = ["B", "KiB", "MiB", "GiB"] as const;

/** Formats a nonnegative byte count with one decimal, capped at GiB. */
export function formatBytes(n: number): string {
  let unit = 0;
  while (n >= 1024 && unit < units.length - 1) {
    n /= 1024;
    unit++;
  }
  return `${n.toFixed(1)} ${units[unit]}`;
}

/** Formats bytes per second without locale-specific separators. */
export function formatRate(n: number): string {
  return `${formatBytes(n)}/s`;
}
