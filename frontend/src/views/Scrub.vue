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

    <!-- Not configured state -->
    <div v-if="notConfigured" class="scrub-empty">
      <i class="pi pi-shield scrub-empty-icon" />
      <p class="scrub-empty-title">{{ $t('settings.scrub.name') }}</p>
      <p class="scrub-empty-desc">{{ $t('settings.scrub.notConfiguredDesc') }}</p>
      <PvButton severity="primary" icon="pi pi-cog" :label="$t('settings.scrub.goToSettings')"
        @click="$router.push({ name: 'settings' })" />
    </div>

    <!-- Error state -->
    <div v-else-if="fetchError" class="scrub-error settings-card">
      <i class="pi pi-exclamation-triangle text-orange-500" />
      <span>{{ fetchError }}</span>
      <PvButton severity="secondary" outlined size="small" icon="pi pi-refresh"
        :label="$t('globals.buttons.retry')" @click="fetchStats" />
    </div>

    <template v-else-if="stats">
      <!-- Stat cards -->
      <div class="scrub-stats-grid">
        <div class="scrub-stat-card">
          <div class="scrub-stat-label">{{ $t('settings.scrub.total') }}</div>
          <div class="scrub-stat-value">{{ fmt(stats.total ?? stats.emails_checked ?? stats.checked) }}</div>
        </div>
        <div class="scrub-stat-card scrub-stat-card--valid">
          <div class="scrub-stat-label">{{ $t('settings.scrub.valid') }}</div>
          <div class="scrub-stat-value">{{ fmt(stats.valid) }}</div>
          <div v-if="total > 0" class="scrub-stat-pct">{{ pct(stats.valid) }}%</div>
        </div>
        <div class="scrub-stat-card scrub-stat-card--invalid">
          <div class="scrub-stat-label">{{ $t('settings.scrub.invalid') }}</div>
          <div class="scrub-stat-value">{{ fmt(stats.invalid) }}</div>
          <div v-if="total > 0" class="scrub-stat-pct">{{ pct(stats.invalid) }}%</div>
        </div>
        <div class="scrub-stat-card scrub-stat-card--risky">
          <div class="scrub-stat-label">{{ $t('settings.scrub.risky') }}</div>
          <div class="scrub-stat-value">{{ fmt(stats.risky ?? stats.catch_all) }}</div>
          <div v-if="total > 0" class="scrub-stat-pct">{{ pct(stats.risky ?? stats.catch_all) }}%</div>
        </div>
        <div v-if="stats.unknown != null" class="scrub-stat-card scrub-stat-card--unknown">
          <div class="scrub-stat-label">{{ $t('settings.scrub.unknown') }}</div>
          <div class="scrub-stat-value">{{ fmt(stats.unknown) }}</div>
          <div v-if="total > 0" class="scrub-stat-pct">{{ pct(stats.unknown) }}%</div>
        </div>
      </div>

      <!-- Breakdown bar -->
      <div v-if="total > 0" class="settings-card">
        <p class="settings-section-label mb-3">Breakdown</p>
        <div class="scrub-bar">
          <div class="scrub-bar-segment scrub-bar--valid" :style="{ width: pct(stats.valid) + '%' }"
            v-tooltip="$t('settings.scrub.valid') + ': ' + pct(stats.valid) + '%'" />
          <div class="scrub-bar-segment scrub-bar--risky" :style="{ width: pct(stats.risky ?? stats.catch_all) + '%' }"
            v-tooltip="$t('settings.scrub.risky') + ': ' + pct(stats.risky ?? stats.catch_all) + '%'" />
          <div class="scrub-bar-segment scrub-bar--invalid" :style="{ width: pct(stats.invalid) + '%' }"
            v-tooltip="$t('settings.scrub.invalid') + ': ' + pct(stats.invalid) + '%'" />
          <div v-if="stats.unknown" class="scrub-bar-segment scrub-bar--unknown"
            :style="{ width: pct(stats.unknown) + '%' }"
            v-tooltip="$t('settings.scrub.unknown') + ': ' + pct(stats.unknown) + '%'" />
        </div>
        <div class="scrub-bar-legend">
          <span class="scrub-legend-item scrub-legend--valid">{{ $t('settings.scrub.valid') }}</span>
          <span class="scrub-legend-item scrub-legend--risky">{{ $t('settings.scrub.risky') }}</span>
          <span class="scrub-legend-item scrub-legend--invalid">{{ $t('settings.scrub.invalid') }}</span>
          <span v-if="stats.unknown" class="scrub-legend-item scrub-legend--unknown">{{ $t('settings.scrub.unknown') }}</span>
        </div>
      </div>

      <!-- Quota card -->
      <div v-if="stats.quota" class="settings-card">
        <p class="settings-section-label mb-3">{{ $t('settings.scrub.quota') }}</p>
        <div class="scrub-quota-row">
          <div class="scrub-quota-item">
            <span class="scrub-quota-label">{{ $t('settings.scrub.used') }}</span>
            <span class="scrub-quota-val">{{ fmt(stats.quota - (stats.remaining ?? 0)) }}</span>
          </div>
          <div class="scrub-quota-item">
            <span class="scrub-quota-label">{{ $t('settings.scrub.remaining') }}</span>
            <span class="scrub-quota-val">{{ fmt(stats.remaining) }}</span>
          </div>
          <div class="scrub-quota-item">
            <span class="scrub-quota-label">{{ $t('settings.scrub.quota') }}</span>
            <span class="scrub-quota-val">{{ fmt(stats.quota) }}</span>
          </div>
        </div>
        <PvProgressBar :value="quotaUsedPct" class="scrub-progress mt-3" />
        <small class="block mt-1 text-color-secondary">{{ quotaUsedPct }}% of daily quota used</small>
      </div>

      <!-- Date / raw fields -->
      <div class="settings-card scrub-meta">
        <span v-if="stats.date" class="text-sm text-color-secondary">
          <i class="pi pi-calendar mr-1" />{{ $t('settings.scrub.lastUpdated') }}: {{ stats.date }}
        </span>
        <!-- Any extra fields the API returns that aren't already displayed -->
        <template v-for="(val, key) in extraFields" :key="key">
          <span class="scrub-extra-field">
            <span class="scrub-extra-key">{{ key }}</span>
            <span class="scrub-extra-val">{{ val }}</span>
          </span>
        </template>
      </div>
    </template>
  </section>
</template>

<script>
import { mapState } from 'pinia';
import { useMainStore } from '../store';

const KNOWN_FIELDS = new Set(['date', 'total', 'emails_checked', 'checked', 'valid', 'invalid',
  'risky', 'catch_all', 'unknown', 'quota', 'remaining']);

export default {
  name: 'ScrubDashboard',

  data() {
    return {
      stats: null,
      notConfigured: false,
      fetchError: null,
    };
  },

  computed: {
    ...mapState(useMainStore, ['loading']),

    total() {
      if (!this.stats) return 0;
      return this.stats.total ?? this.stats.emails_checked ?? this.stats.checked ?? 0;
    },

    quotaUsedPct() {
      if (!this.stats?.quota || !this.stats.remaining) return 0;
      const used = this.stats.quota - this.stats.remaining;
      return Math.round((used / this.stats.quota) * 100);
    },

    extraFields() {
      if (!this.stats) return {};
      return Object.fromEntries(
        Object.entries(this.stats).filter(([k]) => !KNOWN_FIELDS.has(k)),
      );
    },
  },

  methods: {
    async fetchStats() {
      this.fetchError = null;
      this.notConfigured = false;
      try {
        const data = await this.$api.getScrubStats();
        this.stats = data;
      } catch (e) {
        const msg = e.response?.data?.message || '';
        if (e.response?.status === 400 && msg.includes('not enabled')) {
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

    pct(n) {
      if (!n || !this.total) return 0;
      return Math.round((n / this.total) * 100);
    },
  },

  mounted() {
    this.fetchStats();
  },
};
</script>

<style scoped lang="scss">
.scrub-dashboard {
  max-width: 900px;
}

// Empty / not-configured state
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
.scrub-empty-icon {
  font-size: 3rem;
  color: var(--lm-text-subtle);
}
.scrub-empty-title {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--lm-text);
  margin: 0;
}
.scrub-empty-desc {
  color: var(--lm-text-muted);
  font-size: 0.9rem;
  max-width: 380px;
  margin: 0;
}

// Error state
.scrub-error {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  color: var(--lm-text);
}

// Stat cards grid
.scrub-stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  gap: 1rem;
  margin-bottom: 1rem;
}

.scrub-stat-card {
  background: var(--lm-surface);
  border: 1px solid var(--lm-border);
  border-radius: 10px;
  padding: 1.25rem 1.5rem;

  .scrub-stat-label {
    font-size: 0.75rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--lm-text-muted);
    margin-bottom: 0.4rem;
  }
  .scrub-stat-value {
    font-size: 1.75rem;
    font-weight: 700;
    color: var(--lm-text);
    line-height: 1;
  }
  .scrub-stat-pct {
    font-size: 0.8rem;
    color: var(--lm-text-muted);
    margin-top: 0.25rem;
  }

  &--valid   { border-top: 3px solid #22c55e; }
  &--invalid { border-top: 3px solid #ef4444; }
  &--risky   { border-top: 3px solid #f59e0b; }
  &--unknown { border-top: 3px solid #94a3b8; }
}

// Breakdown bar
.scrub-bar {
  display: flex;
  height: 12px;
  border-radius: 6px;
  overflow: hidden;
  background: var(--lm-bg-subtle);
}
.scrub-bar-segment {
  height: 100%;
  transition: width 0.3s ease;
  min-width: 2px;
}
.scrub-bar--valid   { background: #22c55e; }
.scrub-bar--risky   { background: #f59e0b; }
.scrub-bar--invalid { background: #ef4444; }
.scrub-bar--unknown { background: #94a3b8; }

.scrub-bar-legend {
  display: flex;
  gap: 1rem;
  margin-top: 0.75rem;
  flex-wrap: wrap;
}
.scrub-legend-item {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.8rem;
  color: var(--lm-text-muted);

  &::before {
    content: '';
    display: inline-block;
    width: 10px;
    height: 10px;
    border-radius: 2px;
  }
  &.scrub-legend--valid::before   { background: #22c55e; }
  &.scrub-legend--risky::before   { background: #f59e0b; }
  &.scrub-legend--invalid::before { background: #ef4444; }
  &.scrub-legend--unknown::before { background: #94a3b8; }
}

// Quota
.scrub-quota-row {
  display: flex;
  gap: 2rem;
  flex-wrap: wrap;
}
.scrub-quota-item {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}
.scrub-quota-label {
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--lm-text-muted);
}
.scrub-quota-val {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--lm-text);
}

:deep(.scrub-progress.p-progressbar) {
  height: 8px;
  border-radius: 4px;
}

// Meta / extra fields
.scrub-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  align-items: center;
  font-size: 0.85rem;
}
.scrub-extra-field {
  display: flex;
  gap: 0.4rem;
  align-items: center;
}
.scrub-extra-key {
  font-weight: 600;
  color: var(--lm-text-muted);
  text-transform: capitalize;
}
.scrub-extra-val {
  color: var(--lm-text);
}
</style>
