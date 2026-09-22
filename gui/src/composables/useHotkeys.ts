// Page-level keyboard shortcuts: a window keydown listener for the
// component's life, silent while a dialog or a menu is up or the user is
// typing; only Ctrl/Cmd+S reaches the page from a text field.
import { onBeforeUnmount, onMounted } from "vue";
import { dialogState } from "./useDialog";

export function useHotkeys(handler: (event: KeyboardEvent) => void) {
  function onKey(event: KeyboardEvent) {
    if (event.isComposing || dialogState.stack.length) return;
    if (document.querySelector(".v-overlay--active.v-menu")) return;
    const target = event.target instanceof HTMLElement ? event.target : null;
    const typing = !!target?.closest(
      'input:not([type="checkbox"]):not([type="radio"]), textarea, [contenteditable]',
    );
    const save =
      (event.ctrlKey || event.metaKey) && event.key.toLowerCase() === "s";
    if (typing && !save) return;
    handler(event);
  }
  onMounted(() => window.addEventListener("keydown", onKey));
  onBeforeUnmount(() => window.removeEventListener("keydown", onKey));
}

/** focus the first text input under a selector */
export function focusInput(selector: string) {
  document.querySelector<HTMLInputElement>(`${selector} input`)?.focus();
}
