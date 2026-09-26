<script>
  import { onMount, onDestroy } from "svelte";
  import ArrowLeft from "@lucide/svelte/icons/arrow-left";
  import ArrowRight from "@lucide/svelte/icons/arrow-right";
  import { api } from "../lib/api.js";
  import { Button } from "../lib/components/ui/button/index.js";
  import { Card, CardContent } from "../lib/components/ui/card/index.js";
  import { Alert, AlertDescription } from "../lib/components/ui/alert/index.js";
  import { Spinner } from "../lib/components/ui/spinner/index.js";
  import { Kbd } from "../lib/components/ui/kbd/index.js";

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
    <Alert class="mt-4"><AlertDescription class="text-sm">{err}</AlertDescription></Alert>
  {:else if !deck}
    <div class="flex h-80 flex-col items-center justify-center gap-4">
      <Spinner class="size-8" />
      <span class="text-muted-foreground text-sm">Loading card...</span>
    </div>
  {:else}
    <div class="mb-8 text-center">
      <h1 class="text-3xl font-bold tracking-tight">{deck.title}</h1>
      <p class="text-muted-foreground mt-2">{deck.description}</p>
      <div class="mt-4 flex items-center justify-center">
        <div class="bg-muted inline-flex items-center rounded-full border p-1">
          <Button size="sm" class="rounded-full px-4">Flashcards</Button>
          <Button href="/decks/{deck.id}/learn" variant="ghost" size="sm" class="rounded-full px-4">Learn</Button>
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
                  class="border-border bg-card absolute inset-0 flex items-center justify-center overflow-y-auto rounded-xl border p-8 text-center shadow-xl [backface-visibility:hidden]"
                >
                  <div class="text-xl font-medium [&_ul]:list-disc [&_ul]:pl-5 [&_ol]:list-decimal [&_ol]:pl-5 [&_ul]:text-left [&_ol]:text-left [&_img]:max-w-full [&_img]:rounded-lg">{@html item.front}</div>
                </div>
                <div
                  class="border-primary/20 bg-primary/10 absolute inset-0 flex items-center justify-center overflow-y-auto rounded-xl border p-8 text-center shadow-xl [backface-visibility:hidden] [transform:rotateX(180deg)]"
                >
                  <div class="text-lg [&_ul]:list-disc [&_ul]:pl-5 [&_ol]:list-decimal [&_ol]:pl-5 [&_ul]:text-left [&_ol]:text-left [&_img]:max-w-full [&_img]:rounded-lg">{@html item.back}</div>
                </div>
              </button>
            </div>
          {/if}
        {/each}
      </div>
      <p class="text-muted-foreground mt-4 text-center text-sm">
        Click the card or press <Kbd>space</Kbd> to flip it · <Kbd>←</Kbd> <Kbd>→</Kbd> to navigate
      </p>
      <Card class="bg-muted mt-2 shadow-sm">
        <CardContent class="flex items-center justify-between p-2">
          <Button type="button" variant="ghost" onclick={prev} disabled={currentCard === 0} aria-label="Previous card">
            <ArrowLeft class="size-5" />
          </Button>
          <span class="font-mono text-sm">{currentCard + 1}/{total}</span>
          <Button
            type="button"
            variant="ghost"
            onclick={next}
            disabled={currentCard === total - 1}
            aria-label="Next card"
          >
            <ArrowRight class="size-5" />
          </Button>
        </CardContent>
      </Card>
    </div>
  {/if}
</div>
