<script>
  import { session } from "../lib/session.js";
  import DeckSearch from "../components/DeckSearch.svelte";
</script>

{#if $session.status === "authed" && $session.user}
  <div class="hero">
    <div class="hero-content text-center">
      <div>
        <h1 class="text-2xl font-bold">
          Hello, {$session.user.display_name} ({$session.user.username})
        </h1>
        <p class="mt-2 text-base-content/80">Welcome back to memri.</p>
      </div>
    </div>
  </div>
  <DeckSearch />
{:else}
  <div class="hero">
    <div class="hero-content text-center">
      <div class="max-w-2xl">
        <h1 class="text-4xl font-bold tracking-tight">Welcome to memri</h1>
        <p class="mt-4 text-base-content/80">Svelte frontend + Go JSON API. No htmx, no Alpine.</p>
        <div class="mt-8 flex justify-center gap-4">
          <div class="aura">
            <a href="/register" class="btn">Get started</a>
          </div>
          <div class="aura">
            <a href="/login" class="btn">Login</a>
          </div>
        </div>
      </div>
    </div>
  </div>
{/if}
