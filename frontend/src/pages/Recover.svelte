<script>
  import { api } from "../lib/api.js";
  import { loadMe } from "../lib/session.js";
  import { goto } from "../lib/router.js";

  let username = $state("");
  let code = $state("");
  let password = $state("");
  let err = $state("");
  let busy = $state(false);

  async function submit(e) {
    e.preventDefault();
    err = "";
    busy = true;
    try {
      await api.recover(username.trim(), code.trim(), password);
      await loadMe();
      goto("/");
    } catch (e2) {
      err = e2?.payload?.error || "Recovery failed.";
    } finally {
      busy = false;
    }
  }
</script>

<div class="mx-auto max-w-sm">
  <h1 class="text-2xl font-bold">Recover account</h1>
  {#if err}
    <div role="alert" class="alert alert-error mt-3"><span>{err}</span></div>
  {/if}
  <form onsubmit={submit} class="mt-6 space-y-4">
    <fieldset class="fieldset">
      <legend class="fieldset-legend">Username</legend>
      <input name="username" required class="input w-full" placeholder="Username" bind:value={username} />
    </fieldset>
    <fieldset class="fieldset">
      <legend class="fieldset-legend">Recovery code</legend>
      <input name="code" required class="input w-full font-mono" placeholder="Recovery code" bind:value={code} />
    </fieldset>
    <fieldset class="fieldset">
      <legend class="fieldset-legend">New password</legend>
      <input
        name="password"
        type="password"
        required
        minlength="8"
        class="input w-full"
        placeholder="New password"
        bind:value={password}
      />
    </fieldset>
    <button type="submit" class="btn btn-primary w-full" disabled={busy}>Reset password</button>
  </form>
</div>
