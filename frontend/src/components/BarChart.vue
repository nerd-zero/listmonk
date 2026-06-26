<template>
  <section class="bar-chart" ref="rootEl">
    <canvas class="bar-chart-canvas" />
  </section>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import Chart from 'chart.js/auto';

const DEFAULT = {
  type: 'bar',
  data: {},
  options: {
    responsive: true,
    indexAxis: 'y',
    plugins: {
      legend: { display: false },
      tooltip: {
        backgroundColor: '#fff',
        borderColor: '#ddd',
        borderWidth: 1,
        titleColor: '#666',
        bodyColor: '#666',
        bodyFont: { size: 15 },
        bodySpacing: 10,
        padding: 10,
      },
    },
  },
};

const props = withDefaults(defineProps<{
  data?: object;
}>(), {
  data: () => ({}),
});

const rootEl = ref<HTMLElement | null>(null);

onMounted(() => {
  const ctx = rootEl.value?.querySelector('.bar-chart-canvas') as HTMLCanvasElement;
  // eslint-disable-next-line no-new
  new Chart(ctx, { ...DEFAULT, data: props.data as never });
});
</script>
