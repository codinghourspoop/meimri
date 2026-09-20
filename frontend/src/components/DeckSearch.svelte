<script>
  import Search from "@lucide/svelte/icons/search";
  import { api } from "../lib/api.js";

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

<section class="mx-auto mt-8 max-w-2xl">
  <label class="input w-full">
    <Search class="size-5" />
    <input
      placeholder="Search decks..."
      autocomplete="off"
      class="w-full"
      bind:value={q}
      oninput={onInput}
    />
  </label>
  <div class="mt-4 space-y-2" aria-live="polite">
    {#if results === null}
      <p class="mt-4 text-center text-sm text-base-content/75">Type to search your decks.</p>
    {:else if results.length === 0}
      <div role="alert" class="alert mt-4">
        <span class="text-sm">No decks found.</span>
      </div>
    {:else}
      {#each results as d (d.id)}
        <a href="/decks/{d.id}" class="card bg-base-200 block p-4 shadow-sm transition hover:bg-base-300">
          <span class="font-semibold">{d.title}</span>
          <span class="mt-1 block text-sm text-base-content/80">{d.description}</span>
          <span class="mt-2 block font-mono text-xs text-base-content/75">
            {d.card_count} {d.card_count === 1 ? "card" : "cards"}
          </span>
        </a>
      {/each}
    {/if}
  </div>
</section>
