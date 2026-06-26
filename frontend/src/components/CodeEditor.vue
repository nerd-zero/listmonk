<template>
  <div ref="editorEl" class="code-editor" />
</template>

<script setup lang="ts">
import {
  ref, watch, onMounted, onBeforeUnmount, nextTick,
} from 'vue';
import { EditorState } from '@codemirror/state';
import {
  EditorView, keymap, highlightActiveLine, lineNumbers, highlightActiveLineGutter,
} from '@codemirror/view';
import { markdown } from '@codemirror/lang-markdown';
import { javascript } from '@codemirror/lang-javascript';
import { css } from '@codemirror/lang-css';
import { html } from '@codemirror/lang-html';
import {
  defaultKeymap, history, historyKeymap, indentWithTab,
} from '@codemirror/commands';
import { defaultHighlightStyle, syntaxHighlighting, bracketMatching } from '@codemirror/language';
import { search, searchKeymap, highlightSelectionMatches } from '@codemirror/search';
import { vsCodeLight } from './editor-theme';

const props = withDefaults(defineProps<{
  modelValue?: string;
  lang?: string;
  disabled?: boolean;
}>(), {
  modelValue: '',
  lang: 'html',
  disabled: false,
});

const emit = defineEmits(['update:modelValue']);

const editorEl = ref<HTMLElement | null>(null);
let editor: EditorView | null = null;
let internalUpdate = false;

onMounted(() => {
  const onUpdate = EditorView.updateListener.of((update) => {
    if (update.docChanged) {
      internalUpdate = true;
      emit('update:modelValue', update.state.doc.toString());
    }
  });

  let langs: any[] = [];
  switch (props.lang) {
    case 'html': langs = [html()]; break;
    case 'css': langs = [css()]; break;
    case 'javascript': langs = [javascript()]; break;
    case 'markdown': langs = [markdown()]; break;
    default: langs = [html()];
  }

  const stateCfg = EditorState.create({
    doc: props.modelValue,
    extensions: [
      EditorView.baseTheme({}),
      ...langs,
      history(),
      highlightActiveLine(),
      bracketMatching(),
      highlightSelectionMatches(),
      lineNumbers(),
      highlightActiveLineGutter(),
      keymap.of([...defaultKeymap, ...historyKeymap, ...searchKeymap, indentWithTab]),
      EditorState.readOnly.of(props.disabled),
      EditorView.editable.of(!props.disabled),
      syntaxHighlighting(defaultHighlightStyle, { fallback: true }),
      EditorView.lineWrapping,
      vsCodeLight,
      search({ top: true }),
      onUpdate,
    ],
  });

  editor = new EditorView({ state: stateCfg, parent: editorEl.value! });

  nextTick(() => {
    window.setTimeout(() => { editor?.focus(); }, 100);
  });
});

onBeforeUnmount(() => {
  editor?.destroy();
});

watch(
  () => props.modelValue,
  (val) => {
    if (!internalUpdate && editor) {
      editor.dispatch({
        changes: { from: 0, to: editor.state.doc.length, insert: val },
      });
    }
    internalUpdate = false;
  },
);
</script>
