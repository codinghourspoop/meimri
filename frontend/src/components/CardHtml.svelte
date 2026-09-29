<script>
  let { html = "", class: className = "", tag = "div" } = $props();
</script>

<!-- Card fields are authored in a rich-text editor and stored as HTML,
     so they must be rendered as markup rather than escaped text. -->
<svelte:element this={tag} class="card-html {className}">{@html html}</svelte:element>

<style>
  .card-html :global(ul) {
    list-style-type: disc;
    padding-left: 1.25rem;
    text-align: left;
  }
  .card-html :global(ol) {
    list-style-type: decimal;
    padding-left: 1.25rem;
    text-align: left;
  }
  .card-html :global(img) {
    max-width: 100%;
    border-radius: 0.5rem;
  }
  .card-html :global(p + p),
  .card-html :global(li + li) {
    margin-top: 0.25rem;
  }
  .card-html :global(ul),
  .card-html :global(ol),
  .card-html :global(blockquote) {
    text-align: left;
  }
</style>