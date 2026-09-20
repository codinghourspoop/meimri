<script>
  import { api } from "../lib/api.js";
  import { loadMe, pendingRecoveryCodes } from "../lib/session.js";
  import { goto } from "../lib/router.js";

  let username = $state("");
  let displayName = $state("");
  let password = $state("");
  let err = $state("");
  let busy = $state(false);

  async function submit(e) {
    e.preventDefault();
    err = "";
    busy = true;
    try {
      const data = await api.register(username.trim(), displayName.trim(), password);
      pendingRecoveryCodes.set(data.recovery_codes ?? []);
      await loadMe();
      goto("/recovery-codes");
    } catch (e2) {
      err = e2?.payload?.error || "Registration failed.";
    } finally {
      busy = false;
    }
  }
</script>

<div class="mx-auto max-w-sm">
  <h1 class="text-2xl font-bold">Create account</h1>
  {#if err}
    <div role="alert" class="alert alert-error mt-3"><span>{err}</span></div>
  {/if}
  <form onsubmit={submit} class="mt-6 space-y-4">
    <fieldset class="fieldset">
      <legend class="fieldset-legend">Username (3-32 chars, letters/numbers/_/-)</legend>
      <input
        name="username"
        required
        pattern="[a-zA-Z0-9_-]{3,32}"
        class="input w-full"
        placeholder="Username"
        bind:value={username}
      />
    </fieldset>
    <fieldset class="fieldset">
      <legend class="fieldset-legend">Display name</legend>
      <input name="display_name" required class="input w-full" placeholder="Display name" bind:value={displayName} />
    </fieldset>
    <fieldset class="fieldset">
      <legend class="fieldset-legend">Password (min 8 chars)</legend>
      <input
        name="password"
        type="password"
        required
        minlength="8"
        class="input w-full"
        placeholder="Password"
        bind:value={password}
      />
    </fieldset>
    <button type="submit" class="btn btn-primary w-full" disabled={busy}>Create account</button>
  </form>
</div>
