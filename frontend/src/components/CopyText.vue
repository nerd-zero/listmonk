<template>
  <a href="#" class="copy-text" ref="textEl" @click.prevent="onClick">
    <template v-if="!hideText">{{ text }}</template>
    <i class="pi pi-copy" />
  </a>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { useGlobal } from '../composables/useGlobal';

const props = withDefaults(defineProps<{
  text?: string;
  hideText?: boolean;
}>(), {
  text: '',
  hideText: false,
});

const { t } = useI18n();
const { $utils } = useGlobal();

function onClick(e: Event) {
  e.preventDefault();
  e.stopPropagation();

  const input = document.createElement('input');
  input.setAttribute('type', 'text');
  input.style.opacity = '0';
  input.value = props.text;
  document.body.appendChild(input);
  input.select();
  document.execCommand('copy');
  document.body.removeChild(input);

  $utils.toast(t('globals.messages.copied'));
}
</script>
