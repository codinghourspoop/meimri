<script>
  import User from "@lucide/svelte/icons/user";
  import Lock from "@lucide/svelte/icons/lock";
  import { api } from "../lib/api.js";
  import { loadMe } from "../lib/session.js";
  import { goto } from "../lib/router.js";

  let username = $state("");
  let password = $state("");
  let err = $state("");
  let busy = $state(false);

  async function submit(e) {
    e.preventDefault();
    err = "";
    busy = true;
    try {
      await api.login(username.trim(), password);
      await loadMe();
      goto("/");
    } catch (e2) {
      err = e2?.payload?.error || "Invalid username or password.";
    } finally {
      busy = false;
    }
  }
</script>

<div class="mx-auto max-w-sm">
  <h1 class="text-2xl font-bold">Login</h1>
  {#if err}
    <div role="alert" class="alert alert-error mt-3"><span>{err}</span></div>
  {/if}
  <form onsubmit={submit} class="mt-6 space-y-4">
    <label class="input w-full">
      <User class="size-5" />
      <input name="username" required placeholder="Username" class="w-full" bind:value={username} />
    </label>
    <label class="input w-full">
      <Lock class="size-5" />
      <input name="password" placeholder="Password" type="password" required class="w-full" bind:value={password} />
    </label>
    <div class="flex w-full justify-center">
      <button type="submit" class="btn btn-primary btn-wide" disabled={busy}>Login</button>
    </div>
  </form>
  <p class="mt-4 text-center text-sm">
    <a href="/recover" class="link">Recover with code</a> · <a href="/register" class="link">Register</a>
  </p>
</div>
