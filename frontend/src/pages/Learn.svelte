<script>
  import { onMount, onDestroy } from "svelte";
  import X from "@lucide/svelte/icons/x";
  import RotateCcw from "@lucide/svelte/icons/rotate-ccw";
  import Trophy from "@lucide/svelte/icons/trophy";
  import { api } from "../lib/api.js";
  import { loadSettings } from "../lib/settings.js";
  import { Button } from "../lib/components/ui/button/index.js";
  import { Input } from "../lib/components/ui/input/index.js";
  import { Alert, AlertDescription } from "../lib/components/ui/alert/index.js";
  import { Progress } from "../lib/components/ui/progress/index.js";
  import { Spinner } from "../lib/components/ui/spinner/index.js";
  import { Separator } from "../lib/components/ui/separator/index.js";
  import CardHtml from "../components/CardHtml.svelte";

  let { id } = $props();

  const params = new URLSearchParams(window.location.search);
  const passesParam = parseInt(params.get("passes") ?? "", 10);
  // Resolved in onMount: an explicit ?passes= wins, otherwise the account's
  // saved default applies, which may still be in flight on the first render.
  let passesRequired = 3;

  let title = $state("");
  let cards = $state([]);
  let loading = $state(true);
  let loadError = $state("");

  // Session state (ported from learn.js).
  let queue = $state([]);
  let current = $state(null);
  let feedback = $state(null);
  let answer = $state("");
  let stages = $state({});
  let totalUnits = $state(0);
  let mastered = $state(0);
  let total = $state(0);
  let attempts = $state(0);
  let correctFirst = $state(0);
  let seen = $state({});
  let missed = $state({});
  let done = $state(false);
  let empty = $state(false);
  let advanceTimer = null;

  let doneUnits = $derived(Object.values(stages).reduce((a, b) => a + b, 0));
  let progress = $derived(totalUnits ? Math.min(100, Math.round((doneUnits / totalUnits) * 100)) : 0);
  let progressLabel = $derived(`${doneUnits} of ${totalUnits}`);

  function norm(s) {
    return stripHtml(s).replace(/\s+/g, " ").toLowerCase();
  }
  function stripHtml(html) {
    const el = document.createElement("div");
    el.innerHTML = html ?? "";
    return (el.textContent || "").trim();
  }
  function shuffle(a) {
    for (let i = a.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1));
      [a[i], a[j]] = [a[j], a[i]];
    }
    return a;
  }
  function requeueSoon(item) {
    const pos = 1 + Math.floor(Math.random() * (Math.min(4, queue.length) + 1));
    queue.splice(Math.min(pos, queue.length), 0, item);
  }
  function requeueLater(item) {
    const pos = Math.floor(Math.random() * (queue.length + 1));
    queue.splice(pos, 0, item);
  }
  function advance(idx, stage) {
    const ns = stage + 1;
    stages[idx] = Math.max(stages[idx] || 0, ns);
    if (ns >= passesRequired) mastered++;
    else requeueLater({ idx, stage: ns });
  }

  function next() {
    if (advanceTimer) {
      clearTimeout(advanceTimer);
      advanceTimer = null;
    }
    feedback = null;
    answer = "";
    if (queue.length === 0) {
      current = null;
      done = true;
      return;
    }
    const item = queue.shift();
    const card = cards[item.idx];
    if (item.stage === 0 && total > 1 && Math.random() >= 0.5) {
      current = {
        idx: item.idx,
        stage: item.stage,
        front: card.front,
        back: card.back,
        type: "mc",
        options: buildOptions(item.idx),
        picked: -1,
      };
    } else if (item.stage === 0 && total > 1) {
      const tf = buildTF(item.idx);
      current = {
        idx: item.idx,
        stage: item.stage,
        front: card.front,
        back: card.back,
        type: "tf",
        tfStatement: tf.statement,
        tfAnswer: tf.answer,
        picked: null,
      };
    } else {
      current = { idx: item.idx, stage: item.stage, front: card.front, back: card.back, type: "written", picked: -1 };
    }
  }

  function buildOptions(idx) {
    const correct = cards[idx].back;
    const pool = [];
    const seenSet = new Set([norm(correct)]);
    for (let i = 0; i < cards.length && pool.length < 20; i++) {
      if (i === idx) continue;
      const b = cards[i].back;
      const k = norm(b);
      if (!k || seenSet.has(k)) continue;
      seenSet.add(k);
      pool.push(b);
    }
    shuffle(pool);
    const opts = pool.slice(0, 3).map((text) => ({ text, correct: false }));
    opts.push({ text: correct, correct: true });
    return shuffle(opts);
  }

  function buildTF(idx) {
    const card = cards[idx];
    const n = norm(card.back);
    const pool = [];
    for (let i = 0; i < cards.length; i++) {
      if (i === idx) continue;
      const b = cards[i].back;
      if (!b || norm(b) === n || pool.includes(b)) continue;
      pool.push(b);
    }
    const showCorrect = pool.length === 0 || Math.random() < 0.5;
    return {
      statement: showCorrect ? card.back : pool[Math.floor(Math.random() * pool.length)],
      answer: showCorrect,
    };
  }

  function mcButtonClass(opt) {
    if (feedback && opt.correct) return "border-green-600 bg-green-600/10";
    if (feedback && current.picked === current.options.indexOf(opt))
      return "border-red-600 bg-red-600/10 learn-shake";
    if (feedback) return "opacity-60";
    return "hover:border-primary/60";
  }

  function mcBadgeClass(opt) {
    if (feedback && opt.correct) return "bg-green-600 text-white";
    if (feedback && current.picked === current.options.indexOf(opt)) return "bg-red-600 text-white";
    return "bg-muted text-muted-foreground";
  }

  function restart() {
    if (advanceTimer) clearTimeout(advanceTimer);
    totalUnits = passesRequired * total;
    stages = {};
    const idxs = cards.map((_, i) => i);
    shuffle(idxs);
    queue = idxs.map((idx) => ({ idx, stage: 0 }));
    mastered = 0;
    attempts = 0;
    correctFirst = 0;
    seen = {};
    missed = {};
    done = false;
    feedback = null;
    answer = "";
    next();
  }

  function firstTry(idx) {
    return !seen[idx];
  }

  function answerMC(i) {
    if (!current || current.type !== "mc" || feedback) return;
    current.picked = i;
    const opt = current.options[i];
    attempts++;
    const first = firstTry(current.idx);
    seen[current.idx] = true;
    if (opt && opt.correct) {
      if (first) correctFirst++;
      feedback = { correct: true, message: "Nice!" };
      advance(current.idx, current.stage);
      advanceTimer = setTimeout(next, 650);
    } else {
      missed[current.idx] = (missed[current.idx] || 0) + 1;
      feedback = { correct: false, message: "Not quite." };
      requeueSoon({ idx: current.idx, stage: 0 });
      advanceTimer = setTimeout(next, 2200);
    }
  }

  function answerTF(value) {
    if (!current || current.type !== "tf" || feedback) return;
    current.picked = value;
    attempts++;
    const first = firstTry(current.idx);
    seen[current.idx] = true;
    if (value === current.tfAnswer) {
      if (first) correctFirst++;
      feedback = { correct: true, message: "Nice!" };
      advance(current.idx, current.stage);
      advanceTimer = setTimeout(next, 650);
    } else {
      missed[current.idx] = (missed[current.idx] || 0) + 1;
      feedback = { correct: false, message: "Not quite." };
      requeueSoon({ idx: current.idx, stage: 0 });
      advanceTimer = setTimeout(next, 2200);
    }
  }

  function submitWritten() {
    if (!current || current.type !== "written" || (feedback && !feedback.correct)) return;
    if (feedback && feedback.correct) {
      next();
      return;
    }
    attempts++;
    const first = firstTry(current.idx);
    seen[current.idx] = true;
    if (norm(answer) === norm(current.back)) {
      if (first) correctFirst++;
      feedback = { correct: true, message: "Correct!" };
      advance(current.idx, current.stage);
      advanceTimer = setTimeout(next, 650);
    } else {
      missed[current.idx] = (missed[current.idx] || 0) + 1;
      feedback = { correct: false, message: "The correct answer is:" };
      requeueSoon({ idx: current.idx, stage: 0 });
    }
  }

  function dontKnow() {
    if (!current || feedback) return;
    attempts++;
    seen[current.idx] = true;
    missed[current.idx] = (missed[current.idx] || 0) + 1;
    feedback = { correct: false, message: "The correct answer is:" };
    if (current.type === "written") answer = "";
    else advanceTimer = setTimeout(next, 2200);
    requeueSoon({ idx: current.idx, stage: 0 });
  }

  function cont() {
    next();
  }

  function missedList() {
    return Object.keys(missed).map((k) => ({ ...cards[Number(k)], count: missed[k] }));
  }

  function onKey(e) {
    if (done || empty || !current) return;
    if (current.type === "mc") {
      if (!feedback && ["1", "2", "3", "4"].includes(e.key)) {
        const i = Number(e.key) - 1;
        if (i < current.options.length) answerMC(i);
        return;
      }
      if (e.key === "Enter" && feedback && !feedback.correct) {
        e.preventDefault();
        cont();
      }
      return;
    }
    if (current.type === "tf") {
      if (!feedback && (e.key === "t" || e.key === "T")) {
        answerTF(true);
        return;
      }
      if (!feedback && (e.key === "f" || e.key === "F")) {
        answerTF(false);
        return;
      }
      if (e.key === "Enter" && feedback && !feedback.correct) {
        e.preventDefault();
        cont();
      }
      return;
    }
    if (e.key === "Enter" && e.target?.tagName === "INPUT") {
      e.preventDefault();
      if (feedback) cont();
      else submitWritten();
    } else if (e.key === "Enter" && feedback) {
      e.preventDefault();
      cont();
    }
  }

  onMount(async () => {
    window.addEventListener("keydown", onKey);
    try {
      const saved = await loadSettings();
      passesRequired =
        passesParam >= 1 && passesParam <= 10
          ? passesParam
          : Math.min(10, Math.max(1, saved?.default_passes ?? 3));
      const data = await api.getLearn(id, passesRequired);
      title = data.title;
      cards = (data.cards || []).filter((c) => c && (c.front !== "" || c.back !== ""));
      total = cards.length;
      if (total === 0) {
        empty = true;
        return;
      }
      restart();
    } catch (e) {
      loadError = e?.payload?.error || "Failed to load learn session.";
    } finally {
      loading = false;
    }
  });
  onDestroy(() => {
    window.removeEventListener("keydown", onKey);
    if (advanceTimer) clearTimeout(advanceTimer);
  });
</script>

<div class="mx-auto w-full max-w-lg">
  {#if loading}
    <div class="flex h-80 flex-col items-center justify-center gap-4">
      <Spinner class="size-8" />
      <span class="text-muted-foreground text-sm">Loading learn session...</span>
    </div>
  {:else if loadError}
    <Alert class="mt-4"><AlertDescription class="text-sm">{loadError}</AlertDescription></Alert>
  {:else}
    {#if !done}
      <div class="flex items-center gap-3">
        <Button
          href="/decks/{id}"
          variant="ghost"
          size="icon-sm"
          class="shrink-0"
          title="Exit to {title}"
          aria-label="Exit learn mode"
        >
          <X class="size-5" />
        </Button>
        <div class="min-w-0 flex-1">
          <div class="text-muted-foreground flex items-baseline justify-between text-xs">
            <span class="truncate font-medium">{progressLabel}</span>
            <span class="font-mono">{progress}%</span>
          </div>
          <Progress value={progress} max={100} class="mt-1 h-1.5" />
        </div>
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          class="shrink-0"
          title="Restart session"
          aria-label="Restart session"
          onclick={restart}
        >
          <RotateCcw class="size-4" />
        </Button>
      </div>
    {/if}

    {#if empty}
      <Alert class="mt-6"><AlertDescription class="text-sm">This deck has no cards yet.</AlertDescription></Alert>
    {/if}

    {#if !done && !empty && current}
      <div class="mt-6">
        <CardHtml html={current.front} class="min-h-24 text-center text-2xl font-bold text-balance" />

        {#if current.type === "mc"}
          <div class="mt-6 grid gap-2">
            {#each current.options as opt, i (i)}
              <Button
                type="button"
                variant="outline"
                onclick={() => answerMC(i)}
                disabled={!!feedback}
                class="h-auto w-full justify-start gap-3 p-3 text-left {mcButtonClass(opt)}"
              >
                <span
                  class="flex size-7 shrink-0 items-center justify-center rounded-md font-mono text-xs font-bold {mcBadgeClass(
                    opt,
                  )}"
                >
                  {String.fromCharCode(65 + i)}
                </span>
                <CardHtml html={opt.text} tag="span" class="flex-1 text-left" />
              </Button>
            {/each}
            {#if feedback && !feedback.correct}
              <Button type="button" variant="outline" onclick={cont} class="mt-1 w-full p-3 text-sm">
                <span class="text-muted-foreground">Correct answer:&nbsp;</span>
                <CardHtml html={current.back} tag="span" class="font-bold" />
              </Button>
            {/if}
          </div>
        {:else if current.type === "tf"}
          <div class="mt-6">
            <div class="border-border bg-card rounded-xl border p-5 text-center">
              <CardHtml html={current.tfStatement} class="text-xl font-bold" />
            </div>
            <div class="mt-2 grid grid-cols-2 gap-2">
              <Button type="button" variant="outline" onclick={() => answerTF(true)} disabled={!!feedback}>
                <span
                  class="bg-muted text-muted-foreground flex size-7 shrink-0 items-center justify-center rounded-md font-mono text-xs font-bold"
                  >T</span
                >
                <span>True</span>
              </Button>
              <Button type="button" variant="outline" onclick={() => answerTF(false)} disabled={!!feedback}>
                <span
                  class="bg-muted text-muted-foreground flex size-7 shrink-0 items-center justify-center rounded-md font-mono text-xs font-bold"
                  >F</span
                >
                <span>False</span>
              </Button>
            </div>
            {#if feedback && !feedback.correct}
              <Button type="button" variant="outline" onclick={cont} class="mt-2 w-full p-3 text-sm">
                <span class="text-muted-foreground">Correct answer:&nbsp;</span>
                <CardHtml html={current.back} tag="span" class="font-bold" />
              </Button>
            {/if}
          </div>
        {:else if current.type === "written"}
          <div class="mt-6">
            <Input
              type="text"
              placeholder="Type the answer"
              autocomplete="off"
              class="h-12 text-center text-lg"
              bind:value={answer}
              disabled={!!feedback}
            />
            {#if feedback && !feedback.correct}
              <div class="mt-3 text-center text-sm">
                <p>
                  <span class="text-muted-foreground">Correct answer: </span><CardHtml
                    html={current.back}
                    tag="span"
                    class="inline font-bold"
                  />
                </p>
                {#if answer}
                  <p class="text-muted-foreground mt-0.5">You wrote: <span>{answer}</span></p>
                {/if}
              </div>
            {/if}
            {#if !feedback}
              <div class="mt-3">
                <Button type="button" class="w-full" onclick={submitWritten}>Check</Button>
                <Button type="button" variant="ghost" size="sm" class="mt-1 w-full" onclick={dontKnow}
                  >I don't know</Button
                >
              </div>
            {/if}
            {#if feedback && !feedback.correct}
              <Button type="button" class="mt-3 w-full" onclick={cont}>Continue</Button>
            {/if}
          </div>
        {/if}
      </div>
    {/if}

    {#if done}
      {@const missedItems = missedList()}
      <div class="mt-8 text-center">
        <div class="bg-primary/15 text-primary mx-auto flex size-16 items-center justify-center rounded-full">
          <Trophy class="size-7" />
        </div>
        <h2 class="mt-3 text-xl font-bold">Session complete</h2>
        <p class="text-muted-foreground mt-4 text-sm">
          <span class="text-foreground font-mono font-bold"
            >{total ? Math.round((correctFirst / total) * 100) : 0}</span
          ><span class="font-mono">% correct first try · </span><span class="font-mono">{attempts}</span><span
            class="font-mono"
          >
            attempts</span
          >
        </p>
        {#if missedItems.length > 0}
          <div class="mt-5 text-left">
            <ul class="divide-border divide-y overflow-hidden rounded-xl border">
              {#each missedItems as m, i (i)}
                <li class="flex items-baseline justify-between gap-4 px-4 py-2.5 text-sm">
                  <CardHtml html={m.front} tag="span" class="font-semibold" />
                  <CardHtml html={m.back} tag="span" class="text-muted-foreground text-right" />
                </li>
              {/each}
            </ul>
          </div>
        {/if}
        <div class="mt-5 flex justify-center gap-2">
          <Button type="button" onclick={restart}>Study again</Button>
          <Button href="/decks/{id}" variant="ghost">Flashcards</Button>
        </div>
      </div>
    {/if}
  {/if}
</div>
