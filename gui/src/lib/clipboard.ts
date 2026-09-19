// Copying text: the Clipboard API where the page is a secure context, the
// selection-and-execCommand path elsewhere (v2rayA is usually served over
// plain http on a LAN address, where navigator.clipboard is undefined).
export async function copyText(text: string): Promise<void> {
  if (navigator.clipboard && window.isSecureContext) {
    await navigator.clipboard.writeText(text);
    return;
  }
  const area = document.createElement("textarea");
  area.value = text;
  area.setAttribute("readonly", "");
  area.style.position = "fixed";
  area.style.opacity = "0";
  document.body.appendChild(area);
  area.select();
  try {
    if (!document.execCommand("copy")) throw new Error("copy command failed");
  } finally {
    area.remove();
  }
}
