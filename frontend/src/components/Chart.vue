<template>
  <section class="chart" ref="rootEl">
    <canvas class="chart-canvas" />
  </section>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import Chart from 'chart.js/auto';

const DEFAULT_DONUT = {
  type: 'doughnut',
  data: {},
  options: {
    responsive: true,
    cutout: '70%',
    maintainAspectRatio: false,
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
        callbacks: {
          label: (item: any) => {
            const data = item.chart.data.datasets[item.datasetIndex];
            const total = data.data.reduce((acc: number, val: number) => acc + val, 0);
            const val = data.data[item.dataIndex];
            const percentage = ((val / total) * 100).toFixed(2);
            return `${val} (${percentage}%)`;
          },
        },
      },
    },
  },
};

const DEFAULT_LINE = {
  type: 'line',
  data: {},
  options: {
    responsive: true,
    lineTension: 0.5,
    maintainAspectRatio: false,
    interaction: { intersect: false, axis: 'index' },
    plugins: {
      legend: { display: false },
      tooltip: {
        backgroundColor: '#fff',
        borderColor: '#ddd',
        borderWidth: 1,
        bodyColor: '#666',
        displayColors: true,
        bodyFont: { size: 15 },
        bodySpacing: 10,
        padding: 10,
      },
    },
    scales: {
      x: { grid: { display: false } },
      y: { grid: { display: false }, ticks: { precision: 0 } },
    },
  },
};

const DEFAULT_BAR = {
  type: 'bar',
  data: {},
  options: {
    responsive: true,
    indexAxis: 'y',
    barThickness: 40,
    maintainAspectRatio: false,
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
    scales: {
      x: { grid: { display: false } },
      y: { grid: { display: false } },
    },
  },
};

const props = withDefaults(defineProps<{
  data?: object;
  type?: string;
  onClick?:(e: unknown) => void;
}>(), {
  data: () => ({}),
  type: 'line',
  onClick: () => () => {},
});

const rootEl = ref<HTMLElement | null>(null);

onMounted(() => {
  if (!props.data) return;

  const ctx = rootEl.value?.querySelector('.chart-canvas') as HTMLCanvasElement;

  let def: Record<string, unknown> = {};
  switch (props.type) {
    case 'donut': def = DEFAULT_DONUT; break;
    case 'bar': def = DEFAULT_BAR; break;
    default: def = DEFAULT_LINE; break;
  }

  const conf: any = { ...def, data: props.data };
  if (props.onClick) {
    conf.options.onClick = props.onClick;
  }
  // eslint-disable-next-line no-new
  new Chart(ctx, conf);
});
</script>

<style scoped>
.chart { position: relative; width: 100%; height: 100%; }
</style>
