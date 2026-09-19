// size-gate: gzip size of a build, per file and summed, first load apart
// from the rest. First load is what index.html references directly; the
// rest is loaded on demand.
//
//   node scripts/size-gate.mjs <dist dir> [--gate <KB>]
// With --gate the run fails when JS + CSS gzip of the whole build exceeds
// the limit; without it the numbers are only reported.
import { readdirSync, readFileSync, statSync } from "node:fs";
import { join, relative, extname } from "node:path";
import { gzipSync } from "node:zlib";

const [dist, ...rest] = process.argv.slice(2);
if (!dist) {
  console.error("usage: size-gate <dist dir> [--gate <KB>]");
  process.exit(2);
}
const gateIndex = rest.indexOf("--gate");
const gateKB = gateIndex >= 0 ? Number(rest[gateIndex + 1]) : null;

function* walk(dir) {
  for (const e of readdirSync(dir, { withFileTypes: true })) {
    const p = join(dir, e.name);
    if (e.isDirectory()) yield* walk(p);
    else yield p;
  }
}

const html = readFileSync(join(dist, "index.html"), "utf8");
const firstLoad = new Set(
  [...html.matchAll(/(?:src|href)="\.?\/?([^"]+\.(?:js|css))"/g)].map(
    (m) => m[1],
  ),
);

const rows = [];
for (const file of walk(dist)) {
  const ext = extname(file);
  if (ext !== ".js" && ext !== ".css") continue;
  const rel = relative(dist, file);
  if (/^(?:sw|workbox-[^/]+|registerSW)\.js$/.test(rel)) continue;
  rows.push({
    file: rel,
    ext,
    raw: statSync(file).size,
    gzip: gzipSync(readFileSync(file)).length,
    first: firstLoad.has(rel),
  });
}
rows.sort((a, b) => b.gzip - a.gzip);
const kb = (n) => (n / 1024).toFixed(1).padStart(7);
for (const r of rows)
  console.log(
    `${kb(r.gzip)} KB gzip  ${kb(r.raw)} KB raw  ${r.first ? "first" : "lazy "}  ${r.file}`,
  );
const sum = (pred) => rows.filter(pred).reduce((s, r) => s + r.gzip, 0);
const total = sum(() => true);
console.log(
  `\nJS ${kb(sum((r) => r.ext === ".js"))} KB, CSS ${kb(sum((r) => r.ext === ".css"))} KB, total ${kb(total)} KB gzip; first load ${kb(sum((r) => r.first))} KB`,
);
if (gateKB !== null && total > gateKB * 1024) {
  console.log(`over the ${gateKB} KB gate`);
  process.exit(1);
}
