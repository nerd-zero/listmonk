<template>
  <section class="scrub-dashboard">
    <div class="page-header">
      <h1 class="page-title">
        <i class="pi pi-shield" />
        {{ $t('settings.scrub.dashboard') }}
      </h1>
      <PvButton severity="secondary" outlined icon="pi pi-refresh" :loading="loading.settings"
        :label="$t('globals.buttons.refresh')" @click="fetchStats" />
    </div>

    <!-- Not configured -->
    <div v-if="notConfigured" class="scrub-empty">
      <i class="pi pi-shield scrub-empty-icon" />
      <p class="scrub-empty-title">{{ $t('settings.scrub.name') }}</p>
      <p class="scrub-empty-desc">{{ $t('settings.scrub.notConfiguredDesc') }}</p>
      <PvButton severity="primary" icon="pi pi-cog" :label="$t('settings.scrub.goToSettings')"
        @click="$router.push({ name: 'settings' })" />
    </div>

    <!-- Error -->
    <div v-else-if="fetchError" class="scrub-error settings-card">
      <i class="pi pi-exclamation-triangle" style="color:#f59e0b" />
      <span>{{ fetchError }}</span>
      <PvButton severity="secondary" outlined size="small" icon="pi pi-refresh"
        :label="$t('globals.buttons.retry')" @click="fetchStats" />
    </div>

    <template v-else-if="rows.length">
      <!-- Activity chart with inline summary -->
      <div class="settings-card">
        <div class="scrub-chart-header">
          <span class="settings-section-label">Daily Activity — Last 30 Days</span>
          <div class="scrub-summary">
            <div class="scrub-summary-item">
              <span class="scrub-n">{{ fmt(periodTotal) }}</span>
              <span class="scrub-l">total</span>
            </div>
            <span class="scrub-summary-sep">·</span>
            <div class="scrub-summary-item">
              <span class="scrub-n">{{ fmt(todayCount) }}</span>
              <span class="scrub-l">today</span>
            </div>
            <span class="scrub-summary-sep">·</span>
            <div class="scrub-summary-item">
              <span class="scrub-n">{{ activeDays }}</span>
              <span class="scrub-l">of {{ rows.length }} active days</span>
            </div>
            <template v-if="peakDay.emailScrubs > 0">
              <span class="scrub-summary-sep">·</span>
              <div class="scrub-summary-item">
                <span class="scrub-n">{{ fmt(peakDay.emailScrubs) }}</span>
                <span class="scrub-l">peak on {{ peakDay.date }}</span>
              </div>
            </template>
          </div>
        </div>
        <div class="scrub-chart">
          <div v-for="row in rows" :key="row.date" class="scrub-chart-col"
            v-tooltip.top="row.date + ': ' + fmt(row.emailScrubs)">
            <div class="scrub-chart-bar"
              :style="{ height: barHeight(row.emailScrubs) }"
              :class="{ 'scrub-chart-bar--today': row.date === todayDate }" />
          </div>
        </div>
        <div class="scrub-chart-labels">
          <span>{{ rows[0]?.date }}</span>
          <span>{{ rows[rows.length - 1]?.date }}</span>
        </div>
      </div>

      <!-- Last 14 days table -->
      <div class="settings-card">
        <p class="settings-section-label mb-3">Recent Activity</p>
        <PvDataTable :value="recentRows" size="small" striped-rows>
          <PvColumn field="date" header="Date" />
          <PvColumn field="emailScrubs" header="Emails Scrubbed">
            <template #body="{ data }">
              <span :class="{ 'font-semibold': data.emailScrubs > 0 }">
                {{ fmt(data.emailScrubs) }}
              </span>
            </template>
          </PvColumn>
          <PvColumn header="Activity">
            <template #body="{ data }">
              <div v-if="data.emailScrubs > 0" class="scrub-inline-bar">
                <div class="scrub-inline-bar-fill"
                  :style="{ width: Math.round((data.emailScrubs / (peakDay.emailScrubs || 1)) * 100) + '%' }" />
              </div>
              <span v-else class="text-color-secondary text-sm">—</span>
            </template>
          </PvColumn>
        </PvDataTable>
      </div>
    </template>
  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import { mapState } from 'pinia';
import { useMainStore } from '../store';

export default defineComponent({
  name: 'ScrubDashboard',

  data() {
    return {
      rows: [],
      notConfigured: false,
      fetchError: null,
    };
  },

  computed: {
    ...mapState(useMainStore, ['loading']),

    todayDate() {
      return new Date().toISOString().slice(0, 10);
    },

    periodTotal() {
      return this.rows.reduce((s, r) => s + (r.emailScrubs || 0), 0);
    },

    todayCount() {
      return this.rows.find((r) => r.date === this.todayDate)?.emailScrubs ?? 0;
    },

    peakDay() {
      return this.rows.reduce((best, r) => (r.emailScrubs > (best.emailScrubs || 0) ? r : best), {});
    },

    activeDays() {
      return this.rows.filter((r) => r.emailScrubs > 0).length;
    },

    recentRows() {
      return [...this.rows].reverse().slice(0, 14);
    },
  },

  methods: {
    async fetchStats() {
      this.fetchError = null;
      this.notConfigured = false;
      try {
        const data = await this.$api.getScrubStats();
        this.rows = Array.isArray(data) ? data : (data.data ?? [data]);
      } catch (e) {
        const msg = e.response?.data?.message || '';
        if (e.response?.status === 400 && msg.toLowerCase().includes('not enabled')) {
          this.notConfigured = true;
        } else {
          this.fetchError = msg || this.$t('settings.scrub.statsError');
        }
      }
    },

    fmt(n) {
      if (n == null) return '—';
      return Number(n).toLocaleString();
    },

    barHeight(val) {
      const max = this.peakDay.emailScrubs || 1;
      const h = Math.max(2, Math.round((val / max) * 100));
      return `${h}%`;
    },
  },

  mounted() {
    this.fetchStats();
  },
});
</script>

<style scoped lang="scss">
.scrub-dashboard {
  max-width: 960px;
}

.scrub-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  padding: 4rem 2rem;
  text-align: center;
  background: var(--lm-surface);
  border: 1px solid var(--lm-border);
  border-radius: 12px;
}
.scrub-empty-icon { font-size: 3rem; color: var(--lm-text-subtle); }
.scrub-empty-title { font-size: 1.1rem; font-weight: 600; color: var(--lm-text); margin: 0; }
.scrub-empty-desc { color: var(--lm-text-muted); font-size: 0.9rem; max-width: 380px; margin: 0; }

.scrub-error {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

// Chart header with inline summary
.scrub-chart-header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 1rem;
  flex-wrap: wrap;
  margin-bottom: 1.25rem;
}

.scrub-summary {
  display: flex;
  align-items: baseline;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.scrub-summary-item {
  display: flex;
  align-items: baseline;
  gap: 0.3rem;

  .scrub-n {
    font-size: 1.35rem;
    font-weight: 700;
    color: var(--lm-text);
    line-height: 1;
  }
  .scrub-l {
    font-size: 0.8rem;
    color: var(--lm-text-muted);
  }
}

.scrub-summary-sep { color: var(--lm-border); font-size: 0.9rem; }

// Bar chart
.scrub-chart {
  display: flex;
  align-items: flex-end;
  gap: 3px;
  height: 120px;
  padding-bottom: 0;
}
.scrub-chart-col {
  flex: 1;
  height: 100%;
  display: flex;
  align-items: flex-end;
  cursor: default;
}
.scrub-chart-bar {
  width: 100%;
  background: var(--lm-primary-border);
  border-radius: 2px 2px 0 0;
  min-height: 2px;
  transition: background 0.15s;

  &--today { background: var(--lm-primary); }

  .scrub-chart-col:hover & { background: var(--lm-primary); }
}

.scrub-chart-labels {
  display: flex;
  justify-content: space-between;
  margin-top: 0.5rem;
  font-size: 0.72rem;
  color: var(--lm-text-subtle);
}

// Inline bar in table
.scrub-inline-bar {
  height: 6px;
  background: var(--lm-border);
  border-radius: 3px;
  overflow: hidden;
  width: 100%;
  max-width: 200px;
}
.scrub-inline-bar-fill {
  height: 100%;
  background: var(--lm-primary);
  border-radius: 3px;
}
</style>
