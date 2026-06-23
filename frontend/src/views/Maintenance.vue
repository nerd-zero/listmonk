<template>
  <div class="maintenance-page">
    <div class="page-header">
      <h1 class="page-title">{{ $t('maintenance.title') }}</h1>
    </div>

    <p class="page-desc">{{ $t('maintenance.help') }}</p>

    <div class="maintenance-grid">
      <!-- Subscribers -->
      <div class="maint-card">
        <div class="maint-card-header">
          <i class="pi pi-users maint-icon" />
          <h4 class="maint-title">{{ $t('globals.terms.subscribers') }}</h4>
        </div>
        <div class="maint-field">
          <label class="maint-label">Data</label>
          <PvSelect v-model="subscriberType" :options="subscriberTypeOptions" option-label="label" option-value="value" class="w-full" />
          <small class="maint-help">{{ $t('maintenance.orphanHelp') }}</small>
        </div>
        <div class="maint-footer">
          <PvButton severity="danger" :loading="loading.maintenance" @click="deleteSubscribers"
            icon="pi pi-trash" :label="$t('globals.buttons.delete')" />
        </div>
      </div>

      <!-- Subscriptions -->
      <div class="maint-card">
        <div class="maint-card-header">
          <i class="pi pi-list maint-icon" />
          <h4 class="maint-title">{{ $tc('globals.terms.subscriptions', 2) }}</h4>
        </div>
        <div class="maint-field-row">
          <div class="maint-field">
            <label class="maint-label">Data</label>
            <PvSelect v-model="subscriptionType" :options="subscriptionTypeOptions" option-label="label" option-value="value" class="w-full" />
          </div>
          <div class="maint-field">
            <label class="maint-label">{{ $t('maintenance.olderThan') }}</label>
            <PvDatePicker v-model="subscriptionDate" required class="w-full" :date-format="'yy-mm-dd'" />
          </div>
        </div>
        <div class="maint-footer">
          <PvButton severity="danger" :loading="loading.maintenance" @click="deleteSubscriptions"
            icon="pi pi-trash" :label="$t('globals.buttons.delete')" />
        </div>
      </div>

      <!-- Analytics -->
      <div class="maint-card">
        <div class="maint-card-header">
          <i class="pi pi-chart-bar maint-icon" />
          <h4 class="maint-title">{{ $t('globals.terms.analytics') }}</h4>
        </div>
        <div class="maint-field-row">
          <div class="maint-field">
            <label class="maint-label">Data</label>
            <PvSelect v-model="analyticsType" :options="analyticsTypeOptions" option-label="label" option-value="value" class="w-full" />
          </div>
          <div class="maint-field">
            <label class="maint-label">{{ $t('maintenance.olderThan') }}</label>
            <PvDatePicker v-model="analyticsDate" required class="w-full" :date-format="'yy-mm-dd'" />
          </div>
        </div>
        <div class="maint-footer">
          <PvButton severity="danger" :loading="loading.maintenance" @click="deleteAnalytics"
            icon="pi pi-trash" :label="$t('globals.buttons.delete')" />
        </div>
      </div>

      <!-- Analytics Export -->
      <div class="maint-card">
        <div class="maint-card-header">
          <i class="pi pi-download maint-icon" />
          <h4 class="maint-title">{{ $t('subscribers.export') }}</h4>
        </div>
        <div class="maint-field-row">
          <div class="maint-field">
            <label class="maint-label">Data</label>
            <PvSelect v-model="exportType" :options="exportTypeOptions" option-label="label" option-value="value" class="w-full" />
          </div>
          <div class="maint-field">
            <label class="maint-label">{{ $t('analytics.fromDate') }}</label>
            <PvDatePicker v-model="exportDate" required class="w-full" :date-format="'yy-mm-dd'" />
          </div>
        </div>
        <div class="maint-footer">
          <a :href="exportURL">
            <PvButton severity="primary" icon="pi pi-download" :label="$t('subscribers.export')" />
          </a>
        </div>
      </div>

      <!-- Database -->
      <form class="maint-card maint-card--wide" @submit.prevent="onUpdateDBSettings">
        <div class="maint-card-header">
          <i class="pi pi-database maint-icon" />
          <h4 class="maint-title">{{ $t('maintenance.database.title') }}</h4>
        </div>
        <div class="maint-db-section">
          <h5 class="maint-subtitle">Vacuum</h5>
          <p class="maint-help">{{ $t('maintenance.database.vacuumHelp') }}</p>
          <div class="maint-field-row">
            <div class="maint-field maint-field--inline">
              <label class="maint-label">{{ $t('globals.buttons.enabled') }}</label>
              <PvToggleSwitch v-model="dbSettings.vacuum" />
            </div>
            <div class="maint-field" :class="{ 'maint-field--disabled': !dbSettings.vacuum }">
              <label class="maint-label">{{ $t('settings.maintenance.cron') }}</label>
              <PvInputText v-model="dbSettings.vacuum_cron_interval" placeholder="0 2 * * *"
                :disabled="!dbSettings.vacuum"
                pattern="((\*|[0-9,\-\/]+)\s+){4}(\*|[0-9,\-\/]+)" class="w-full" />
            </div>
          </div>
        </div>
        <div class="maint-footer">
          <PvButton severity="primary" type="submit" :loading="loading.settings || isLoading"
            :label="$t('globals.buttons.save')" />
        </div>
      </form>
    </div>
  </div>
</template>

<script>
import dayjs from 'dayjs';
import { mapState } from 'pinia';
import { useMainStore } from '../store';

export default {
  components: {
  },

  data() {
    return {
      isLoading: false,
      subscriberType: 'orphan',
      analyticsType: 'all',
      subscriptionType: 'optin',
      analyticsDate: dayjs().subtract(7, 'day').toDate(),
      subscriptionDate: dayjs().subtract(7, 'day').toDate(),
      exportType: 'views',
      exportDate: dayjs().subtract(30, 'day').toDate(),
      dbSettings: {
        vacuum: false,
        vacuum_cron_interval: '0 2 * * *',
      },
    };
  },

  mounted() {
    this.loadDBSettings();
  },

  methods: {
    formatDateTime(s) {
      return dayjs(s).format('YYYY-MM-DD');
    },

    deleteSubscribers() {
      this.$utils.confirm(
        null,
        () => {
          this.$api.deleteGCSubscribers(this.subscriberType).then((data) => {
            this.$utils.toast(this.$t(
              'globals.messages.deletedCount',
              { name: this.$tc('globals.terms.subscribers', 2), num: data.count },
            ));
          });
        },
      );
    },

    deleteSubscriptions() {
      this.$utils.confirm(
        null,
        () => {
          this.$api.deleteGCSubscriptions(this.subscriptionDate).then((data) => {
            this.$utils.toast(this.$t(
              'globals.messages.deletedCount',
              { name: this.$tc('globals.terms.subscriptions', 2), num: data.count },
            ));
          });
        },
      );
    },

    deleteAnalytics() {
      this.$utils.confirm(
        null,
        () => {
          this.$api.deleteGCCampaignAnalytics(this.analyticsType, this.analyticsDate)
            .then(() => {
              this.$utils.toast(this.$t('globals.messages.done'));
            });
        },
      );
    },

    loadDBSettings() {
      this.$api.getSettings().then((data) => {
        if (data['maintenance.db'] !== undefined) {
          this.dbSettings = { ...data['maintenance.db'] };
        }
      });
    },

    async onUpdateDBSettings() {
      this.isLoading = true;
      const data = await this.$api.updateSettingsByKey('maintenance.db', this.dbSettings);
      await this.$root.awaitRestart(data);
      this.isLoading = false;
    },
  },

  computed: {
    ...mapState(useMainStore, ['loading']),

    exportURL() {
      const since = encodeURIComponent(dayjs(this.exportDate).toISOString());
      return `/api/maintenance/analytics/${this.exportType}/export?since=${since}`;
    },

    subscriberTypeOptions() {
      return [
        { label: this.$t('dashboard.orphanSubs'), value: 'orphan' },
        { label: this.$t('subscribers.status.blocklisted'), value: 'blocklisted' },
      ];
    },

    subscriptionTypeOptions() {
      return [
        { label: this.$t('maintenance.maintenance.unconfirmedOptins'), value: 'optin' },
      ];
    },

    analyticsTypeOptions() {
      return [
        { label: this.$t('globals.terms.all'), value: 'all' },
        { label: this.$t('dashboard.campaignViews'), value: 'views' },
        { label: this.$t('dashboard.linkClicks'), value: 'clicks' },
      ];
    },

    exportTypeOptions() {
      return [
        { label: this.$t('dashboard.campaignViews'), value: 'views' },
        { label: this.$t('dashboard.linkClicks'), value: 'clicks' },
      ];
    },
  },
};
</script>

<style scoped lang="scss">
.maintenance-page { display: flex; flex-direction: column; gap: 1.5rem; }
.page-header { display: flex; align-items: center; }
.page-title { font-size: 1.5rem; font-weight: 700; color: #0f172a; margin: 0; }
.page-desc { font-size: 0.875rem; color: #64748b; margin: 0; }

.maintenance-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1.25rem;

  @media (max-width: 900px) { grid-template-columns: 1fr; }
}

.maint-card {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;

  &--wide { grid-column: 1 / -1; }
}

.maint-card-header { display: flex; align-items: center; gap: 0.6rem; border-bottom: 1px solid #f1f5f9; padding-bottom: 0.75rem; }
.maint-icon { font-size: 1rem; color: #3b82f6; }
.maint-title { font-size: 1rem; font-weight: 700; color: #0f172a; margin: 0; }
.maint-subtitle { font-size: 0.9rem; font-weight: 600; color: #374151; margin: 0; }

.maint-field { display: flex; flex-direction: column; gap: 0.3rem; flex: 1; }
.maint-field--inline { flex-direction: row; align-items: center; justify-content: space-between; }
.maint-field--disabled { opacity: 0.45; pointer-events: none; }
.maint-field-row { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
.maint-label { font-size: 0.8rem; font-weight: 600; color: #374151; }
.maint-help { font-size: 0.75rem; color: #94a3b8; line-height: 1.4; margin: 0; }
.maint-footer { display: flex; justify-content: flex-end; }

.maint-db-section { display: flex; flex-direction: column; gap: 0.75rem; }
</style>
