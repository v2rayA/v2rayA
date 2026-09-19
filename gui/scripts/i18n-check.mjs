// i18n-check: every locale carries the same key set as en, and every key
// the sources reference exists in en. Dynamic keys (built at runtime from
// a backend error code, a log level, an export mode) are declared by
// prefix here rather than found by scanning.
//
//   node scripts/i18n-check.mjs   (exit 1 on any missing key)
import {
  readdirSync,
  readFileSync,
  writeFileSync,
  mkdtempSync,
  rmSync,
} from "node:fs";
import { join, dirname, resolve } from "node:path";
import { tmpdir } from "node:os";
import { fileURLToPath, pathToFileURL } from "node:url";

const root = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const localeDir = join(root, "src", "locales");
const srcDir = join(root, "src");

// Prefixes whose keys are chosen at runtime; the checker only demands that
// the prefix exists in en.
const dynamicPrefixes = [
  "backend.",
  "log.",
  "operations.export",
  "setting.options.",
];

// The locale files are ESM without a package "type"; load them through a
// temporary .mjs copy so Node parses them as modules.
async function loadLocale(file) {
  const tmp = mkdtempSync(join(tmpdir(), "i18n-"));
  const copy = join(tmp, "m.mjs");
  writeFileSync(copy, readFileSync(join(localeDir, file), "utf8"));
  try {
    return (await import(pathToFileURL(copy).href)).default;
  } finally {
    rmSync(tmp, { recursive: true, force: true });
  }
}

function flatten(obj, prefix = "", out = new Map()) {
  for (const [k, v] of Object.entries(obj)) {
    const key = prefix ? `${prefix}.${k}` : k;
    if (v && typeof v === "object") flatten(v, key, out);
    else out.set(key, v);
  }
  return out;
}

function* walk(dir) {
  for (const e of readdirSync(dir, { withFileTypes: true })) {
    const p = join(dir, e.name);
    if (e.isDirectory()) {
      if (e.name !== "locales") yield* walk(p);
    } else if (/\.(vue|js|ts)$/.test(e.name)) yield p;
  }
}

const files = readdirSync(localeDir).filter(
  (f) => /^[a-z-]+\.js$/.test(f) && f !== "index.js",
);
const locales = Object.fromEntries(
  await Promise.all(files.map(async (f) => [f, flatten(await loadLocale(f))])),
);
const en = locales["en.js"];
let failed = false;

for (const [file, keys] of Object.entries(locales)) {
  if (file === "en.js") continue;
  const missing = [...en.keys()].filter((k) => !keys.has(k));
  const extra = [...keys.keys()].filter((k) => !en.has(k));
  if (missing.length || extra.length) {
    failed = true;
    console.log(`${file}: ${missing.length} missing, ${extra.length} extra`);
    for (const k of missing.slice(0, 20)) console.log(`  missing ${k}`);
    for (const k of extra.slice(0, 20)) console.log(`  extra   ${k}`);
  }
}

// Literal keys in $t("…"), $tc("…"), t("…"), i18n.global.t("…"), i18n.global.te("…").
const keyRe = /(?:\$t|\$tc|\bt|\.te|\.t)\(\s*["'`]([a-zA-Z0-9_.-]+)["'`]/g;
const used = new Map();
for (const file of walk(srcDir)) {
  const text = readFileSync(file, "utf8");
  for (const m of text.matchAll(keyRe)) {
    const key = m[1];
    if (!used.has(key)) used.set(key, file.slice(root.length + 1));
  }
}
const unknown = [...used].filter(
  ([k]) =>
    !en.has(k) &&
    !dynamicPrefixes.some(
      (p) => k.startsWith(p) && [...en.keys()].some((e) => e.startsWith(p)),
    ),
);
if (unknown.length) {
  failed = true;
  console.log(`${unknown.length} referenced keys missing from en:`);
  for (const [k, f] of unknown.slice(0, 40)) console.log(`  ${k}  (${f})`);
}

console.log(
  `${files.length} locales, ${en.size} keys in en, ${used.size} literal keys referenced`,
);
process.exit(failed ? 1 : 0);
