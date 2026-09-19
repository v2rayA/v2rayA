import { markRaw, reactive, readonly, type Component } from "vue";

// Dialogs open programmatically with
// a component: the host renders them from this stack, the caller gets a
// handle and a promise for the result. Nested dialogs close child first.
export interface DialogEntry {
  id: number;
  component: Component;
  props: Record<string, unknown>;
  /** ESC and the scrim do nothing; only the dialog's own controls close it */
  persistent: boolean;
  /** Vuetify's max-width */
  width: number | string;
  resolve: (value: unknown) => void;
  open: boolean;
}

export interface DialogHandle<T = unknown> {
  id: number;
  close(result?: T): void;
  result: Promise<T | undefined>;
}

const state = reactive({ stack: [] as DialogEntry[] });
let nextId = 1;

export function openDialog<T = unknown>(
  component: Component,
  props: Record<string, unknown> = {},
  options: { persistent?: boolean; width?: number | string } = {},
): DialogHandle<T> {
  let resolve!: (v: unknown) => void;
  const result = new Promise<T | undefined>(
    (r) => (resolve = r as (v: unknown) => void),
  );
  const entry: DialogEntry = {
    id: nextId++,
    component: markRaw(component),
    props,
    persistent: options.persistent ?? false,
    width: options.width ?? 640,
    resolve,
    open: true,
  };
  state.stack.push(entry);
  return {
    id: entry.id,
    close: (value?: T) => closeDialog(entry.id, value),
    result,
  };
}

/** closeDialog resolves and removes one dialog; the host calls it after the closing transition too. */
export function closeDialog(id: number, result?: unknown): void {
  const i = state.stack.findIndex((d) => d.id === id);
  if (i < 0) return;
  const [entry] = state.stack.splice(i, 1);
  entry.resolve(result);
}

/** closeAllDialogs closes from the top of the stack down; the session reset calls it. */
export function closeAllDialogs(): void {
  while (state.stack.length)
    closeDialog(state.stack[state.stack.length - 1].id);
}

export const dialogState = readonly(state);

export function useDialog() {
  return { open: openDialog, close: closeDialog };
}
