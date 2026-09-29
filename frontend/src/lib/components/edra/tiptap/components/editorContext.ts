import { getContext, setContext } from 'svelte';

import type { Editor } from '../Editor.ts';

const EDITOR_CTX = Symbol('editor');

interface EditorBox {
	current: Editor | undefined;
}

export const getEditor = (): Editor => {
	const box = getContext<EditorBox>(EDITOR_CTX);

	if (!box?.current) {
		throw new Error('No editor found in context. Did you wrap your component in <Tiptap>?');
	}

	return box.current;
};

export const setEditor = (box: EditorBox): void => {
	setContext(EDITOR_CTX, box);
};
