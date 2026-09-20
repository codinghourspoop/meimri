<script>
  import { onMount } from "svelte";
  import { api } from "../lib/api.js";
  import { session, loadMe } from "../lib/session.js";

  let items = $state(null);
  let err = $state("");
  let busy = $state(false);

  onMount(async () => {
    try {
      const data = await api.notifications();
      items = data.notifications ?? [];
      session.update((s) => ({ ...s, unread: data.unread ?? 0 }));
    } catch (e) {
      err = e?.payload?.error || "Failed to load notifications.";
      items = [];
    }
  });

  async function markAll(e) {
    e.preventDefault();
    busy = true;
    try {
      await api.markNotificationsRead();
      await loadMe();
      const data = await api.notifications();
      items = data.notifications ?? [];
    } catch (e2) {
      err = e2?.payload?.error || "Failed to mark notifications read.";
    } finally {
      busy = false;
    }
  }
</script>

<h1 class="text-2xl font-bold">Notifications</h1>
{#if err}
  <div role="alert" class="alert alert-error mt-4"><span class="text-sm">{err}</span></div>
{/if}
{#if items === null}
  <div class="flex h-40 flex-col items-center justify-center gap-4">
    <span class="loading loading-spinner loading-lg"></span>
  </div>
{:else if items.length === 0}
  <div role="alert" class="alert mt-4"><span class="text-sm">No notifications.</span></div>
{:else}
  <ul class="mt-4 space-y-2">
    {#each items as n, i (i)}
      <li
        class="card card-border p-3 text-sm"
        class:bg-base-100={!n.is_read}
        class:bg-base-200={n.is_read}
        class:opacity-75={n.is_read}
      >
        {n.body}
      </li>
    {/each}
  </ul>
  <div class="mt-4">
    <button type="button" class="btn btn-outline btn-sm" disabled={busy} onclick={markAll}>Mark all read</button>
  </div>
{/if}
