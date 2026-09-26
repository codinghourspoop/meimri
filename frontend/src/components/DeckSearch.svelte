<script>
  import Search from "@lucide/svelte/icons/search";
  import { api } from "../lib/api.js";
  import { Input } from "../lib/components/ui/input/index.js";
  import { Alert, AlertDescription } from "../lib/components/ui/alert/index.js";

  let q = $state("");
  let results = $state(null);
  let timer = $state(null);

  async function search(query) {
    if (!query.trim()) {
      results = null;
      return;
    }
    try {
      results = await api.searchDecks(query);
    } catch {
      results = [];
    }
  }

  function onInput() {
    clearTimeout(timer);
    timer = setTimeout(() => search(q), 250);
  }
</script>

<section class="mx-auto mt-6 max-w-2xl">
  <div class="relative">
    <Search class="text-muted-foreground pointer-events-none absolute top-1/2 left-3.5 size-5 -translate-y-1/2" />
    <Input
      placeholder="Search decks..."
      autocomplete="off"
      class="h-12 pl-11 text-base"
      bind:value={q}
      oninput={onInput}
    />
  </div>
  <div class="mt-4 space-y-2" aria-live="polite">
    {#if results === null}
      <p class="text-muted-foreground mt-4 text-center text-sm">Type to search your decks.</p>
    {:else if results.length === 0}
      <Alert class="mt-4">
        <AlertDescription class="text-sm">No decks found.</AlertDescription>
      </Alert>
    {:else}
      {#each results as d (d.id)}
        <a
          href="/decks/{d.id}"
          class="border-border bg-card hover:bg-muted block rounded-xl border p-4 shadow-sm transition"
        >
          <span class="font-semibold">{d.title}</span>
          <span class="text-muted-foreground mt-1 block text-sm">{d.description}</span>
          <span class="text-muted-foreground mt-2 block font-mono text-xs">
            {d.card_count} {d.card_count === 1 ? "card" : "cards"}
          </span>
        </a>
      {/each}
    {/if}
  </div>
</section>
