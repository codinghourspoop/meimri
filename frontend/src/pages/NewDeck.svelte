<script>
  import Plus from "@lucide/svelte/icons/plus";
  import X from "@lucide/svelte/icons/x";
  import { api } from "../lib/api.js";
  import { goto } from "../lib/router.js";
  import { session } from "../lib/session.js";
  import { onMount } from "svelte";

  let title = $state("");
  let description = $state("");
  let cards = $state([{ front: "", back: "" }]);
  let err = $state("");
  let busy = $state(false);

  onMount(() => {
    if ($session.status === "guest") goto("/login");
  });

  function addCard() {
    cards = [...cards, { front: "", back: "" }];
  }
  function removeCard(i) {
    cards = cards.filter((_, j) => j !== i);
    if (cards.length === 0) cards = [{ front: "", back: "" }];
  }

  async function submit(e) {
    e.preventDefault();
    err = "";
    busy = true;
    try {
      const data = await api.createDeck(
        title.trim(),
        description.trim(),
        cards.map((c) => ({ front: c.front.trim(), back: c.back.trim() })).filter((c) => c.front || c.back),
      );
      goto(`/decks/${data.id}`);
    } catch (e2) {
      err = e2?.payload?.error || "Failed to create deck.";
    } finally {
      busy = false;
    }
  }
</script>

<div class="mx-auto max-w-2xl">
  <h1 class="text-2xl font-bold">New deck</h1>
  {#if err}
    <div role="alert" class="alert alert-error mt-3"><span>{err}</span></div>
  {/if}
  <form onsubmit={submit} class="mt-6 space-y-4">
    <fieldset class="fieldset">
      <legend class="fieldset-legend">Title</legend>
      <input name="title" required class="input w-full" placeholder="Deck title" bind:value={title} />
    </fieldset>
    <fieldset class="fieldset">
      <legend class="fieldset-legend">Description</legend>
      <input name="description" class="input w-full" placeholder="What is this deck about?" bind:value={description} />
    </fieldset>
    <div class="space-y-2">
      <span class="fieldset-legend">Cards</span>
      {#each cards as card, i (i)}
        <div class="flex gap-2">
          <input class="input w-full" placeholder="Front" bind:value={card.front} />
          <input class="input w-full" placeholder="Back" bind:value={card.back} />
          <button
            type="button"
            class="btn btn-ghost btn-square shrink-0"
            aria-label="Remove card"
            onclick={() => removeCard(i)}
          >
            <X class="size-5" />
          </button>
        </div>
      {/each}
      <div class="aura">
        <button type="button" class="btn btn-sm" onclick={addCard}>
          <Plus class="size-4" />
          Add card
        </button>
      </div>
    </div>
    <button type="submit" class="btn btn-primary" disabled={busy}>Create deck</button>
  </form>
</div>
