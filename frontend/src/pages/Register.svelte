<script>
  import { api } from "../lib/api.js";
  import { loadMe, pendingRecoveryCodes } from "../lib/session.js";
  import { goto } from "../lib/router.js";
  import { Button } from "../lib/components/ui/button/index.js";
  import { Input } from "../lib/components/ui/input/index.js";
  import { Label } from "../lib/components/ui/label/index.js";
  import { Card, CardContent, CardHeader, CardTitle } from "../lib/components/ui/card/index.js";
  import { Alert, AlertDescription } from "../lib/components/ui/alert/index.js";

  let username = $state("");
  let displayName = $state("");
  let password = $state("");
  let err = $state("");
  let busy = $state(false);

  async function submit(e) {
    e.preventDefault();
    err = "";
    busy = true;
    try {
      const data = await api.register(username.trim(), displayName.trim(), password);
      pendingRecoveryCodes.set(data.recovery_codes ?? []);
      await loadMe();
      goto("/recovery-codes");
    } catch (e2) {
      err = e2?.payload?.error || "Registration failed.";
    } finally {
      busy = false;
    }
  }
</script>

<Card class="mx-auto max-w-sm">
  <CardHeader>
    <CardTitle class="text-2xl">Create account</CardTitle>
  </CardHeader>
  <CardContent>
    {#if err}
      <Alert variant="destructive" class="mb-4"><AlertDescription>{err}</AlertDescription></Alert>
    {/if}
    <form onsubmit={submit} class="space-y-4">
      <div class="space-y-2">
        <Label for="reg-username">Username (3-32 chars, letters/numbers/_/-)</Label>
        <Input
          id="reg-username"
          name="username"
          required
          pattern="[a-zA-Z0-9_-]{3,32}"
          placeholder="Username"
          bind:value={username}
        />
      </div>
      <div class="space-y-2">
        <Label for="reg-display">Display name</Label>
        <Input id="reg-display" name="display_name" required placeholder="Display name" bind:value={displayName} />
      </div>
      <div class="space-y-2">
        <Label for="reg-password">Password (min 8 chars)</Label>
        <Input
          id="reg-password"
          name="password"
          type="password"
          required
          minlength="8"
          placeholder="Password"
          bind:value={password}
        />
      </div>
      <Button type="submit" class="w-full" disabled={busy}>Create account</Button>
    </form>
  </CardContent>
</Card>
