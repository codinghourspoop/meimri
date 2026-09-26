<script>
  import { pendingRecoveryCodes } from "../lib/session.js";
  import { goto } from "../lib/router.js";
  import { onMount } from "svelte";
  import { Button } from "../lib/components/ui/button/index.js";
  import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../lib/components/ui/card/index.js";

  let codes = $state(null);
  const unsub = pendingRecoveryCodes.subscribe((v) => (codes = v));
  onMount(() => {
    if (!codes) goto("/");
    return unsub;
  });
</script>

<Card class="bg-muted mx-auto max-w-lg">
  <CardHeader>
    <CardTitle class="text-2xl">Save your recovery codes</CardTitle>
    <CardDescription>
      Store these somewhere safe. Each code can be used once to reset your password. This is the only
      time they will be shown.
    </CardDescription>
  </CardHeader>
  <CardContent>
    {#if codes}
      <ul class="bg-card space-y-1 rounded-lg border p-4 font-mono text-sm">
        {#each codes as c}
          <li><code class="bg-muted block rounded px-2 py-1">{c}</code></li>
        {/each}
      </ul>
    {/if}
    <div class="mt-4">
      <Button href="/">Continue to home</Button>
    </div>
  </CardContent>
</Card>
