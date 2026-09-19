import { reactive, readonly } from "vue";

// A full-page overlay while something runs. Several callers may hold one
// at once; the overlay stays until the last handle closes.
const state = reactive({ open: new Set<number>() });
let nextId = 1;

export function openLoading(): { close(): void } {
  const id = nextId++;
  state.open.add(id);
  return { close: () => state.open.delete(id) };
}

/** closeAllLoadings drops every handle; the session reset calls it. */
export function closeAllLoadings(): void {
  state.open.clear();
}

export const loadingState = readonly(state);

export function useLoading() {
  return { open: openLoading };
}
