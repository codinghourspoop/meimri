<script>
  import { api } from "../lib/api.js";
  import { loadMe } from "../lib/session.js";
  import { goto } from "../lib/router.js";
  import { Button } from "../lib/components/ui/button/index.js";
  import { Input } from "../lib/components/ui/input/index.js";
  import { Label } from "../lib/components/ui/label/index.js";
  import { Card, CardContent, CardHeader, CardTitle } from "../lib/components/ui/card/index.js";
  import { Alert, AlertDescription } from "../lib/components/ui/alert/index.js";

  let username = $state("");
  let code = $state("");
  let password = $state("");
  let err = $state("");
  let busy = $state(false);

  async function submit(e) {
    e.preventDefault();
    err = "";
    busy = true;
    try {
      await api.recover(username.trim(), code.trim(), password);
      await loadMe();
      goto("/");
    } catch (e2) {
      err = e2?.payload?.error || "Recovery failed.";
    } finally {
      busy = false;
    }
  }
</script>

<Card class="mx-auto max-w-sm">
  <CardHeader>
    <CardTitle class="text-2xl">Recover account</CardTitle>
  </CardHeader>
  <CardContent>
    {#if err}
      <Alert variant="destructive" class="mb-4"><AlertDescription>{err}</AlertDescription></Alert>
    {/if}
    <form onsubmit={submit} class="space-y-4">
      <div class="space-y-2">
        <Label for="rec-username">Username</Label>
        <Input id="rec-username" name="username" required placeholder="Username" bind:value={username} />
      </div>
      <div class="space-y-2">
        <Label for="rec-code">Recovery code</Label>
        <Input
          id="rec-code"
          name="code"
          required
          class="font-mono"
          placeholder="Recovery code"
          bind:value={code}
        />
      </div>
      <div class="space-y-2">
        <Label for="rec-password">New password</Label>
        <Input
          id="rec-password"
          name="password"
          type="password"
          required
          minlength="8"
          placeholder="New password"
          bind:value={password}
        />
      </div>
      <Button type="submit" class="w-full" disabled={busy}>Reset password</Button>
    </form>
  </CardContent>
</Card>
