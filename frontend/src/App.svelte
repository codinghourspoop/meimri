<script>
  import { onMount } from "svelte";
  import { route, goto } from "./lib/router.js";
  import { session, loadMe } from "./lib/session.js";
  import Nav from "./components/Nav.svelte";
  import Home from "./pages/Home.svelte";
  import Login from "./pages/Login.svelte";
  import Register from "./pages/Register.svelte";
  import Recover from "./pages/Recover.svelte";
  import RecoveryCodes from "./pages/RecoveryCodes.svelte";
  import NewDeck from "./pages/NewDeck.svelte";
  import DeckView from "./pages/DeckView.svelte";
  import Learn from "./pages/Learn.svelte";
  import Notifications from "./pages/Notifications.svelte";
  import { matchRoute } from "./lib/router.js";

  onMount(() => {
    loadMe();
    // SPA link interception: same-origin <a href> navigates without reload.
    const onClick = (e) => {
      const a = e.target.closest?.("a[href]");
      if (!a) return;
      const href = a.getAttribute("href");
      if (!href || !href.startsWith("/") || href.startsWith("//")) return;
      if (e.metaKey || e.ctrlKey || e.shiftKey || e.altKey || e.button !== 0) return;
      if (a.target && a.target !== "_self") return;
      e.preventDefault();
      goto(href);
    };
    document.addEventListener("click", onClick);
    return () => document.removeEventListener("click", onClick);
  });

  let path = $derived($route);
</script>

<svelte:head>
  <title>memri</title>
</svelte:head>

<div class="min-h-screen bg-base-100 font-sans text-base-content antialiased">
  <Nav />

  <main class="mx-auto max-w-5xl px-4 py-8">
  {#key path}
    {#if $session.status === "loading"}
    <div class="flex h-80 flex-col items-center justify-center gap-4">
      <span class="loading loading-spinner loading-lg"></span>
    </div>
  {:else if path === "/"}
    <Home />
  {:else if path === "/login"}
    <Login />
  {:else if path === "/register"}
    <Register />
  {:else if path === "/recover"}
    <Recover />
  {:else if path === "/recovery-codes"}
    <RecoveryCodes />
  {:else if path === "/decks/new"}
    <NewDeck />
  {:else if path === "/notifications"}
    <Notifications />
  {:else if matchRoute("/decks/:id/learn", path)}
    <Learn id={matchRoute("/decks/:id/learn", path).id} />
  {:else if matchRoute("/decks/:id", path)}
    <DeckView id={matchRoute("/decks/:id", path).id} />
  {:else}
    <div role="alert" class="alert mt-4">
      <span class="text-sm">Page not found. <a href="/" class="link">Go home</a>.</span>
    </div>
    {/if}
    {/key}
  </main>
</div>
