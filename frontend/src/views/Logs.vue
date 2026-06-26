<template>
  <div class="logs-page">
    <div class="page-header">
      <h1 class="page-title">{{ $t('logs.title') }}</h1>
    </div>
    <div class="table-card">
      <log-view :loading="loading.logs" :lines="lines" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { storeToRefs } from 'pinia';
import { useMainStore } from '../store';
import { useGlobal } from '../composables/useGlobal';
import LogView from '../components/LogView.vue';

const { $api } = useGlobal();
const { loading } = storeToRefs(useMainStore());

const lines = ref<unknown[]>([]);
let pollId: ReturnType<typeof setInterval> | null = null;

function getLogs() {
  $api.getLogs().then((data: unknown) => {
    lines.value = data as unknown[];
  });
}

onMounted(() => {
  getLogs();
  pollId = setInterval(() => getLogs(), 10000);
});

onUnmounted(() => {
  if (pollId !== null) clearInterval(pollId);
});
</script>

<style scoped lang="scss">
.logs-page { display: flex; flex-direction: column; gap: 1.5rem; }
</style>
