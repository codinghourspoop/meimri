<script>
  import Plus from "@lucide/svelte/icons/plus";
  import X from "@lucide/svelte/icons/x";
  import ArrowLeftRight from "@lucide/svelte/icons/arrow-left-right";
  import Bold from "@lucide/svelte/icons/bold";
  import Italic from "@lucide/svelte/icons/italic";
  import Underline from "@lucide/svelte/icons/underline";
  import Strikethrough from "@lucide/svelte/icons/strikethrough";
  import List from "@lucide/svelte/icons/list";
  import ListOrdered from "@lucide/svelte/icons/list-ordered";
  import RemoveFormatting from "@lucide/svelte/icons/remove-formatting";
  import Image from "@lucide/svelte/icons/image";
  import Mic from "@lucide/svelte/icons/mic";
  import { api } from "../lib/api.js";
  import { goto } from "../lib/router.js";
  import { session } from "../lib/session.js";
  import { onMount } from "svelte";
  import { Button } from "../lib/components/ui/button/index.js";
  import { Input } from "../lib/components/ui/input/index.js";
  import { Label } from "../lib/components/ui/label/index.js";
  import { Alert, AlertDescription } from "../lib/components/ui/alert/index.js";
  import RichField from "../components/RichField.svelte";

  let uid = 1;
  const newCard = () => ({ id: uid++, front: "", back: "" });

  let title = $state("");
  let description = $state("");
  let cards = $state([newCard()]);
  let err = $state("");
  let busy = $state(false);

  // Shared-toolbar state: which field editor is focused.
  let editors = {};
  let focused = $state(null);
  let activeKey = $state("");
  let uiTick = $state(0);

  onMount(() => {
    if ($session.status === "guest") goto("/login");
  });

  function stripHtml(html) {
    const el = document.createElement("div");
    el.innerHTML = html ?? "";
    return (el.textContent || "").trim();
  }

  function register(quill, key) {
    editors[key] = quill;
  }

  function setActive(quill, key) {
    focused = quill;
    activeKey = key;
    cachedFormat = quill.getFormat() || {};
    uiTick++;
  }

  // Last-known format of the focused editor. Updated only from event
  // handlers — never read Quill during render, it can steal focus back.
  let cachedFormat = $state({});

  function activeField(cardId) {
    const [id, field] = (activeKey || "").split(":");
    return id === String(cardId) ? field : "front";
  }

  function activeLabel() {
    if (!activeKey) return "";
    const [id, field] = activeKey.split(":");
    const idx = cards.findIndex((c) => String(c.id) === id);
    if (idx < 0) return "";
    return `${field === "front" ? "Term" : "Definition"} ${idx + 1}`;
  }

  function toggleInline(name) {
    if (!focused) return;
    focused.format(name, !focused.getFormat()[name]);
    cachedFormat = focused.getFormat() || {};
    uiTick++;
  }

  function toggleList(kind) {
    if (!focused) return;
    const cur = focused.getFormat().list;
    focused.format("list", cur === kind ? false : kind);
    cachedFormat = focused.getFormat() || {};
    uiTick++;
  }

  function clearFormat() {
    if (!focused) return;
    const r = focused.getSelection();
    if (r && r.length > 0) focused.removeFormat(r.index, r.length);
    focused.format("list", false);
    cachedFormat = focused.getFormat() || {};
    uiTick++;
  }

  // Prevent toolbar clicks from stealing the editor's selection.
  // Attached to the buttons themselves (interactive elements), so no
  // container-level mousedown handler is needed.
  function keepSelection(e) {
    e.preventDefault();
  }

  // Image-URL bar state (per card, no backend needed).
  let imgCardId = $state(null);
  let imgUrl = $state("");

  function toggleImage(cardId) {
    imgCardId = imgCardId === cardId ? null : cardId;
    imgUrl = "";
  }

  function normalizeUrl(raw) {
    const url = (raw || "").trim();
    if (!url) return "";
    if (/^[a-z][a-z0-9+.-]*:/i.test(url)) return url;
    return `https://${url}`;
  }

  function insertImageUrl(cardId) {
    const url = normalizeUrl(imgUrl);
    if (!url) return;
    const field = activeField(cardId);
    const ed = editors[`${cardId}:${field}`];
    if (!ed) return;
    const range = ed.getSelection(true);
    const index = range ? range.index : ed.getLength();
    ed.insertEmbed(index, "image", url, "user");
    ed.setSelection(index + 1, 0);
    imgCardId = null;
    imgUrl = "";
  }

  function swapCard(card) {
    const f = card.front;
    card.front = card.back;
    card.back = f;
  }

  function addCard() {
    cards = [...cards, newCard()];
  }
  function removeCard(id) {
    cards = cards.filter((c) => c.id !== id);
    if (activeKey.startsWith(`${id}:`)) {
      focused = null;
      activeKey = "";
    }
    delete editors[`${id}:front`];
    delete editors[`${id}:back`];
    if (cards.length === 0) cards = [newCard()];
  }

  async function submit(e) {
    e.preventDefault();
    err = "";
    busy = true;
    try {
      const payload = cards
        .map((c) => ({ front: c.front || "", back: c.back || "" }))
        .filter((c) => stripHtml(c.front) || stripHtml(c.back));
      const data = await api.createDeck(title.trim(), description.trim(), payload);
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
    <Alert variant="destructive" class="mt-3"><AlertDescription>{err}</AlertDescription></Alert>
  {/if}
  <form onsubmit={submit} class="mt-6 space-y-4">
    <div class="space-y-2">
      <Label for="deck-title">Title</Label>
      <Input id="deck-title" name="title" required placeholder="Deck title" bind:value={title} />
    </div>
    <div class="space-y-2">
      <Label for="deck-description">Description</Label>
      <Input
        id="deck-description"
        name="description"
        placeholder="What is this deck about?"
        bind:value={description}
      />
    </div>
    <div class="space-y-2">
      <Label>Cards</Label>
      <!-- Prevent toolbar clicks from stealing the editor's selection via keepSelection on each button -->
      <div
        class="bg-background/95 sticky top-0 z-10 flex items-center gap-0.5 rounded-xl border px-2 py-1.5 backdrop-blur"
      >
        <Button
          variant="ghost"
          size="icon-sm"
          aria-label="Bold"
          disabled={!focused}
          class={cachedFormat.bold ? "bg-muted text-primary" : ""}
          onmousedown={keepSelection}
          onclick={() => toggleInline("bold")}
        >
          <Bold />
        </Button>
        <Button
          variant="ghost"
          size="icon-sm"
          aria-label="Italic"
          disabled={!focused}
          class={cachedFormat.italic ? "bg-muted text-primary" : ""}
          onmousedown={keepSelection}
          onclick={() => toggleInline("italic")}
        >
          <Italic />
        </Button>
        <Button
          variant="ghost"
          size="icon-sm"
          aria-label="Underline"
          disabled={!focused}
          class={cachedFormat.underline ? "bg-muted text-primary" : ""}
          onmousedown={keepSelection}
          onclick={() => toggleInline("underline")}
        >
          <Underline />
        </Button>
        <Button
          variant="ghost"
          size="icon-sm"
          aria-label="Strikethrough"
          disabled={!focused}
          class={cachedFormat.strike ? "bg-muted text-primary" : ""}
          onmousedown={keepSelection}
          onclick={() => toggleInline("strike")}
        >
          <Strikethrough />
        </Button>
        <Button
          variant="ghost"
          size="icon-sm"
          aria-label="Bullet list"
          disabled={!focused}
          class={cachedFormat.list === "bullet" ? "bg-muted text-primary" : ""}
          onmousedown={keepSelection}
          onclick={() => toggleList("bullet")}
        >
          <List />
        </Button>
        <Button
          variant="ghost"
          size="icon-sm"
          aria-label="Numbered list"
          disabled={!focused}
          class={cachedFormat.list === "ordered" ? "bg-muted text-primary" : ""}
          onmousedown={keepSelection}
          onclick={() => toggleList("ordered")}
        >
          <ListOrdered />
        </Button>
        <Button
          variant="ghost"
          size="icon-sm"
          aria-label="Clear formatting"
          disabled={!focused}
          onmousedown={keepSelection}
          onclick={clearFormat}
        >
          <RemoveFormatting />
        </Button>
      </div>
      {#each cards as card (card.id)}
        {@const idx = cards.findIndex((c) => c.id === card.id)}
        <div class="rounded-xl border">
          <div class="flex items-center justify-end gap-1 px-2 pt-1">
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              aria-label="Swap term and definition"
              title="Swap term and definition"
              onclick={() => swapCard(card)}
            >
              <ArrowLeftRight />
            </Button>
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              aria-label="Remove card"
              onclick={() => removeCard(card.id)}
            >
              <X />
            </Button>
          </div>
          <div class="flex gap-2 px-3 pt-1 pb-3">
            <div class="flex flex-col gap-1 pt-6">
              <Button
                type="button"
                variant="ghost"
                size="icon-sm"
                aria-label="Add image"
                title="Add image from URL"
                class={imgCardId === card.id ? "bg-muted text-primary" : ""}
                onclick={() => toggleImage(card.id)}
              >
                <Image />
              </Button>
              <Button
                type="button"
                variant="ghost"
                size="icon-sm"
                aria-label="Record audio (coming soon)"
                title="Voice recording coming soon"
                disabled
              >
                <Mic />
              </Button>
            </div>
            <div class="min-w-0 flex-1 space-y-3">
              <RichField
                label="Term"
                placeholder="Add a term…"
                fieldKey="{card.id}:front"
                bind:value={card.front}
                onReady={register}
                onActive={setActive}
                onSel={() => uiTick++}
              />
              <RichField
                label="Definition"
                placeholder="Add a definition…"
                fieldKey="{card.id}:back"
                bind:value={card.back}
                onReady={register}
                onActive={setActive}
                onSel={() => uiTick++}
              />
            </div>
          </div>
          <div class="text-muted-foreground px-3 pb-2 font-mono text-xs">Card {idx + 1}</div>
          {#if imgCardId === card.id}
            <div class="flex gap-2 px-3 pb-3">
              <Input
                placeholder="Paste an image URL…"
                bind:value={imgUrl}
                onkeydown={(e) => {
                  if (e.key === "Enter") {
                    e.preventDefault();
                    insertImageUrl(card.id);
                  }
                }}
              />
              <Button type="button" size="sm" class="shrink-0" onclick={() => insertImageUrl(card.id)}>
                Insert
              </Button>
            </div>
          {/if}
        </div>
      {/each}
      <Button type="button" variant="outline" size="sm" onclick={addCard}>
        <Plus class="size-4" />
        Add card
      </Button>
    </div>
    <Button type="submit" disabled={busy}>Create deck</Button>
  </form>
</div>
