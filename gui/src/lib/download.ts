// Saving text as a file: an object URL behind a synthetic link, so the page
// needs no server round-trip and works on the plain-http LAN address too.
import dayjs from "dayjs";

/**
 * exportName is what a node export is called: "export-" and the moment it was
 * taken, down to the second, in a form every file system accepts.
 */
export function exportName(at: Date = new Date()): string {
  return `export-${dayjs(at).format("YYYY-MM-DD_HH-mm-ss")}.txt`;
}

/** saveText offers `text` to the browser to keep under `name` */
export function saveText(name: string, text: string): void {
  const url = URL.createObjectURL(new Blob([text], { type: "text/plain" }));
  const link = document.createElement("a");
  link.href = url;
  link.download = name;
  document.body.appendChild(link);
  link.click();
  link.remove();
  // the download reads the URL after the click returns
  setTimeout(() => URL.revokeObjectURL(url));
}
