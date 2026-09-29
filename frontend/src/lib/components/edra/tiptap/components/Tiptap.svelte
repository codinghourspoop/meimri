<script lang="ts">
	import type { Editor } from '../Editor.ts';
	import { setEditor } from './editorContext.js';
	import { type Snippet } from 'svelte';

	let { editor, children }: { editor: Editor | undefined; children?: Snippet } = $props();

	// Context must be set during init, but `editor` may arrive later.
	// A stable box with a getter always reads the latest prop value.
	const box = {
		get current() {
			return editor;
		}
	};
	setEditor(box);
</script>

{#if editor && children}
	{@render children()}
{/if}
