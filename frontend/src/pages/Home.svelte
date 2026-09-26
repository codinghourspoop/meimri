<script>
  import { session } from "../lib/session.js";
  import { Button } from "../lib/components/ui/button/index.js";
  import DeckSearch from "../components/DeckSearch.svelte";
</script>

{#if $session.status === "authed" && $session.user}
  <div class="py-6 text-center">
    <h1 class="text-2xl font-bold">
      Hello, {$session.user.display_name} ({$session.user.username})
    </h1>
    <p class="text-muted-foreground mt-2">Welcome back to memri.</p>
  </div>
  <DeckSearch />
{:else}
  <div class="py-16 text-center">
    <div class="mx-auto max-w-2xl">
      <h1 class="text-4xl font-bold tracking-tight">Welcome to memri</h1>
      <p class="text-muted-foreground mt-4">Svelte frontend + Go JSON API. No htmx, no Alpine.</p>
      <div class="mt-8 flex justify-center gap-4">
        <div class="aura">
          <Button href="/register">Get started</Button>
        </div>
        <div class="aura">
          <Button href="/login" variant="outline">Login</Button>
        </div>
      </div>
    </div>
  </div>
{/if}
