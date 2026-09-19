// "Refresh" for a notice that asks for one (the backend is still busy
// with the previous request): the shell registers the current page's
// sync; without one the page reloads.
let refresher: (() => Promise<void> | void) | null = null;

export function setRefresher(fn: (() => Promise<void> | void) | null): void {
  refresher = fn;
}

export function refresh(): void {
  if (refresher) void refresher();
  else location.reload();
}
