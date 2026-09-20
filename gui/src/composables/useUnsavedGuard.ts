import { onBeforeUnmount, onMounted } from "vue";

// The browser asks before a reload or a closed tab would drop unsaved edits.
export function useUnsavedGuard(dirty: () => boolean) {
  const guard = (event: BeforeUnloadEvent) => {
    if (dirty()) event.preventDefault();
  };
  onMounted(() => window.addEventListener("beforeunload", guard));
  onBeforeUnmount(() => window.removeEventListener("beforeunload", guard));
}
