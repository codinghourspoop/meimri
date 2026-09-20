<script>
  import { pendingRecoveryCodes } from "../lib/session.js";
  import { goto } from "../lib/router.js";
  import { onMount } from "svelte";

  let codes = $state(null);
  const unsub = pendingRecoveryCodes.subscribe((v) => (codes = v));
  onMount(() => {
    if (!codes) goto("/");
    return unsub;
  });
</script>

<div class="mx-auto max-w-lg">
  <div class="card bg-base-200 shadow-sm">
    <div class="card-body">
      <h1 class="card-title text-2xl">Save your recovery codes</h1>
      <p class="text-sm text-base-content/80">
        Store these somewhere safe. Each code can be used once to reset your password.
        This is the only time they will be shown.
      </p>
      {#if codes}
        <ul class="menu bg-base-100 mt-4 gap-1 rounded-box font-mono text-sm">
          {#each codes as c}
            <li><span class="kbd kbd-sm w-full justify-start">{c}</span></li>
          {/each}
        </ul>
      {/if}
      <div class="card-actions mt-4">
        <a href="/" class="btn btn-primary">Continue to home</a>
      </div>
    </div>
  </div>
</div>
