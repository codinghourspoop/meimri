<script>
  import Quill from "quill";
  import "quill/dist/quill.snow.css";
  import { onMount } from "svelte";

  let {
    value = $bindable(""),
    label = "",
    placeholder = "Write…",
    fieldKey = "",
    onReady = null,
    onActive = null,
    onSel = null,
  } = $props();

  let container = $state(null);
  let quill = $state(null);

  function currentHtml() {
    if (!quill) return "";
    return quill.getText().trim() ? quill.root.innerHTML : "";
  }

  onMount(() => {
    quill = new Quill(container, {
      theme: "snow",
      placeholder,
      modules: { toolbar: false },
    });
    if (value) quill.root.innerHTML = value;
    quill.on("text-change", () => {
      value = currentHtml();
      queueMicrotask(() => onSel?.());
    });
    quill.on("selection-change", (range) => {
      // Defer: this emit can fire re-entrantly from inside Quill's own
      // update cycle — never write Svelte state synchronously here.
      const r = range;
      queueMicrotask(() => {
        if (r) onActive?.(quill, fieldKey);
        onSel?.();
      });
    });
    onReady?.(quill, fieldKey);
  });

  // Push external value changes in, but never clobber while the user is typing.
  $effect(() => {
    const v = value ?? "";
    if (quill && !quill.hasFocus() && quill.root.innerHTML !== v) {
      quill.root.innerHTML = v;
    }
  });
</script>

<div>
  {#if label}
    <span class="text-muted-foreground mb-1 block text-xs font-medium">{label}</span>
  {/if}
  <div class="rich-quill">
    <div bind:this={container}></div>
  </div>
</div>

<style>
  .rich-quill {
    color: var(--color-foreground);
    font-size: 0.875rem;
  }
  .rich-quill :global(.ql-container.ql-snow) {
    border-color: var(--color-border);
    border-radius: var(--radius-lg);
    font-size: inherit;
  }
  .rich-quill :global(.ql-editor) {
    min-height: 2.5rem;
  }
  .rich-quill :global(.ql-editor.ql-blank::before) {
    color: var(--color-muted-foreground);
    font-style: normal;
  }
  .rich-quill :global(.ql-editor img) {
    max-width: 100%;
    border-radius: 0.5rem;
  }
</style>
