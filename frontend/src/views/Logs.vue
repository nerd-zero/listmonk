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

<script>
import { mapState } from 'pinia';
import { useMainStore } from '../store';
import LogView from '../components/LogView.vue';

export default {
  components: {
    LogView,
  },

  data() {
    return {
      lines: [],
      pollId: null,
    };
  },

  methods: {
    getLogs() {
      this.$api.getLogs().then((data) => {
        this.lines = data;
      });
    },
  },

  computed: {
    ...mapState(useMainStore, ['logs', 'loading']),
  },

  mounted() {
    this.getLogs();

    // Update the logs every 10 seconds.
    this.pollId = setInterval(() => this.getLogs(), 10000);
  },

  unmounted() {
    clearInterval(this.pollId);
  },
};
</script>

<style scoped lang="scss">
.logs-page { display: flex; flex-direction: column; gap: 1.5rem; }
.page-header { display: flex; align-items: center; }
.page-title { font-size: 1.5rem; font-weight: 700; color: #0f172a; margin: 0; }
.table-card { background: #fff; border: 1px solid #e2e8f0; border-radius: 12px; overflow: hidden; }
</style>
