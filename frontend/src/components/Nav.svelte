<script>
  import Plus from "@lucide/svelte/icons/plus";
  import Bell from "@lucide/svelte/icons/bell";
  import LogOut from "@lucide/svelte/icons/log-out";
  import { session, logout } from "../lib/session.js";
  import { goto } from "../lib/router.js";

  async function doLogout(e) {
    e.preventDefault();
    await logout();
    goto("/");
  }
</script>

<nav class="navbar bg-base-200 shadow-sm">
  <div class="navbar-start">
    <a href="/" class="btn btn-ghost text-xl">Memri</a>
  </div>
  <div class="navbar-end">
    {#if $session.status === "authed"}
      <a href="/decks/new" class="btn btn-ghost btn-circle" aria-label="New deck" title="New deck">
        <Plus class="size-5" />
      </a>
      <a href="/notifications" class="btn btn-ghost btn-circle" aria-label="Notifications">
        <span class="indicator">
          <Bell class="size-5" />
          <span class="badge badge-xs badge-primary indicator-item">{$session.unread}</span>
        </span>
      </a>
      <button class="btn btn-ghost btn-circle" aria-label="Logout" onclick={doLogout}>
        <LogOut class="size-5" />
      </button>
    {/if}
  </div>
</nav>
