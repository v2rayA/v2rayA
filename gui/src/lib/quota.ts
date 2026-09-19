// The subscription's quota, read back from the text the backend writes
// (server/service/subscription.go, SubscriptionUserInfo.String): one of
// "Used X GiB / Y GiB", "Used X GiB", "Total Y GiB", each optionally
// followed by " · Expires YYYY-MM-DD". The GUI shows a bar only when both
// sides are known.
export interface Quota {
  /** the used amount as the backend wrote it, "" when unknown */
  used: string;
  /** the total as the backend wrote it, "" when unknown */
  total: string;
  expires: string;
  /** 0–100 when both sides are known, else undefined */
  percent?: number;
}

const units: Record<string, number> = {
  B: 1,
  KiB: 1024,
  MiB: 1024 ** 2,
  GiB: 1024 ** 3,
  TiB: 1024 ** 4,
};

const bytes = (amount: string) => {
  const m = /^([\d.]+) (\w+)$/.exec(amount);
  return m ? Number(m[1]) * (units[m[2]] ?? 1) : NaN;
};

/** parseQuota reads the backend's text; null when it says nothing about the quota. */
export function parseQuota(info: string | undefined): Quota | null {
  if (!info) return null;
  const expires = /Expires (\S+)/.exec(info)?.[1] ?? "";
  const both = /Used ([\d.]+ \w+) \/ ([\d.]+ \w+)/.exec(info);
  const usedOnly = /Used ([\d.]+ \w+)/.exec(info);
  const totalOnly = /Total ([\d.]+ \w+)/.exec(info);
  if (both) {
    const total = bytes(both[2]);
    return {
      used: both[1],
      total: both[2],
      expires,
      percent:
        total > 0
          ? Math.min(100, Number(((bytes(both[1]) / total) * 100).toFixed(1)))
          : undefined,
    };
  }
  if (usedOnly) return { used: usedOnly[1], total: "", expires };
  if (totalOnly) return { used: "", total: totalOnly[1], expires };
  return expires ? { used: "", total: "", expires } : null;
}
