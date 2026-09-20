import { writable } from "svelte/store";

function currentPath() {
  return window.location.pathname;
}

export const route = writable(currentPath());
export const routeState = writable(history.state);

function onNavigate() {
  route.set(currentPath());
  routeState.set(history.state);
}

window.addEventListener("popstate", onNavigate);

export function goto(path, state) {
  history.pushState(state ?? null, "", path);
  onNavigate();
  window.scrollTo(0, 0);
}

/** Match /decks/:id and /decks/:id/learn style patterns. */
export function matchRoute(pattern, path) {
  const pParts = pattern.split("/").filter(Boolean);
  const pathParts = path.split("/").filter(Boolean);
  if (pParts.length !== pathParts.length) return null;
  const params = {};
  for (let i = 0; i < pParts.length; i++) {
    if (pParts[i].startsWith(":")) params[pParts[i].slice(1)] = decodeURIComponent(pathParts[i]);
    else if (pParts[i] !== pathParts[i]) return null;
  }
  return params;
}
