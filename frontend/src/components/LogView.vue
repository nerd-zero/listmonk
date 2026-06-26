<template>
  <section class="log-view">
    <div v-if="loading" class="flex justify-center p-8">
      <PvProgressSpinner style="width:2rem;height:2rem" />
    </div>
    <div class="lines" ref="linesEl">
      <template v-for="(l, i) in lines">
        <template v-if="l">
          <span :set="line = splitLine(String(l))" :key="i" class="line">
            <span class="timestamp">{{ line.timestamp }}&nbsp;</span>
            <span v-if="line.file !== '*'" class="file">{{ line.file }}:&nbsp;</span>
            <span class="log-message">{{ line.message }}</span>
          </span>
        </template>
      </template>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue';

// Regexp for splitting log lines: 2021/05/01 00:00:00.000000 init.go:99: message
const reFormatLine = /^([0-9\s:/]+\.[0-9]{6}) (.+?\.go:[0-9]+|\*):\s(.+)$/;

const props = withDefaults(defineProps<{
  loading?: boolean;
  lines?: unknown[];
}>(), {
  loading: false,
  lines: () => [],
});

const linesEl = ref<HTMLElement | null>(null);

function splitLine(l: string) {
  const parts = l.split(reFormatLine);
  if (parts.length !== 5) {
    return { timestamp: '', file: '', message: l };
  }
  return { timestamp: parts[1], file: parts[2], message: parts[3] };
}

watch(
  () => props.lines,
  () => {
    nextTick(() => {
      if (linesEl.value) {
        linesEl.value.scrollTop = linesEl.value.scrollHeight;
      }
    });
  },
);
</script>
