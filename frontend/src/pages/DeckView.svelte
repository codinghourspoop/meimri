<script>
  import { onMount, onDestroy } from "svelte";
  import ArrowLeft from "@lucide/svelte/icons/arrow-left";
  import ArrowRight from "@lucide/svelte/icons/arrow-right";
  import { api } from "../lib/api.js";

  let { id } = $props();

  let deck = $state(null);
  let err = $state("");
  let currentCard = $state(0);
  let flipped = $state(false);

  let total = $derived(deck ? deck.cards.length : 0);

  async function load() {
    err = "";
    deck = null;
    try {
      deck = await api.getDeck(id);
    } catch (e) {
      err = e?.payload?.error || "Deck not found.";
    }
  }

  onMount(load);

  function prev() {
    if (currentCard > 0) {
      currentCard--;
      flipped = false;
    }
  }
  function next() {
    if (currentCard < total - 1) {
      currentCard++;
      flipped = false;
    }
  }
  function flip() {
    flipped = !flipped;
  }
  function onKey(e) {
    if (e.key === "ArrowLeft") prev();
    else if (e.key === "ArrowRight") next();
    else if (e.key === "ArrowUp" || e.key === "ArrowDown" || e.key === " " || e.key === "Enter") {
      if (e.target?.tagName !== "INPUT") {
        e.preventDefault();
        flip();
      }
    }
  }

  onMount(() => {
    window.addEventListener("keydown", onKey);
  });
  onDestroy(() => window.removeEventListener("keydown", onKey));
</script>

<div class="mx-auto max-w-2xl">
  {#if err}
    <div role="alert" class="alert mt-4"><span class="text-sm">{err}</span></div>
  {:else if !deck}
    <div class="flex h-80 flex-col items-center justify-center gap-4">
      <span class="loading loading-spinner loading-lg"></span>
      <span class="text-sm text-base-content/75">Loading card...</span>
    </div>
  {:else}
    <div class="mb-8 text-center">
      <h1 class="text-3xl font-bold tracking-tight">{deck.title}</h1>
      <p class="mt-2 text-base-content/80">{deck.description}</p>
      <div class="mt-4 flex items-center justify-center">
        <div class="inline-flex items-center rounded-full border border-base-300 bg-base-200 p-1">
          <span class="btn btn-sm btn-primary rounded-full px-4">Flashcards</span>
          <a href="/decks/{deck.id}/learn" class="btn btn-sm btn-ghost rounded-full px-4">Learn</a>
        </div>
      </div>
    </div>
    <div class="mx-auto w-full max-w-md">
      <div class="[perspective:1000px]">
        {#each deck.cards as item, index (index)}
          {#if currentCard === index}
            <div>
              <button
                type="button"
                onclick={flip}
                aria-label={flipped ? "Show question" : "Show answer"}
                class="relative h-80 w-full text-left transition-transform duration-300 [transform-style:preserve-3d] {flipped
                  ? '[transform:rotateX(180deg)]'
                  : '[transform:rotateX(0deg)]'}"
              >
                <div
                  class="card bg-base-100 border-base-300 absolute inset-0 flex items-center justify-center border p-8 text-center shadow-xl [backface-visibility:hidden]"
                >
                  <p class="text-xl font-medium">{item.front}</p>
                </div>
                <div
                  class="card bg-primary/10 border-primary/20 absolute inset-0 flex items-center justify-center border p-8 text-center shadow-xl [backface-visibility:hidden] [transform:rotateX(180deg)]"
                >
                  <p class="text-lg">{item.back}</p>
                </div>
              </button>
            </div>
          {/if}
        {/each}
      </div>
      <p class="mt-4 text-center text-sm text-base-content/75">
        Click the card or press <span class="kbd kbd-xs">space</span> to flip it ·
        <span class="kbd kbd-xs">←</span> <span class="kbd kbd-xs">→</span> to navigate
      </p>
      <div class="card bg-base-200 mt-2 shadow-sm">
        <div class="card-body flex-row items-center justify-between p-2">
          <button type="button" onclick={prev} disabled={currentCard === 0} class="btn btn-ghost" aria-label="Previous card">
            <ArrowLeft class="size-5" />
          </button>
          <span class="font-mono text-sm">{currentCard + 1}/{total}</span>
          <button
            type="button"
            onclick={next}
            disabled={currentCard === total - 1}
            class="btn btn-ghost"
            aria-label="Next card"
          >
            <ArrowRight class="size-5" />
          </button>
        </div>
      </div>
    </div>
  {/if}
</div>
