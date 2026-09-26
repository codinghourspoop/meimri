<script>
  import Plus from "@lucide/svelte/icons/plus";
  import LogOut from "@lucide/svelte/icons/log-out";
  import Sun from "@lucide/svelte/icons/sun";
  import Moon from "@lucide/svelte/icons/moon";
  import { toggleMode } from "mode-watcher";
  import { Button } from "../lib/components/ui/button/index.js";
  import { session, logout } from "../lib/session.js";
  import { goto } from "../lib/router.js";
  import Notifications from "./Notifications.svelte";

  async function doLogout(e) {
    e.preventDefault();
    await logout();
    goto("/");
  }
</script>

<header class="border-border bg-card">
  <div class="mx-auto flex h-14 max-w-5xl items-center justify-between px-4">
    <a href="/" class="text-lg font-semibold tracking-tight">Memri</a>
    <div class="flex items-center gap-1">
      <Button onclick={toggleMode} variant="ghost" size="icon" aria-label="Toggle theme">
        <Sun class="size-5 scale-100 rotate-0 transition-all dark:scale-0 dark:-rotate-90" />
        <Moon class="absolute size-5 scale-0 rotate-90 transition-all dark:scale-100 dark:rotate-0" />
        <span class="sr-only">Toggle theme</span>
      </Button>
      {#if $session.status === "authed"}
        <Button href="/decks/new" variant="ghost" size="icon" aria-label="New deck" title="New deck">
          <Plus class="size-5" />
        </Button>

<Notifications/>

        <Button variant="ghost" size="icon" aria-label="Logout" onclick={doLogout}>
          <LogOut class="size-5" />
        </Button>
      {/if}
    </div>
  </div>
</header>
