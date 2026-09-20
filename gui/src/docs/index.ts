// The in-app documentation: one Markdown file per section and locale under
// src/docs/<locale>/. A locale without its own files reads the English
// ones. Files are compiled to HTML at build time and loaded on demand, so
// the docs cost nothing until the page is opened.
export const sections = [
  "quick-start",
  "transparent-proxy",
  "routing",
  "inbounds",
  "parameters",
  "troubleshooting",
] as const;
export type Section = (typeof sections)[number];

export const isSection = (value: string): value is Section =>
  (sections as readonly string[]).includes(value);

const pages = import.meta.glob<string>("./*/*.md", { import: "default" });

/** the locales with their own files; anything else falls back to English */
const written = new Set(Object.keys(pages).map((key) => key.split("/")[1]));

export function docsLocale(locale: string): string {
  return written.has(locale) ? locale : "en";
}

export async function loadSection(
  locale: string,
  section: Section,
): Promise<string> {
  const load =
    pages[`./${docsLocale(locale)}/${section}.md`] ??
    pages[`./en/${section}.md`];
  return load ? load() : "";
}
