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
          <h4 class="maint-title">{{ $t('globals.terms.subscriptions', 2) }}</h4>
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

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import dayjs from 'dayjs';
import { storeToRefs } from 'pinia';
import { useI18n } from 'vue-i18n';
import { useMainStore } from '../store';
import { useGlobal } from '../composables/useGlobal';
import { getMaintenance } from '../api/generated/endpoints/maintenance/maintenance';
import { getSettings as settingsApi } from '../api/generated/endpoints/settings/settings';

const { $utils } = useGlobal();
const { gcSubscribers, gcSubscriptions, gcCampaignAnalytics } = getMaintenance();
const { getSettings, getServerConfig, updateSettingByKey } = settingsApi();
const { t, tc } = useI18n();
const { loading } = storeToRefs(useMainStore());

const isLoading = ref(false);
const subscriberType = ref('orphan');
const analyticsType = ref('all');
const subscriptionType = ref('optin');
const analyticsDate = ref(dayjs().subtract(7, 'day').toDate());
const subscriptionDate = ref(dayjs().subtract(7, 'day').toDate());
const exportType = ref('views');
const exportDate = ref(dayjs().subtract(30, 'day').toDate());
const dbSettings = ref({ vacuum: false, vacuum_cron_interval: '0 2 * * *' });

const exportURL = computed(() => {
  const since = encodeURIComponent(dayjs(exportDate.value).toISOString());
  return `/api/maintenance/analytics/${exportType.value}/export?since=${since}`;
});

const subscriberTypeOptions = computed(() => [
  { label: t('dashboard.orphanSubs'), value: 'orphan' },
  { label: t('subscribers.status.blocklisted'), value: 'blocklisted' },
]);

const subscriptionTypeOptions = computed(() => [
  { label: t('maintenance.maintenance.unconfirmedOptins'), value: 'optin' },
]);

const analyticsTypeOptions = computed(() => [
  { label: t('globals.terms.all'), value: 'all' },
  { label: t('dashboard.campaignViews'), value: 'views' },
  { label: t('dashboard.linkClicks'), value: 'clicks' },
]);

const exportTypeOptions = computed(() => [
  { label: t('dashboard.campaignViews'), value: 'views' },
  { label: t('dashboard.linkClicks'), value: 'clicks' },
]);

// eslint-disable-next-line @typescript-eslint/no-unused-vars
function formatDateTime(s: any) { return dayjs(s).format('YYYY-MM-DD'); }

function deleteSubscribers() {
  $utils.confirm(null, () => {
    gcSubscribers(subscriberType.value).then((data: any) => {
      $utils.toast(t('globals.messages.deletedCount', { name: tc('globals.terms.subscribers', 2), num: data.count }));
    });
  });
}

function deleteSubscriptions() {
  $utils.confirm(null, () => {
    const beforeDate = dayjs(subscriptionDate.value).toISOString();
    gcSubscriptions({ before_date: beforeDate }).then((data: any) => {
      $utils.toast(t('globals.messages.deletedCount', { name: tc('globals.terms.subscriptions', 2), num: data.count }));
    });
  });
}

function deleteAnalytics() {
  $utils.confirm(null, () => {
    const beforeDate = dayjs(analyticsDate.value).toISOString();
    gcCampaignAnalytics(analyticsType.value, { before_date: beforeDate }).then(() => {
      $utils.toast(t('globals.messages.done'));
    });
  });
}

function loadDBSettings() {
  getSettings().then((data: any) => {
    if (data['maintenance.db'] !== undefined) {
      dbSettings.value = { ...data['maintenance.db'] };
    }
  });
}

async function onUpdateDBSettings() {
  isLoading.value = true;
  try {
    await updateSettingByKey('maintenance.db', dbSettings.value);
    await getServerConfig();
  } finally {
    isLoading.value = false;
  }
}

onMounted(() => { loadDBSettings(); });
</script>

<style scoped lang="scss">
.maintenance-page { display: flex; flex-direction: column; gap: 1.5rem; }

.page-desc { font-size: 0.875rem; color: var(--lm-text-muted); margin: 0; }

.maintenance-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1.25rem;

  @media (max-width: 900px) { grid-template-columns: 1fr; }
}

.maint-card {
  background: var(--lm-surface);
  border: 1px solid var(--lm-border);
  border-radius: 12px;
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.maint-card-header { display: flex; align-items: center; gap: 0.6rem; border-bottom: 1px solid var(--lm-bg-subtle); padding-bottom: 0.75rem; }
.maint-icon { font-size: 1rem; color: var(--lm-primary); }
.maint-title { font-size: 1rem; font-weight: 700; color: var(--lm-text); margin: 0; }
.maint-subtitle { font-size: 0.9rem; font-weight: 600; color: var(--lm-text); margin: 0; }

.maint-field { display: flex; flex-direction: column; gap: 0.3rem; flex: 1; }
.maint-field--inline { flex-direction: row; align-items: center; justify-content: space-between; }
.maint-field--disabled { opacity: 0.45; pointer-events: none; }
.maint-field-row { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
.maint-label { font-size: 0.8rem; font-weight: 600; color: var(--lm-text); }
.maint-card--wide { grid-column: 1 / -1; }
.maint-help { font-size: 0.75rem; color: var(--lm-text-subtle); line-height: 1.4; margin: 0; }
.maint-footer { display: flex; justify-content: flex-end; }

.maint-db-section { display: flex; flex-direction: column; gap: 0.75rem; }
</style>
