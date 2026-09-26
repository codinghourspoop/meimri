<script>
  import User from "@lucide/svelte/icons/user";
  import Lock from "@lucide/svelte/icons/lock";
  import { api } from "../lib/api.js";
  import { loadMe } from "../lib/session.js";
  import { goto } from "../lib/router.js";
  import { Button } from "../lib/components/ui/button/index.js";
  import { Input } from "../lib/components/ui/input/index.js";
  import { Card, CardContent, CardHeader, CardTitle } from "../lib/components/ui/card/index.js";
  import { Alert, AlertDescription } from "../lib/components/ui/alert/index.js";

  let username = $state("");
  let password = $state("");
  let err = $state("");
  let busy = $state(false);

  async function submit(e) {
    e.preventDefault();
    err = "";
    busy = true;
    try {
      await api.login(username.trim(), password);
      await loadMe();
      goto("/");
    } catch (e2) {
      err = e2?.payload?.error || "Invalid username or password.";
    } finally {
      busy = false;
    }
  }
</script>

<Card class="mx-auto max-w-sm">
  <CardHeader>
    <CardTitle class="text-2xl">Login</CardTitle>
  </CardHeader>
  <CardContent>
    {#if err}
      <Alert variant="destructive" class="mb-4"><AlertDescription>{err}</AlertDescription></Alert>
    {/if}
    <form onsubmit={submit} class="space-y-4">
      <div class="relative">
        <User class="text-muted-foreground pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2" />
        <Input name="username" required placeholder="Username" class="pl-9" bind:value={username} />
      </div>
      <div class="relative">
        <Lock class="text-muted-foreground pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2" />
        <Input
          name="password"
          placeholder="Password"
          type="password"
          required
          class="pl-9"
          bind:value={password}
        />
      </div>
      <Button type="submit" class="w-full" disabled={busy}>Login</Button>
    </form>
    <p class="mt-4 text-center text-sm">
      <Button href="/recover" variant="link" class="h-auto p-0 text-sm">Recover with code</Button>
      <span class="text-muted-foreground"> · </span>
      <Button href="/register" variant="link" class="h-auto p-0 text-sm">Register</Button>
    </p>
  </CardContent>
</Card>
