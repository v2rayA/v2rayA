// Vite plugin: serve the Lucide icon font as a virtual stylesheet that
// carries only the glyphs this interface uses.
//
// lucide-static ships 2100+ icons in five font formats; the interface uses
// a few dozen. Shipping the whole set embedded the fonts into the v2rayA
// binary (the legacy SVG font alone was 20 MB), and a typo in an icon name
// rendered nothing without anyone noticing. This plugin scans the sources
// for icon names, refuses the build on a name the font does not have, and
// subsets the woff2 to the names found, inlined as a data URL so nothing
// else has to be copied around.
import { createRequire } from "node:module";
import { readFileSync, readdirSync, statSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const require = createRequire(import.meta.url);
const subsetFont = require("subset-font");

const VIRTUAL_ID = "virtual:lucide-icons.css";
const RESOLVED_ID = "\0" + VIRTUAL_ID;

const here = dirname(fileURLToPath(import.meta.url));
const srcDir = resolve(here, "../src");
const fontDir = dirname(require.resolve("lucide-static/font/lucide.css"));

function walk(dir, out = []) {
  for (const name of readdirSync(dir)) {
    const p = join(dir, name);
    if (statSync(p).isDirectory()) walk(p, out);
    else if (/\.(vue|js|mjs|ts)$/.test(name) && !/\.spec\./.test(name)) out.push(p);
  }
  return out;
}

// Every way an icon name appears in the sources:
//   <i class="lucide icon-name">, icon-left="name", icon-right="name",
//   icon="name", icon: "name" (dialog options), and the values of the Buefy
//   pack's internalIcons map in plugins/buefy.js.
const PATTERNS = [
  // class="lucide icon-name" and 'icon-name' inside :class expressions; the
  // lookbehind keeps CSS class names such as with-icon-alert out, and the
  // exclusions are Buefy's own props (icon-left="…", icon-right-clickable).
  /(?<![\w-])icon-(?!left\b|right\b|right-click)([a-z0-9]+(?:-[a-z0-9]+)*)\b(?!=)/g,
  // a static prop only: `:icon="icon"` binds a variable, not a name
  /(?<!:)\bicon(?:-left|-right)?=["']([a-z0-9]+(?:-[a-z0-9]+)*)["']/g,
  /\bicon:\s*["'][^"']*?\b(?:icon-)?([a-z0-9]+(?:-[a-z0-9]+)*)["']/g,
];
const INTERNAL_ICONS = /internalIcons:\s*\{([^}]*)\}/;

export function collectIconNames(files) {
  const names = new Set();
  for (const file of files) {
    const text = readFileSync(file, "utf8");
    for (const re of PATTERNS) {
      re.lastIndex = 0;
      let m;
      while ((m = re.exec(text))) names.add(m[1]);
    }
    const internal = INTERNAL_ICONS.exec(text);
    if (internal) {
      for (const m of internal[1].matchAll(/:\s*"([a-z0-9]+(?:-[a-z0-9]+)*)"/g)) names.add(m[1]);
    }
  }
  // size modifiers of the pack config, not glyphs
  names.delete("md");
  names.delete("lg");
  return names;
}

export default function lucideSubset() {
  return {
    name: "lucide-subset",
    resolveId(id) {
      return id === VIRTUAL_ID ? RESOLVED_ID : null;
    },
    async load(id) {
      if (id !== RESOLVED_ID) return null;
      const codepoints = JSON.parse(readFileSync(join(fontDir, "codepoints.json"), "utf8"));
      const files = walk(srcDir);
      const wanted = collectIconNames(files);
      const used = [...wanted].filter((n) => n in codepoints).sort();
      const unknown = [...wanted].filter((n) => !(n in codepoints)).sort();
      if (unknown.length) {
        this.error(`icon names not in the Lucide font: ${unknown.join(", ")}`);
      }
      const text = used.map((n) => String.fromCodePoint(codepoints[n])).join("");
      const full = readFileSync(join(fontDir, "lucide.woff2"));
      const subset = await subsetFont(full, text, { targetFormat: "woff2" });
      const rules = used.map((n) => `.icon-${n}::before{content:"\\${codepoints[n].toString(16)}"}`).join("\n");
      const css = `@font-face{font-family:"lucide";font-display:block;src:url(data:font/woff2;base64,${subset.toString("base64")}) format("woff2")}
[class^="icon-"],[class*=" icon-"]{font-family:"lucide"!important;font-size:inherit;font-style:normal;-webkit-font-smoothing:antialiased;-moz-osx-font-smoothing:grayscale}
${rules}
`;
      this.warn(`lucide: ${used.length} icons, ${subset.length} bytes of font`);
      return css;
    },
  };
}
