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

<script lang="ts">
import { defineComponent } from 'vue';
import { mapState } from 'pinia';
import { useMainStore } from '../store';
import LogView from '../components/LogView.vue';

export default defineComponent({
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
});
</script>

<style scoped lang="scss">
.logs-page { display: flex; flex-direction: column; gap: 1.5rem; }

</style>
