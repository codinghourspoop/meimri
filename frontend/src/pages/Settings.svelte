<script>
  import { onMount, onDestroy } from "svelte";
  import Check from "@lucide/svelte/icons/check";
  import Monitor from "@lucide/svelte/icons/monitor";
  import Moon from "@lucide/svelte/icons/moon";
  import Sun from "@lucide/svelte/icons/sun";
  import { api } from "../lib/api.js";
  import { goto } from "../lib/router.js";
  import { loadMe, pendingRecoveryCodes, session } from "../lib/session.js";
  import { loadSettings, saveSettings, stageTheme } from "../lib/settings.js";
  import { Alert, AlertDescription } from "../lib/components/ui/alert/index.js";
  import { Button } from "../lib/components/ui/button/index.js";
  import { Input } from "../lib/components/ui/input/index.js";
  import { Label } from "../lib/components/ui/label/index.js";
  import { Separator } from "../lib/components/ui/separator/index.js";
  import { Spinner } from "../lib/components/ui/spinner/index.js";

  const MIN_PASSES = 1;
  const MAX_PASSES = 10;
  const MAX_NAME = 64;

  // Idle time before an edit is written. Long enough that typing a display name
  // does not fire one request per keystroke, short enough that it reads as
  // automatic rather than as a hidden Save button.
  const DEBOUNCE_MS = 600;

  const themes = [
    { value: "dark", label: "Dark", icon: Moon },
    { value: "light", label: "Light", icon: Sun },
    { value: "system", label: "System", icon: Monitor },
  ];

  let loading = $state(true);
  let status = $state("idle");
  let statusErr = $state("");

  let username = $state("");
  let displayName = $state("");
  let passes = $state(3);
  let theme = $state("system");

  let regenBusy = $state(false);
  let regenErr = $state("");

  // Autosave bookkeeping. These are deliberately not $state: the effect below
  // reads them to decide whether to write, and making them reactive would let
  // flipping one schedule another save.
  let timer = null;
  // Identifies the most recent write so a slow earlier response cannot mark a
  // newer edit as saved, or overwrite the store with stale values.
  let latest = 0;
  // What the server is known to hold. Comparing against it is what stops the
  // initial load, and the server's own trimming of a field, from being
  // mistaken for an edit worth writing back.
  let saved = { display_name: "", default_passes: 3, theme: "system" };
  let sessionName = "";

  // The number input reports null while it is empty, so its value is only
  // trustworthy once it parses. Anything unparseable falls back to the server
  // default rather than sending NaN.
  function clampPasses(n) {
    const v = Math.trunc(Number(n));
    if (!Number.isFinite(v)) return 3;
    return Math.min(MAX_PASSES, Math.max(MIN_PASSES, v));
  }

  // Reads every field the form can write.
  function payload() {
    return {
      display_name: displayName.trim(),
      default_passes: clampPasses(passes),
      theme,
    };
  }

  function storedOf(data) {
    return {
      display_name: data.display_name,
      default_passes: data.default_passes,
      theme: data.theme,
    };
  }

  // A half-typed display name is not a setting worth sending. The field is
  // already marked required, so the browser blocks the case this misses.
  function valid() {
    const name = displayName.trim();
    return name !== "" && name.length <= MAX_NAME;
  }

  function schedule() {
    if (timer) clearTimeout(timer);
    if (!valid()) {
      status = "idle";
      return;
    }
    timer = setTimeout(() => {
      timer = null;
      save();
    }, DEBOUNCE_MS);
  }

  async function save({ quiet = false } = {}) {
    const id = ++latest;
    if (!quiet) {
      status = "saving";
      statusErr = "";
    }
    try {
      const data = await saveSettings(payload());
      saved = storedOf(data);
      if (quiet || id !== latest) return;
      // Mirror back what the server normalised, so a trimmed or clamped field
      // agrees with what was stored. `saved` already matches, so the effect
      // does not read these as fresh edits.
      displayName = data.display_name;
      passes = data.default_passes;
      if (data.display_name !== sessionName) {
        sessionName = data.display_name;
        // The home greeting reads the display name off the session.
        loadMe();
      }
      status = "saved";
    } catch (e) {
      if (quiet || id !== latest) return;
      status = "error";
      statusErr = e?.payload?.error || "Could not save settings.";
    }
  }

  $effect(() => {
    const next = payload();
    // Nothing to write until the form has loaded, and nothing to write when the
    // fields already agree with what the server holds.
    const unchanged =
      next.display_name === saved.display_name &&
      next.default_passes === saved.default_passes &&
      next.theme === saved.theme;
    if (loading || unchanged) return;
    schedule();
  });

  function pickTheme(next) {
    theme = next;
    // Paint it now; the write rides along with the next autosave so it is not
    // sent with a stale display name.
    stageTheme(next);
  }

  function preventSubmit(e) {
    e.preventDefault();
  }

  onMount(async () => {
    if ($session.status === "guest") {
      goto("/login");
      return;
    }
    const data = await loadSettings();
    if (data) {
      username = data.username;
      displayName = data.display_name;
      passes = data.default_passes;
      theme = data.theme;
      saved = storedOf(data);
      sessionName = data.display_name;
    }
    loading = false;
  });

  onDestroy(() => {
    // Navigating away mid-debounce would otherwise drop the last edit. The
    // component is gone by the time this resolves, so nothing is rendered.
    if (timer) {
      clearTimeout(timer);
      timer = null;
      if (valid()) save({ quiet: true });
    }
  });

  async function regenerate() {
    regenErr = "";
    regenBusy = true;
    try {
      const data = await api.regenerateRecoveryCodes();
      pendingRecoveryCodes.set(data.recovery_codes);
      goto("/recovery-codes");
    } catch (e2) {
      regenErr = e2?.payload?.error || "Failed to regenerate recovery codes.";
    } finally {
      regenBusy = false;
    }
  }
</script>

<div class="mx-auto max-w-2xl">
  <h1 class="text-2xl font-bold">Settings</h1>
  <p class="text-muted-foreground mt-1 text-sm" aria-live="polite">
    {#if loading}
      Loading settings...
    {:else if status === "saving"}
      <span class="inline-flex items-center gap-1.5"><Spinner class="size-3.5" /> Saving...</span>
    {:else if status === "saved"}
      <span class="inline-flex items-center gap-1.5"><Check class="size-3.5" /> Saved</span>
    {:else if status === "error"}
      <span class="text-destructive">{statusErr}</span>
    {:else}
      Changes save automatically.
    {/if}
  </p>

  {#if !loading}
    <form onsubmit={preventSubmit} class="mt-6 space-y-8">
      <section class="space-y-4">
        <div>
          <h2 class="text-lg font-semibold">Profile</h2>
          <p class="text-muted-foreground text-sm">How you show up across memri.</p>
        </div>
        <div class="space-y-2">
          <Label for="settings-username">Username</Label>
          <Input id="settings-username" name="username" value={username} disabled />
          <p class="text-muted-foreground text-sm">Your username cannot be changed.</p>
        </div>
        <div class="space-y-2">
          <Label for="settings-display-name">Display name</Label>
          <Input
            id="settings-display-name"
            name="display_name"
            required
            maxlength={MAX_NAME}
            autocomplete="nickname"
            placeholder="Display name"
            aria-invalid={displayName.trim() === ""}
            bind:value={displayName}
          />
          {#if displayName.trim() === ""}
            <p class="text-destructive text-sm">A display name is required.</p>
          {/if}
        </div>
      </section>

      <Separator />

      <section class="space-y-4">
        <div>
          <h2 class="text-lg font-semibold">Study</h2>
          <p class="text-muted-foreground text-sm">Defaults for new learning sessions.</p>
        </div>
        <div class="space-y-2">
          <Label for="settings-passes">Passes to master a card</Label>
          <Input
            id="settings-passes"
            name="default_passes"
            type="number"
            min={MIN_PASSES}
            max={MAX_PASSES}
            step="1"
            required
            class="w-24"
            bind:value={passes}
          />
          <p class="text-muted-foreground text-sm">
            How many correct answers a card needs before it counts as mastered. More passes means
            longer sessions, but cards you keep getting wrong stay in rotation until they stick.
          </p>
        </div>
      </section>

      <Separator />

      <section class="space-y-4">
        <div>
          <h2 class="text-lg font-semibold">Appearance</h2>
          <p class="text-muted-foreground text-sm">Saved to your account, not just this browser.</p>
        </div>
        <div class="space-y-2">
          <span class="text-sm leading-none font-medium">Theme</span>
          <div
            class="bg-muted inline-flex w-full items-center rounded-full border p-1 sm:w-auto"
            role="group"
            aria-label="Theme"
          >
            {#each themes as t (t.value)}
              {@const Icon = t.icon}
              <Button
                type="button"
                size="sm"
                aria-pressed={theme === t.value}
                variant={theme === t.value ? "default" : "ghost"}
                class="flex-1 justify-center gap-2 rounded-full px-4 sm:flex-none"
                onclick={() => pickTheme(t.value)}
              >
                <Icon class="size-4" />
                {t.label}
              </Button>
            {/each}
          </div>
        </div>
      </section>

      <Separator />

      <section class="space-y-4">
        <div>
          <h2 class="text-lg font-semibold">Account</h2>
          <p class="text-muted-foreground text-sm">Access and password recovery.</p>
        </div>
        {#if regenErr}
          <Alert variant="destructive"><AlertDescription>{regenErr}</AlertDescription></Alert>
        {/if}
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div class="min-w-0 flex-1">
            <p class="text-sm font-medium">Recovery codes</p>
            <p class="text-muted-foreground text-sm">
              Your only way back in if you forget your password. Issuing a new set immediately
              invalidates the one you have now.
            </p>
          </div>
          <Button type="button" variant="outline" class="shrink-0" onclick={regenerate} disabled={regenBusy}>
            {#if regenBusy}<Spinner />{/if}
            Regenerate
          </Button>
        </div>
      </section>
    </form>
  {/if}
</div>
