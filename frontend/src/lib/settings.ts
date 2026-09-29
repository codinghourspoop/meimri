import { get, writable } from "svelte/store";
import { setMode } from "mode-watcher";
import { api, type Settings, type SettingsInput, type Theme } from "./api.js";
import { session } from "./session.js";

export const settings = writable<Settings | null>(null);

// Set once the user picks a theme from the nav toggle. mode-watcher keeps that
// choice in localStorage, but a later load of the stored settings would
// otherwise undo it on the way back to the settings page.
let themePicked = false;

let inflight: Promise<Settings | null> | null = null;

// loadSettings resolves to the caller's saved preferences, fetching them at
// most once per page load no matter how many components ask. It resolves to
// null when the fetch fails rather than rejecting, so callers that only want
// the value do not each need their own error path.
export function loadSettings(): Promise<Settings | null> {
  // A value already in the store is newer than the initial fetch: saveSettings
  // replaces it with whatever the server has just confirmed. Handing back the
  // stale original would let a caller overwrite the newer one.
  const cached = get(settings);
  if (cached) return Promise.resolve(cached);
  if (!inflight) {
    inflight = api.getSettings().then(
      (data) => {
        settings.set(data);
        if (!themePicked) setMode(data.theme);
        return data;
      },
      () => {
        // Guests get a 401. Do not cache that, so signing in later retries.
        inflight = null;
        return null;
      },
    );
  }
  return inflight;
}

// The signed-in user's preferences are fetched as soon as there is a session,
// rather than by whichever page happens to need them. That is what makes the
// saved theme apply on every page and lets the nav toggle persist from
// anywhere, instead of only after visiting the settings page first.
session.subscribe((s) => {
  if (s.status === "authed") {
    loadSettings();
  } else if (s.status === "guest") {
    // Forget the previous account's preferences so the next login re-reads
    // its own, and lets the new account's saved theme take effect.
    settings.set(null);
    inflight = null;
    themePicked = false;
  }
});

// The server rejects unknown fields, so the writable subset is spelled out
// rather than spread through from a full Settings object.
function toInput(next: SettingsInput): SettingsInput {
  return {
    display_name: next.display_name,
    default_passes: next.default_passes,
    theme: next.theme,
  };
}

// saveSettings replaces every preference at once and mirrors the server's
// response back into the store, so a rejected change never leaves the UI
// claiming it was saved.
export async function saveSettings(next: SettingsInput): Promise<Settings> {
  const saved = await api.updateSettings(toInput(next));
  settings.set(saved);
  setMode(saved.theme);
  return saved;
}

// nextTheme is the nav toggle's step through the three modes. The two-state
// cycle that mode-watcher offers on its own would leave "system" unreachable.
export function nextTheme(current: Theme): Theme {
  if (current === "dark") return "light";
  if (current === "light") return "system";
  return "dark";
}

// stageTheme applies a theme choice locally without persisting it. Callers that
// batch their writes (the settings page autosave) use this so a pending edit to
// another field is never sent with the value it had before the theme changed.
export function stageTheme(theme: Theme): void {
  themePicked = true;
  setMode(theme);
}

// applyTheme switches appearance immediately and, for a signed-in user,
// persists the choice so it follows them to their other devices. A guest gets
// the local mode-watcher behaviour only.
export async function applyTheme(theme: Theme): Promise<void> {
  stageTheme(theme);
  const current = get(settings);
  if (!current || current.theme === theme) return;
  try {
    await saveSettings({ ...current, theme });
  } catch {
    // The mode is already applied locally; a failed write only costs the
    // choice on the next visit.
  }
}