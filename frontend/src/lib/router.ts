import { writable } from "svelte/store";

function currentPath(): string {
  return window.location.pathname;
}

export const route = writable<string>(currentPath());
export const routeState = writable<unknown>(history.state);

function onNavigate(): void {
  route.set(currentPath());
  routeState.set(history.state);
}

window.addEventListener("popstate", onNavigate);

export function goto(path: string, state?: unknown): void {
  history.pushState(state ?? null, "", path);
  onNavigate();
  window.scrollTo(0, 0);
}

/** Match /decks/:id and /decks/:id/learn style patterns. */
export function matchRoute(pattern: string, path: string): Record<string, string> | null {
  const pParts = pattern.split("/").filter(Boolean);
  const pathParts = path.split("/").filter(Boolean);
  if (pParts.length !== pathParts.length) return null;
  const params: Record<string, string> = {};
  for (let i = 0; i < pParts.length; i++) {
    if (pParts[i].startsWith(":")) params[pParts[i].slice(1)] = decodeURIComponent(pathParts[i]);
    else if (pParts[i] !== pathParts[i]) return null;
  }
  return params;
}
