<script>
  import { onMount, onDestroy } from "svelte";
  import X from "@lucide/svelte/icons/x";
  import RotateCcw from "@lucide/svelte/icons/rotate-ccw";
  import Trophy from "@lucide/svelte/icons/trophy";
  import { api } from "../lib/api.js";

  let { id } = $props();

  const params = new URLSearchParams(window.location.search);
  const passesParam = parseInt(params.get("passes") ?? "3", 10);
  const passesRequired = passesParam >= 1 && passesParam <= 10 ? passesParam : 3;

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
    return (s || "").trim().replace(/\s+/g, " ").toLowerCase();
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
    if (feedback && opt.correct) return "border-success bg-success/10";
    if (feedback && current.picked === current.options.indexOf(opt)) return "border-error bg-error/10 learn-shake";
    if (feedback) return "border-base-300 opacity-60";
    return "border-base-300 bg-base-100 hover:border-primary/60";
  }

  function mcBadgeClass(opt) {
    if (feedback && opt.correct) return "bg-success text-success-content";
    if (feedback && current.picked === current.options.indexOf(opt)) return "bg-error text-error-content";
    return "bg-base-300 text-base-content/75";
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
      <span class="loading loading-spinner loading-lg"></span>
      <span class="text-sm text-base-content/75">Loading learn session...</span>
    </div>
  {:else if loadError}
    <div role="alert" class="alert mt-4"><span class="text-sm">{loadError}</span></div>
  {:else}
    {#if !done}
      <div class="flex items-center gap-3">
        <a
          href="/decks/{id}"
          class="btn btn-circle btn-ghost btn-sm shrink-0"
          title="Exit to {title}"
          aria-label="Exit learn mode"
        >
          <X class="size-5" />
        </a>
        <div class="min-w-0 flex-1">
          <div class="flex items-baseline justify-between text-xs text-base-content/75">
            <span class="truncate font-medium">{progressLabel}</span>
            <span class="font-mono">{progress}%</span>
          </div>
          <div class="mt-1 h-1.5 overflow-hidden rounded-full bg-base-300">
            <div class="h-full rounded-full bg-primary transition-all duration-500" style="width: {progress}%"></div>
          </div>
        </div>
        <button
          type="button"
          class="btn btn-circle btn-ghost btn-sm shrink-0"
          title="Restart session"
          aria-label="Restart session"
          onclick={restart}
        >
          <RotateCcw class="size-4" />
        </button>
      </div>
    {/if}

    {#if empty}
      <div class="alert mt-6"><span class="text-sm">This deck has no cards yet.</span></div>
    {/if}

    {#if !done && !empty && current}
      <div class="mt-6">
        <p class="min-h-24 text-center text-2xl font-bold text-balance">{current.front}</p>

        {#if current.type === "mc"}
          <div class="mt-6 grid gap-2">
            {#each current.options as opt, i (i)}
              <button
                type="button"
                onclick={() => answerMC(i)}
                disabled={!!feedback}
                class="flex w-full items-center gap-3 rounded-xl border p-3 text-left transition-all {mcButtonClass(opt)}"
              >
                <span
                  class="flex size-7 shrink-0 items-center justify-center rounded-md font-mono text-xs font-bold {mcBadgeClass(opt)}"
                >
                  {String.fromCharCode(65 + i)}
                </span>
                <span class="flex-1">{opt.text}</span>
              </button>
            {/each}
            {#if feedback && !feedback.correct}
              <button
                type="button"
                onclick={cont}
                class="mt-1 w-full rounded-xl bg-error/10 p-3 text-center text-sm"
              >
                <span class="text-base-content/75">Correct answer: </span>
                <span class="font-bold">{current.back}</span>
              </button>
            {/if}
          </div>
        {:else if current.type === "tf"}
          <div class="mt-6">
            <div class="rounded-xl border border-base-300 bg-base-100 p-5 text-center">
              <p class="text-xl font-bold">{current.tfStatement}</p>
            </div>
            <div class="mt-2 grid grid-cols-2 gap-2">
              <button
                type="button"
                onclick={() => answerTF(true)}
                disabled={!!feedback}
                class="flex items-center justify-center gap-2 rounded-xl border p-3 font-semibold transition-all {feedback && current.tfAnswer
                  ? 'border-success bg-success/10'
                  : feedback && current.picked === true
                    ? 'border-error bg-error/10 learn-shake'
                    : feedback
                      ? 'border-base-300 opacity-60'
                      : 'border-base-300 bg-base-100 hover:border-primary/60'}"
              >
                <span
                  class="flex size-7 shrink-0 items-center justify-center rounded-md bg-base-300 font-mono text-xs font-bold text-base-content/75"
                  >T</span
                >
                <span>True</span>
              </button>
              <button
                type="button"
                onclick={() => answerTF(false)}
                disabled={!!feedback}
                class="flex items-center justify-center gap-2 rounded-xl border p-3 font-semibold transition-all {feedback &&
                !current.tfAnswer
                  ? 'border-success bg-success/10'
                  : feedback && current.picked === false
                    ? 'border-error bg-error/10 learn-shake'
                    : feedback
                      ? 'border-base-300 opacity-60'
                      : 'border-base-300 bg-base-100 hover:border-primary/60'}"
              >
                <span
                  class="flex size-7 shrink-0 items-center justify-center rounded-md bg-base-300 font-mono text-xs font-bold text-base-content/75"
                  >F</span
                >
                <span>False</span>
              </button>
            </div>
            {#if feedback && !feedback.correct}
              <button
                type="button"
                onclick={cont}
                class="mt-2 w-full rounded-xl bg-error/10 p-3 text-center text-sm"
              >
                <span class="text-base-content/75">Correct answer: </span>
                <span class="font-bold">{current.back}</span>
              </button>
            {/if}
          </div>
        {:else if current.type === "written"}
          <div class="mt-6">
            <input
              type="text"
              placeholder="Type the answer"
              autocomplete="off"
              class="input input-lg input-bordered w-full text-center"
              class:input-success={feedback?.correct}
              class:input-error={feedback && !feedback.correct}
              bind:value={answer}
              disabled={!!feedback}
            />
            {#if feedback && !feedback.correct}
              <div class="mt-3 text-center text-sm">
                <p>
                  <span class="text-base-content/75">Correct answer: </span><span class="font-bold"
                    >{current.back}</span
                  >
                </p>
                {#if answer}
                  <p class="mt-0.5 text-base-content/75">You wrote: <span>{answer}</span></p>
                {/if}
              </div>
            {/if}
            {#if !feedback}
              <div class="mt-3">
                <button type="button" class="btn btn-primary btn-block" onclick={submitWritten}>Check</button>
                <button type="button" class="btn btn-ghost btn-block btn-sm mt-1" onclick={dontKnow}
                  >I don't know</button
                >
              </div>
            {/if}
            {#if feedback && !feedback.correct}
              <button type="button" class="btn btn-primary btn-block mt-3" onclick={cont}>Continue</button>
            {/if}
          </div>
        {/if}
      </div>
    {/if}

    {#if done}
      <div class="mt-8 text-center">
        <div class="mx-auto flex size-16 items-center justify-center rounded-full bg-primary/15 text-primary">
          <Trophy class="size-7" />
        </div>
        <h2 class="mt-3 text-xl font-bold">Session complete</h2>
        <p class="mt-4 text-sm text-base-content/75">
          <span class="font-mono font-bold text-base-content"
            >{total ? Math.round((correctFirst / total) * 100) : 0}</span
          ><span class="font-mono">% correct first try · </span><span class="font-mono">{attempts}</span><span
            class="font-mono"
          >
            attempts</span
          >
        </p>
        {#if missedList().length > 0}
          <div class="mt-5 text-left">
            <ul class="divide-y divide-base-300 overflow-hidden rounded-xl border border-base-300">
              {#each missedList() as m, i (i)}
                <li class="flex items-baseline justify-between gap-4 px-4 py-2.5 text-sm">
                  <span class="font-semibold">{m.front}</span>
                  <span class="text-right text-base-content/75">{m.back}</span>
                </li>
              {/each}
            </ul>
          </div>
        {/if}
        <div class="mt-5 flex justify-center gap-2">
          <button type="button" class="btn btn-primary" onclick={restart}>Study again</button>
          <a href="/decks/{id}" class="btn btn-ghost">Flashcards</a>
        </div>
      </div>
    {/if}
  {/if}
</div>
