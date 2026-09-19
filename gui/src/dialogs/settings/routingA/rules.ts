export interface Condition {
  fn: string;
  args: string[];
}
export type Entry = (
  | { kind: "blank" }
  | { kind: "comment"; text: string }
  | { kind: "default"; outbound: string }
  | {
      kind: "outbound";
      name: string;
      fn: string;
      args: Record<string, string>[];
    }
  | { kind: "rule"; conditions: Condition[]; outbound: string }
  | { kind: "raw"; text: string }
) & { raw?: string; original?: string };

// Delimiters inside quotes or function calls belong to the argument, not the rule.
export function splitOutside(text: string, delimiter: string): string[] | null {
  const parts: string[] = [];
  let start = 0;
  let quote = "";
  let depth = 0;
  for (let i = 0; i < text.length; i++) {
    const char = text[i];
    if (quote) {
      if (char === "\\") i++;
      else if (char === quote) quote = "";
    } else if (char === '"' || char === "'") quote = char;
    else if (char === "(") depth++;
    else if (char === ")") {
      if (--depth < 0) return null;
    } else if (depth === 0 && text.startsWith(delimiter, i)) {
      parts.push(text.slice(start, i).trim());
      i += delimiter.length - 1;
      start = i + 1;
    }
  }
  if (quote || depth) return null;
  parts.push(text.slice(start).trim());
  return parts;
}

function condition(text: string): Condition | null {
  const match = /^([a-z][\w]*)\s*\(([\s\S]*)\)$/.exec(text.trim());
  if (!match) return null;
  const args = splitOutside(match[2], ",");
  if (!args || args.some((arg) => !arg)) return null;
  return { fn: match[1], args };
}
function parseLine(text: string): Entry {
  const line = text.trim();
  if (!line) return { kind: "blank" };
  if (line.startsWith("#")) return { kind: "comment", text };
  const fallback = /^default\s*:\s*([^\s#]+)$/.exec(line);
  if (fallback) return { kind: "default", outbound: fallback[1] };
  const outbound = /^outbound\s*:\s*([^\s=]+)\s*=\s*(.*)$/.exec(line);
  if (outbound) {
    const call = condition(outbound[2]);
    if (call) {
      const args: Record<string, string>[] = [];
      for (const arg of call.args) {
        const pair = /^([a-z]+)\s*:\s*(.+)$/.exec(arg);
        if (!pair) return { kind: "raw", text };
        args.push({ [pair[1]]: pair[2] });
      }
      return { kind: "outbound", name: outbound[1], fn: call.fn, args };
    }
  }
  const sides = splitOutside(line, "->");
  if (sides?.length === 2 && /^[^\s#]+$/.test(sides[1])) {
    const calls = splitOutside(sides[0], "&&")?.map(condition);
    if (
      calls?.length &&
      calls.every((call): call is Condition => call !== null)
    )
      return { kind: "rule", conditions: calls, outbound: sides[1] };
  }
  return { kind: "raw", text };
}
function format(entry: Entry): string {
  switch (entry.kind) {
    case "blank":
      return "";
    case "comment":
    case "raw":
      return entry.text;
    case "default":
      return `default: ${entry.outbound}`;
    case "outbound":
      return `outbound: ${entry.name} = ${entry.fn}(${entry.args.flatMap((arg) => Object.entries(arg).map(([key, value]) => `${key}: ${value}`)).join(", ")})`;
    case "rule":
      return `${entry.conditions.map((call) => `${call.fn}(${call.args.join(", ")})`).join(" && ")} -> ${entry.outbound}`;
  }
}
export function parse(text: string): Entry[] {
  return text.split("\n").map((line) => {
    const entry = parseLine(line);
    return { ...entry, raw: line, original: format(entry) };
  });
}
export function serialize(entries: Entry[]): string {
  return entries
    .map((entry) => {
      const formatted = format(entry);
      return entry.raw !== undefined && entry.original === formatted
        ? entry.raw
        : formatted;
    })
    .join("\n");
}
