<script>

    import { Badge } from "../lib/components/ui/badge/index.js";
    import Bell from "@lucide/svelte/icons/bell";
    import { buttonVariants } from "$lib/components/ui/button/index.js";
    import { Input } from "$lib/components/ui/input/index.js";
    import { Label } from "$lib/components/ui/label/index.js";

  import * as Dialog from "$lib/components/ui/dialog/index.js";
  import { onMount } from "svelte";
  import { api } from "../lib/api.js";
  import { session, loadMe } from "../lib/session.js";
  import { Button } from "../lib/components/ui/button/index.js";
  import { Card, CardContent } from "../lib/components/ui/card/index.js";
  import { Alert, AlertDescription } from "../lib/components/ui/alert/index.js";
  import { Spinner } from "../lib/components/ui/spinner/index.js";

  const POLL_MS = 30_000;

  let items = $state(null);
  let err = $state("");
  let busy = $state(false);
  let fetching = false;

  // Fetches the list and syncs the unread badge. Returns the error, or null on
  // success, so callers decide whether it is worth surfacing.
  async function refresh() {
    if (fetching) return null;
    fetching = true;
    try {
      const data = await api.notifications();
      items = data.notifications ?? [];
      session.update((s) => ({ ...s, unread: data.unread ?? 0 }));
      return null;
    } catch (e) {
      return e;
    } finally {
      fetching = false;
    }
  }

  // Background poll: a failed tick is ignored, the next one may well succeed.
  function poll() {
    if (document.hidden || busy) return;
    refresh();
  }

  onMount(() => {
    refresh().then((e) => {
      if (!e) return;
      err = e?.payload?.error || "Failed to load notifications.";
      items = [];
    });
    const id = setInterval(poll, POLL_MS);
    // Catch up as soon as the tab comes back, rather than waiting out the tick.
    document.addEventListener("visibilitychange", poll);
    return () => {
      clearInterval(id);
      document.removeEventListener("visibilitychange", poll);
    };
  });

  async function markAll(e) {
    e.preventDefault();
    busy = true;
    try {
      await api.markNotificationsRead();
      await loadMe();
      const e2 = await refresh();
      if (e2) err = e2?.payload?.error || "Failed to refresh notifications.";
    } catch (e3) {
      err = e3?.payload?.error || "Failed to mark notifications read.";
    } finally {
      busy = false;
    }
  }
</script>

<Dialog.Root>
       <Dialog.Trigger
        >        <Button variant="ghost" size="icon" aria-label="Notifications" class="relative">
          <Bell class="size-5" />
          <Badge class="absolute -top-1 -right-1 h-4 min-w-4 px-1">{$session.unread}</Badge>
        </Button></Dialog.Trigger
       >
       <Dialog.Content class="sm:max-w-md">
        <Dialog.Header>
         <Dialog.Title>Notifications</Dialog.Title>
{#if err}

  <Alert variant="destructive" class="mt-4"><AlertDescription class="text-sm">{err}</AlertDescription></Alert>
{/if}
{#if items === null}
  <div class="flex h-40 flex-col items-center justify-center gap-4">
    <Spinner class="size-8" />
  </div>
{:else if items.length === 0}
  <Alert class="mt-4"><AlertDescription class="text-sm">No notifications.</AlertDescription></Alert>
{:else}
  <ul class="mt-4 space-y-2">
    {#each items as n, i (i)}
      <li>
        <Card class={!n.is_read ? "" : "opacity-75"}>
          <CardContent class="p-3 text-sm">
            {n.body}
          </CardContent>
        </Card>
      </li>
    {/each}
  </ul>

{/if}


        </Dialog.Header>
        <Dialog.Footer class="sm:justify-start">
         <Dialog.Close class={buttonVariants({ variant: "secondary" })}
          >Close</Dialog.Close
                 >

           <button
             type="button"
             class={buttonVariants({ variant: "secondary" })}
             disabled={busy}
             onclick={markAll}>Mark all read</button
           >
        </Dialog.Footer>
       </Dialog.Content>
      </Dialog.Root>
